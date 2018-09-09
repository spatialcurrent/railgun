// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
)

var GO_RAILGUN_COMPRESSION_ALGORITHMS = []string{"none", "bzip2", "gzip", "snappy"}
var GO_RAILGUN_DEFAULT_SALT = "4F56C8C88B38CD8CD96BF8A9724F4BFE"

var processViper = viper.New()

func processOutput(content string, outputUri string, outputCompression string, outputAppend bool, outputPassphrase string, outputSalt string, s3_client *s3.S3) error {
	if outputUri == "stdout" {
		if len(outputPassphrase) > 0 {
			return errors.New("encryption only works with file output")
		}
		fmt.Println(content)
	} else if outputUri == "stderr" {
		if len(outputPassphrase) > 0 {
			return errors.New("encryption only works with file output")
		}
		fmt.Fprintf(os.Stderr, content)
	} else {

		outputWriter, err := grw.WriteToResource(outputUri, outputCompression, outputAppend, s3_client)
		if err != nil {
			return errors.Wrap(err, "error opening output file")
		}

		if len(outputPassphrase) > 0 {

			outputBlock, err := railgun.CreateCipher(outputSalt, outputPassphrase)
			if err != nil {
				return errors.New("error creating cipher for output passphrase")
			}
			outputPlainText := []byte(content + "\n")
			outputCipherText := make([]byte, aes.BlockSize+len(outputPlainText))
			iv := outputCipherText[:aes.BlockSize]
			if _, err := io.ReadFull(rand.Reader, iv); err != nil {
				return errors.Wrap(err, "error generating iv")
			}

			outputStream := cipher.NewCFBEncrypter(outputBlock, iv)
			outputStream.XORKeyStream(outputCipherText[aes.BlockSize:], outputPlainText)

			_, err = outputWriter.Write(outputCipherText)
			if err != nil {
				return errors.Wrap(err, "error writing encrypted data to output file")
			}

		} else {
			_, err = outputWriter.WriteString(content + "\n")
			if err != nil {
				return errors.Wrap(err, "error writing string to output file")
			}
		}

		err = outputWriter.Close()
		if err != nil {
			return errors.Wrap(err, "error closing output file.")
		}

	}
	return nil
}

func processObject(object interface{}, node dfl.Node, vars map[string]interface{}, funcs dfl.FunctionMap) (interface{}, error) {
	if node != nil {
		_, o, err := node.Evaluate(
			vars,
			object,
			funcs,
			[]string{"'", "\"", "`"})
		if err != nil {
			return "", errors.Wrap(err, "error evaluating filter")
		}
		return gss.StringifyMapKeys(o), nil
	}
	return object, nil
}

func handleErrors(errorsChannel chan error) {
	go func() {
		for err := range errorsChannel {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	}()
}

func handleMessages(messages chan string, wg *sync.WaitGroup) {
	go func() {
		for message := range messages {
			fmt.Println(message)
		}
		wg.Done()
	}()
}

func buildOptions(inputLine []byte, inputFormat string, inputHeader []string, inputComment string, inputLazyQuotes bool) (gss.Options, error) {
	options := gss.Options{
		Header:     inputHeader,
		Comment:    inputComment,
		LazyQuotes: inputLazyQuotes,
		Limit:      1,
	}

	if inputFormat == "csv" || inputFormat == "tsv" {
		options.Format = inputFormat
		options.Type = reflect.TypeOf(map[string]interface{}{})
		return options, nil
	}

	if inputFormat == "jsonl" {
		options.Format = "json"
		if str := strings.TrimLeftFunc(string(inputLine), unicode.IsSpace); len(str) > 0 && str[0] == '[' {
			options.Type = reflect.TypeOf([]interface{}{})
		}
		options.Type = reflect.TypeOf(map[string]interface{}{})
		return options, nil
	}

	return options, errors.New("invalid input format for handleInput " + inputFormat)
}

func handleInput(inputLines chan []byte, inputFormat string, inputHeader []string, inputComment string, inputLazyQuotes bool, node dfl.Node, vars map[string]interface{}, funcs dfl.FunctionMap, outputObjects chan interface{}, outputFormat string, errorsChannel chan error, verbose bool) error {

	go func() {

		for inputLine := range inputLines {
			options, err := buildOptions(
				inputLine,
				inputFormat,
				inputHeader,
				inputComment,
				inputLazyQuotes)
			if err != nil {
				errorsChannel <- errors.Wrap(err, "invalid options for input line "+string(inputLine))
				continue
			}
			inputObject, err := options.DeserializeBytes(inputLine, verbose)
			if err != nil {
				errorsChannel <- errors.Wrap(err, "error deserializing input using options "+fmt.Sprint(options))
				continue
			}
			outputObject, err := processObject(inputObject, node, vars, funcs)
			if err != nil {
				switch err.(type) {
				case *gss.ErrEmptyRow:
				default:
					errorsChannel <- errors.Wrap(err, "error processing object")
				}
			} else {
				switch outputObject.(type) {
				case dfl.Null:
				default:
					outputObjects <- outputObject
				}
			}
		}
		close(outputObjects)
	}()
	return nil
}

func handleOutput(outputUri string, format string, compression string, buffer bool, header []string, outputAppend bool, mkdirs bool, outputVars map[string]interface{}, funcs dfl.FunctionMap, objects chan interface{}, errorsChannel chan error, messages chan string, fileDescriptorLimit int, wg *sync.WaitGroup, s3_client *s3.S3, verbose bool) error {

	if outputUri == "stdout" {
		go func() {
			for object := range objects {
				line, err := formatObject(object, format, header)
				if err != nil {
					errorsChannel <- errors.Wrap(err, "error formatting object")
					break
				}
				messages <- line
			}
			close(messages)
			wg.Done()
		}()
		return nil
	}

	if outputUri == "stderr" {
		go func() {
			for object := range objects {
				line, err := formatObject(object, format, header)
				if err != nil {
					errorsChannel <- errors.Wrap(err, "error formatting object")
					break
				}
				//fmt.Fprintf(os.Stderr, line)
				messages <- line
			}
			close(messages)
			wg.Done()
		}()
		return nil
	}

	n, err := dfl.ParseCompile(outputUri)
	if err != nil {
		return errors.Wrap(err, "Error parsing dfl node.")
	}
	outputNode := n

	outputLines := make(chan struct {
		Path string
		Line string
	}, 1000)

	outputPathBuffers := map[string]struct {
		Writer grw.ByteWriteCloser
		Buffer *bytes.Buffer
	}{}
	outputPathSemaphores := map[string]chan struct{}{}
	outputFileDescriptorSemaphore := make(chan struct{}, fileDescriptorLimit)

	if verbose {
		fmt.Println("* created semaphores using file descriptor limit " + fmt.Sprint(fileDescriptorLimit))
	}

	var outputPathMutex = &sync.Mutex{}
	var outputBufferMutex = &sync.Mutex{}
	getOutputPathSemaphore := func(outputPathString string) chan struct{} {
		outputPathMutex.Lock()
		if _, ok := outputPathSemaphores[outputPathString]; !ok {
			outputPathSemaphores[outputPathString] = make(chan struct{}, 1)
		}
		outputPathMutex.Unlock()
		return outputPathSemaphores[outputPathString]
	}

	writeToOutputMemoryBuffer := func(outputPathString string, str string) {
		outputBufferMutex.Lock()
		if _, ok := outputPathBuffers[outputPathString]; !ok {
			outputWriter, outputBuffer, err := grw.WriteBytes(compression)
			if err != nil {
				panic(err)
			}
			outputPathBuffers[outputPathString] = struct {
				Writer grw.ByteWriteCloser
				Buffer *bytes.Buffer
			}{Writer: outputWriter, Buffer: outputBuffer}
		}
		_, err := outputPathBuffers[outputPathString].Writer.WriteString(str)
		if err != nil {
			panic(err)
		}
		outputBufferMutex.Unlock()
	}

	go func() {
		var wgLines sync.WaitGroup
		for line := range outputLines {
			wgLines.Add(1)
			go func(w *sync.WaitGroup, line struct {
				Path string
				Line string
			}) {
				defer w.Done()

				outputPathSemaphore := getOutputPathSemaphore(line.Path)

				outputFileDescriptorSemaphore <- struct{}{}
				outputPathSemaphore <- struct{}{}

				if buffer {
					writeToOutputMemoryBuffer(line.Path, line.Line+"\n")
				} else {
					if mkdirs {
						os.MkdirAll(filepath.Dir(line.Path), 0755)
					}

					outputWriter, err := grw.WriteToResource(line.Path, compression, true, s3_client)
					if err != nil {
						<-outputPathSemaphore
						<-outputFileDescriptorSemaphore
						errorsChannel <- errors.Wrap(err, "error opening file at path "+line.Path)
						return
					}

					_, err = outputWriter.WriteString(line.Line + "\n")
					if err != nil {
						errorsChannel <- errors.Wrap(err, "Error writing string to output file")
					}

					err = outputWriter.Close()
					if err != nil {
						errorsChannel <- errors.Wrap(err, "Error closing output file.")
					}
				}

				<-outputPathSemaphore
				<-outputFileDescriptorSemaphore

			}(&wgLines, line)
		}
		messages <- "* waiting for wgLines to be done"
		wgLines.Wait()
		messages <- "* closing output path semaphores"
		for _, outputPathSemaphore := range outputPathSemaphores {
			close(outputPathSemaphore)
		}
		messages <- "* writing buffers to files"
		for outputPath, outputBuffer := range outputPathBuffers {
			err := outputBuffer.Writer.Close()
			if err != nil {
				messages <- "* error closing output buffer for " + outputPath
			}
			if mkdirs {
				os.MkdirAll(filepath.Dir(outputPath), 0755)
			}
			outputWriter, err := grw.WriteToResource(outputPath, "", outputAppend, s3_client)
			if err != nil {
				messages <- "* error opening output file at " + outputPath
			}
			_, err = outputWriter.Write(outputBuffer.Buffer.Bytes())
			if err != nil {
				messages <- "* error writing buffer to output file at " + outputPath
			}
			err = outputWriter.Close()
			if err != nil {
				messages <- "* error closing output file at " + outputPath
			}
		}
		messages <- "* closing file descriptor semaphore"
		close(outputFileDescriptorSemaphore)
		messages <- "* closing messages"
		close(messages)
		wg.Done()
	}()

	if verbose {
		fmt.Println("* starting to process objects")
	}

	go func() {
		for object := range objects {
			outputPath, err := processObject(object, outputNode, outputVars, funcs)
			if err != nil {
				errorsChannel <- errors.Wrap(err, "Error writing string to output file")
				break
			}

			if reflect.TypeOf(outputPath).Kind() != reflect.String {
				errorsChannel <- errors.Wrap(err, "output path is not a string")
				break
			}

			outputPathString, err := homedir.Expand(outputPath.(string))
			if err != nil {
				errorsChannel <- errors.Wrap(err, "output path cannot be expanded")
				break
			}

			line, err := formatObject(object, format, header)
			if err != nil {
				errorsChannel <- errors.Wrap(err, "error formatting object")
				break
			}

			getOutputPathSemaphore(outputPathString)

			outputLines <- struct {
				Path string
				Line string
			}{Path: outputPathString, Line: line}

		}
		messages <- "* closing output lines"
		close(outputLines)
	}()

	return nil
}

func formatObject(object interface{}, format string, header []string) (string, error) {
	if format == "jsonl" {
		str, err := gss.SerializeString(object, "json", header, -1)
		if err != nil {
			return "", errors.Wrap(err, "error serializing object")
		}
		return str, nil
	}
	str, err := gss.SerializeString(object, format, header, -1)
	if err != nil {
		return "", errors.Wrap(err, "error serializing object")
	}
	return str, nil
}

func processFunction(cmd *cobra.Command, args []string) {

	v := processViper

	v.BindPFlags(cmd.PersistentFlags())
	v.BindPFlags(cmd.Flags())
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	railgun.MergeConfigs(v, v.GetStringArray("config-uri"))

	verbose := v.GetBool("verbose")

	if verbose {
		fmt.Println("=================================================")
		fmt.Println("Configuration:")
		fmt.Println("-------------------------------------------------")
		str, err := gss.SerializeString(v.AllSettings(), "properties", []string{}, -1)
		if err != nil {
			fmt.Println("error getting all settings")
			os.Exit(1)
		}
		fmt.Println(str)
		fmt.Println("=================================================")
	}

	fileDescriptorLimit := v.GetInt("file-descriptor-limit")

	stream := v.GetBool("stream")

	// AWS Flags
	awsDefaultRegion := v.GetString("aws-default-region")
	awsAccessKeyId := v.GetString("aws-access-key-id")
	awsSecretAccessKey := v.GetString("aws-secret-access-key")
	awsSessionToken := v.GetString("aws-session-token")

	// Input Flags
	inputUri := v.GetString("input-uri")
	inputFormat := v.GetString("input-format")
	inputHeader := v.GetStringSlice("input-header")
	inputComment := v.GetString("input-comment")
	inputLazyQuotes := v.GetBool("input-lazy-quotes")
	inputCompression := v.GetString("input-compression")
	inputReaderBufferSize := v.GetInt("input-reader-buffer-size")
	inputPassphrase := v.GetString("input-passphrase")
	inputSalt := v.GetString("input-salt")
	inputLimit := v.GetInt("input-limit")

	// Output Flags
	outputUri := v.GetString("output-uri")
	outputFormat := v.GetString("output-format")
	outputHeader := v.GetStringSlice("output-header")
	outputCompression := v.GetString("output-compression")
	outputBufferMemory := v.GetBool("output-buffer-memory")
	outputAppend := v.GetBool("output-append")
	outputPassphrase := v.GetString("output-passphrase")
	outputSalt := v.GetString("output-salt")
	//outputLimit := v.GetInt("output-limit")
	outputLimit := v.GetInt("output-limit")
	outputMkdirs := v.GetBool("output-mkdirs")

	// DFL Flgas
	dflExpression := v.GetString("dfl-expression")
	dflUri := v.GetString("dfl-uri")
	dflVarsString := v.GetString("dfl-vars")

	// Error Flags
	errorDestination := v.GetString("error-destination")
	errorCompression := v.GetString("error-compression")

	// Logging Flags
	logDestination := v.GetString("log-destination")
	logCompression := v.GetString("log-compression")
	//logFormat := v.GetString("log-format")

	var aws_session *session.Session
	var s3_client *s3.S3

	if strings.HasPrefix(inputUri, "s3://") || strings.HasPrefix(outputUri, "s3://") || strings.HasPrefix(errorDestination, "s3://") || strings.HasPrefix(logDestination, "s3://") {
		aws_session = railgun.ConnectToAWS(awsAccessKeyId, awsSecretAccessKey, awsSessionToken, awsDefaultRegion)
		s3_client = s3.New(aws_session)
	}

	errorWriter, err := grw.WriteToResource(errorDestination, errorCompression, true, s3_client)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error creating error writer"))
		os.Exit(1)
	}

	logWriter, err := grw.WriteToResource(logDestination, logCompression, true, s3_client)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error creating log writer").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	errorsChannel := make(chan error)
	messages := make(chan interface{}, 1000)

	go func(messages chan interface{}) {
		for message := range messages {
			logWriter.WriteString(fmt.Sprint(message) + "\n")
			logWriter.Flush()
		}
	}(messages)

	if errorDestination == logDestination {
		go func(errorsChannel chan error) {
			for err := range errorsChannel {
				messages <- err.Error()
			}
		}(errorsChannel)
	} else {
		go func(errorsChannel chan error) {
			for err := range errorsChannel {
				switch rerr := err.(type) {
				case *railgunerrors.ErrInvalidParameter:
					errorWriter.WriteString(rerr.Error())
				case *railgunerrors.ErrMissing:
					errorWriter.WriteString(rerr.Error())
				default:
					errorWriter.WriteString(rerr.Error())
				}
			}
		}(errorsChannel)
	}

	if len(inputFormat) == 0 || len(inputCompression) == 0 {
		_, inputPath := grw.SplitUri(inputUri)
		_, inputFormatGuess, inputCompressionGuess := railgun.SplitNameFormatCompression(inputPath)
		if len(inputFormat) == 0 {
			inputFormat = inputFormatGuess
		}
		if len(inputCompression) == 0 {
			inputCompression = inputCompressionGuess
		}
	}

	inputReader, inputMetadata, err := grw.ReadFromResource(inputUri, inputCompression, inputReaderBufferSize, false, s3_client)
	if err != nil {
		errorWriter.WriteError(errors.Wrap(err, "error opening resource from uri "+inputUri))
		errorWriter.Close()
		os.Exit(1)
	}

	if len(outputFormat) == 0 || len(outputCompression) == 0 {
		_, outputPath := grw.SplitUri(outputUri)
		_, outputFormatGuess, outputCompressionGuess := railgun.SplitNameFormatCompression(outputPath)
		if len(outputFormat) == 0 {
			outputFormat = outputFormatGuess
		}
		if len(outputCompression) == 0 {
			outputCompression = outputCompressionGuess
		}
	}

	if len(inputFormat) == 0 {
		if inputMetadata != nil {
			if len(inputMetadata.ContentType) > 0 {
				switch inputMetadata.ContentType {
				case "application/json":
					inputFormat = "json"
				case "application/vnd.geo+json":
					inputFormat = "json"
				case "application/toml":
					inputFormat = "toml"
				}
			}
		}
		if len(inputFormat) == 0 && len(outputFormat) > 0 {
			errorWriter.WriteString("Error: Provided no --input-format and could not infer from resource.")
			errorWriter.Close()
			os.Exit(1)
		}
	}

	if len(inputFormat) > 0 && len(outputFormat) == 0 {
		errorWriter.WriteString("Error: Provided input format but no output format.")
		errorWriter.WriteString("Run \"railgun --help\" for more information.")
		errorWriter.Close()
		os.Exit(1)
	}

	if stream {

		if !(outputFormat == "csv" || outputFormat == "tsv" || outputFormat == "jsonl") {
			errorWriter.WriteString("output format " + outputFormat + " is not compatible with streaming")
			errorWriter.Close()
			os.Exit(1)
		}

		if len(outputPassphrase) > 0 {
			errorWriter.WriteString("output passphrase is not compatible with streaming because it uses a block cipher")
			errorWriter.Close()
			os.Exit(1)
		}

		// Stream Processing with Batch Input
		if len(inputPassphrase) > 0 || !(outputFormat == "csv" || outputFormat == "tsv" || outputFormat == "jsonl") {

			inputBytesEncrypted, err := inputReader.ReadAll()
			if err != nil {
				errorWriter.WriteString("error reading from resource")
				errorWriter.Close()
				os.Exit(1)
			}

			err = inputReader.Close()
			if err != nil {
				errorWriter.WriteString(errors.Wrap(err, "error closing input").Error())
				errorWriter.Close()
				os.Exit(1)
			}

			inputBytesPlain, err := railgun.DecryptInput(inputBytesEncrypted, inputPassphrase, inputSalt)
			if err != nil {
				errorWriter.WriteString(errors.Wrap(err, "error decoding input").Error())
				errorWriter.Close()
				os.Exit(1)
			}

			inputType, err := gss.GetType(inputBytesPlain, inputFormat)
			if err != nil {
				errorWriter.WriteString(errors.Wrap(err, "error getting type for input").Error())
				errorWriter.Close()
				os.Exit(1)
			}

			if !(inputType.Kind() == reflect.Array || inputType.Kind() == reflect.Slice) {
				errorWriter.WriteString("input type cannot be streamed as it is not an array or slice but " + fmt.Sprint(inputType))
				errorWriter.Close()
				os.Exit(1)
			}

			inputObject, err := gss.DeserializeBytes(inputBytesPlain, inputFormat, inputHeader, inputComment, inputLazyQuotes, inputLimit, inputType, verbose)
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error deserializing input using format "+inputFormat))
				errorWriter.Close()
				os.Exit(1)
			}

			dflNode, err := railgun.ParseDfl(dflUri, dflExpression)
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error parsing"))
				errorWriter.Close()
				os.Exit(1)
			}
			funcs := dfl.NewFuntionMapWithDefaults()

			dflVars := map[string]interface{}{}
			if len(dflVarsString) > 0 {
				_, dflVarsMap, err := dfl.ParseCompileEvaluateMap(
					dflVarsString,
					map[string]interface{}{},
					map[string]interface{}{},
					funcs,
					dfl.DefaultQuotes)
				if err != nil {
					errorWriter.WriteError(errors.Wrap(err, "error parsing initial dfl vars as map"))
					errorWriter.Close()
					os.Exit(1)
				}
				if m, ok := gss.StringifyMapKeys(dflVarsMap).(map[string]interface{}); ok {
					dflVars = m
				}
			}

			var wgObjects sync.WaitGroup
			var wgMessages sync.WaitGroup
			errorsChannel := make(chan error, 1000)
			outputObjects := make(chan interface{}, 1000)
			messages := make(chan string, 1000)

			wgObjects.Add(1)
			wgMessages.Add(1)
			handleErrors(errorsChannel)
			handleMessages(messages, &wgMessages)
			handleOutput(
				outputUri,
				outputFormat,
				outputCompression,
				outputBufferMemory,
				outputHeader,
				outputAppend,
				outputMkdirs,
				dflVars,
				funcs,
				outputObjects,
				errorsChannel,
				messages,
				fileDescriptorLimit,
				&wgObjects,
				s3_client,
				verbose)

			inputObjectValue := reflect.ValueOf(inputObject)
			inputObjectLength := inputObjectValue.Len()
			for i := 0; i < inputObjectLength; i++ {
				output, err := processObject(inputObjectValue.Index(i).Interface(), dflNode, dflVars, funcs)
				if err != nil {
					errorWriter.WriteString(errors.Wrap(err, "error processing object").Error())
					errorWriter.Close()
					os.Exit(1)
				}
				switch output.(type) {
				case dfl.Null:
				default:
					outputObjects <- output
				}
			}
			close(outputObjects)
			wgObjects.Wait()
			wgMessages.Wait()
			logWriter.Close()
			return
		}

		dflNode, err := railgun.ParseDfl(dflUri, dflExpression)
		if err != nil {
			errorWriter.WriteString(errors.Wrap(err, "error parsing").Error())
			errorWriter.Close()
			os.Exit(1)
		}
		funcs := dfl.NewFuntionMapWithDefaults()

		dflVars := map[string]interface{}{}
		if len(dflVarsString) > 0 {
			_, dflVarsMap, err := dfl.ParseCompileEvaluateMap(
				dflVarsString,
				map[string]interface{}{},
				map[string]interface{}{},
				funcs,
				dfl.DefaultQuotes)
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error parsing initial dfl vars as map"))
				errorWriter.Close()
				os.Exit(1)
			}
			if m, ok := gss.StringifyMapKeys(dflVarsMap).(map[string]interface{}); ok {
				dflVars = m
			}
		}

		if len(inputHeader) == 0 && (inputFormat == "csv" || inputFormat == "tsv") {
			inputBytes, err := inputReader.ReadBytes('\n')
			if err != nil {
				errorWriter.WriteString("error reading header from resource")
				errorWriter.Close()
				os.Exit(1)
			}
			csvReader := csv.NewReader(bytes.NewReader(inputBytes))
			if inputFormat == "tsv" {
				csvReader.Comma = '\t'
			}
			csvReader.LazyQuotes = inputLazyQuotes
			if len(inputComment) > 1 {
				errorWriter.WriteString("go's encoding/csv package only supports single character comment characters")
				errorWriter.Close()
				os.Exit(1)
			} else if len(inputComment) == 1 {
				csvReader.Comment = []rune(inputComment)[0]
			}
			h, err := csvReader.Read()
			if err != nil {
				if err != io.EOF {
					errorWriter.WriteString(errors.Wrap(err, "Error reading header from input with format csv").Error())
					errorWriter.Close()
					os.Exit(1)
				}
			}
			inputHeader = h
		}

		var wgObjects sync.WaitGroup
		var wgMessages sync.WaitGroup
		errorsChannel := make(chan error, 1000)
		messages := make(chan string, 1000)
		inputLines := make(chan []byte, 1000)
		outputObjects := make(chan interface{}, 1000)

		wgObjects.Add(1)
		wgMessages.Add(1)
		handleErrors(errorsChannel)
		handleMessages(messages, &wgMessages)
		handleInput(
			inputLines,
			inputFormat,
			inputHeader,
			inputComment,
			inputLazyQuotes,
			dflNode,
			dflVars,
			funcs,
			outputObjects,
			outputFormat,
			errorsChannel,
			verbose)
		handleOutput(
			outputUri,
			outputFormat,
			outputCompression,
			outputBufferMemory,
			outputHeader,
			outputAppend,
			outputMkdirs,
			dflVars,
			funcs,
			outputObjects,
			errorsChannel,
			messages,
			fileDescriptorLimit,
			&wgObjects,
			s3_client,
			verbose)

		inputCount := 0
		for {
			inputBytes, err := inputReader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					errorWriter.WriteString("error reading line from resource")
					errorWriter.Close()
					os.Exit(1)
				}
			}
			inputLines <- inputBytes
			inputCount += 1
			if inputLimit > 0 && inputCount >= inputLimit {
				break
			}
		}
		err = inputReader.Close()
		if err != nil {
			errorsChannel <- errors.Wrap(err, "error closing input")
		}
		close(inputLines)
		wgObjects.Wait()
		wgMessages.Wait()
		logWriter.Close()
		return

	}

	// Batch Processing

	funcs := dfl.NewFuntionMapWithDefaults()

	dflVars := map[string]interface{}{}
	if len(dflVarsString) > 0 {
		_, dflVarsMap, err := dfl.ParseCompileEvaluateMap(
			dflVarsString,
			map[string]interface{}{},
			map[string]interface{}{},
			funcs,
			dfl.DefaultQuotes)
		if err != nil {
			errorWriter.WriteError(errors.Wrap(err, "error parsing initial dfl vars as map"))
			errorWriter.Close()
			os.Exit(1)
		}
		if m, ok := gss.StringifyMapKeys(dflVarsMap).(map[string]interface{}); ok {
			dflVars = m
		}
	}

	inputBytesEncrypted, err := inputReader.ReadAll()
	if err != nil {
		errorWriter.WriteString("error reading from resource")
		errorWriter.Close()
		os.Exit(1)
	}

	err = inputReader.Close()
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error closing input").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	inputBytesPlain, err := railgun.DecryptInput(inputBytesEncrypted, inputPassphrase, inputSalt)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error decoding input").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	outputString, err := railgun.ProcessInput(inputBytesPlain, inputFormat, inputHeader, inputComment, inputLazyQuotes, inputLimit, dflExpression, dflVars, dflUri, outputFormat, outputHeader, outputLimit, verbose)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error processing input").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	err = processOutput(outputString, outputUri, outputCompression, outputAppend, outputPassphrase, outputSalt, s3_client)
	if err != nil {
		errorWriter.WriteString(errors.Wrap(err, "error processing output").Error())
		errorWriter.Close()
		os.Exit(1)
	}

	logWriter.Close()

}

// processCmd represents the process command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "process input data using provided DFL expression",
	Run:   processFunction,
}

func init() {

	rootCmd.AddCommand(processCmd)

	processCmd.Flags().BoolP("dry-run", "", false, "parse and compile expression, but do not evaluate against context")
	processCmd.Flags().BoolP("pretty", "p", false, "print pretty output")
	processCmd.Flags().BoolP("stream", "s", false, "stream process (context == row rather than encompassing array)")

	// Input Flags
	processCmd.Flags().StringP("input-uri", "i", "stdin", "the input uri")
	processCmd.Flags().StringP("input-compression", "", "", "the input compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	processCmd.Flags().StringP("input-format", "", "", "the input format: "+strings.Join(gss.Formats, ", "))
	processCmd.Flags().StringSliceP("input-header", "", []string{}, "the input header, if the stdin input has no header.")
	processCmd.Flags().StringP("input-comment", "c", "", "the comment character for the input, e.g, #")
	processCmd.Flags().BoolP("input-lazy-quotes", "", false, "allows lazy quotes for CSV and TSV")
	processCmd.Flags().StringP("input-passphrase", "", "", "input passphrase for AES-256 encryption")
	processCmd.Flags().StringP("input-salt", "", GO_RAILGUN_DEFAULT_SALT, "input salt for AES-256 encryption")
	processCmd.Flags().IntP("input-reader-buffer-size", "", 4096, "the buffer size for the input reader")
	processCmd.Flags().IntP("input-limit", "", -1, "maximum number of objects to read from input")

	// Output Flags
	processCmd.Flags().StringP("output-uri", "o", "stdout", "the output uri (a dfl expression itself)")
	processCmd.Flags().StringP("output-compression", "", "", "the output compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	processCmd.Flags().StringP("output-format", "", "", "the output format: "+strings.Join(gss.Formats, ", "))
	processCmd.Flags().StringSliceP("output-header", "", []string{}, "the output header")
	processCmd.Flags().StringP("output-passphrase", "", "", "output passphrase for AES-256 encryption")
	processCmd.Flags().StringP("output-salt", "", "", "output salt for AES-256 encryption")
	processCmd.Flags().IntP("output-limit", "", -1, "maximum number of objects to send to output")
	processCmd.Flags().BoolP("output-append", "", false, "append to output files")
	processCmd.Flags().BoolP("output-buffer-memory", "", false, "buffer output in memory")
	processCmd.Flags().BoolP("output-mkdirs", "", false, "make directories if missing for output files")

	// DFL Flags
	processCmd.Flags().StringP("dfl-expression", "", "", "DFL expression to use")
	processCmd.Flags().StringP("dfl-uri", "", "", "URI to DFL file to use")
	processCmd.Flags().StringP("dfl-vars", "", "", "initial variables to use when evaluating DFL expression")

}
