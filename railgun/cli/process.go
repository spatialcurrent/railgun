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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	//"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spatialcurrent/cobra"
	"github.com/spatialcurrent/viper"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	//rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/athenaiterator"
	"github.com/spatialcurrent/railgun/railgun/config"
	rlogger "github.com/spatialcurrent/railgun/railgun/logger"
	"github.com/spatialcurrent/railgun/railgun/util"
)

var GO_RAILGUN_COMPRESSION_ALGORITHMS = []string{"none", "bzip2", "gzip", "snappy"}
var GO_RAILGUN_DEFAULT_SALT = "4F56C8C88B38CD8CD96BF8A9724F4BFE"

var processViper = viper.New()

//outputUri string, outputCompression string, outputAppend bool, outputPassphrase string, outputSalt string,
func processOutput(content string, output *config.Output, s3_client *s3.S3) error {
	if output.Uri == "stdout" {
		if output.IsEncrypted() {
			return errors.New("encryption only works with file output")
		}
		fmt.Println(content)
	} else if output.Uri == "stderr" {
		if output.IsEncrypted() {
			return errors.New("encryption only works with file output")
		}
		fmt.Fprintf(os.Stderr, content)
	} else {

		outputWriter, err := grw.WriteToResource(output.Uri, output.Compression, output.Append, s3_client)
		if err != nil {
			return errors.Wrap(err, "error opening output file")
		}

		if output.IsEncrypted() {

			outputBlock, err := util.CreateCipher(output.Salt, output.Passphrase)
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
			_, err = outputWriter.WriteLine(content)
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

func processObject(object interface{}, node dfl.Node, vars map[string]interface{}) (interface{}, error) {
	if node != nil {
		_, o, err := node.Evaluate(
			vars,
			object,
			dfl.DefaultFunctionMap,
			[]string{"'", "\"", "`"})
		if err != nil {
			return "", errors.Wrap(err, "error evaluating filter")
		}
		return gss.StringifyMapKeys(o), nil
	}
	return object, nil
}

func buildOptions(inputLine []byte, inputFormat string, inputHeader []string, inputComment string, inputLazyQuotes bool) (gss.Options, error) {
	options := gss.Options{
		Header:     inputHeader,
		Comment:    inputComment,
		LazyQuotes: inputLazyQuotes,
		SkipLines:  0,
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

func handleInput(inputLines chan []byte, input *config.Input, node dfl.Node, vars map[string]interface{}, outputObjects chan interface{}, outputFormat string, errorsChannel chan error, verbose bool) error {

	go func() {

		for inputLine := range inputLines {
			options, err := buildOptions(
				inputLine,
				input.Format,
				input.Header,
				input.Comment,
				input.LazyQuotes)
			if err != nil {
				errorsChannel <- errors.Wrap(err, "invalid options for input line "+string(inputLine))
				continue
			}
			inputObject, err := options.DeserializeBytes(inputLine, verbose)
			if err != nil {
				errorsChannel <- errors.Wrap(err, "error deserializing input using options "+fmt.Sprint(options))
				continue
			}
			outputObject, err := processObject(inputObject, node, vars)
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

func handleOutput(output *config.Output, outputVars map[string]interface{}, objects chan interface{}, errorsChannel chan error, messages chan interface{}, fileDescriptorLimit int, wg *sync.WaitGroup, s3_client *s3.S3, verbose bool) error {

	if output.Uri == "stdout" {
		go func() {
			for object := range objects {
				line, err := formatObject(object, output.Format, output.Header)
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

	if output.Uri == "stderr" {
		go func() {
			for object := range objects {
				line, err := formatObject(object, output.Format, output.Header)
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

	n, err := dfl.ParseCompile(output.Uri)
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

	writeToOutputMemoryBuffer := func(outputPathString string, line string) {
		outputBufferMutex.Lock()
		if _, ok := outputPathBuffers[outputPathString]; !ok {
			//outputWriter, outputBuffer, err := grw.WriteSnappyBytes(output.Compression)
			outputWriter, outputBuffer, err := grw.WriteBytes(output.Compression)
			if err != nil {
				panic(err)
			}
			outputPathBuffers[outputPathString] = struct {
				Writer grw.ByteWriteCloser
				Buffer *bytes.Buffer
			}{Writer: outputWriter, Buffer: outputBuffer}
		}
		_, err := outputPathBuffers[outputPathString].Writer.WriteLine(line)
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

				if output.BufferMemory {
					writeToOutputMemoryBuffer(line.Path, line.Line)
				} else {
					if output.Mkdirs {
						os.MkdirAll(filepath.Dir(line.Path), 0755)
					}

					outputWriter, err := grw.WriteToResource(line.Path, output.Compression, true, s3_client)
					if err != nil {
						<-outputPathSemaphore
						<-outputFileDescriptorSemaphore
						errorsChannel <- errors.Wrap(err, "error opening file at path "+line.Path)
						return
					}

					_, err = outputWriter.WriteLine(line.Line)
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
			if output.Mkdirs {
				os.MkdirAll(filepath.Dir(outputPath), 0755)
			}
			outputWriter, err := grw.WriteToResource(outputPath, "", output.Append, s3_client)
			if err != nil {
				messages <- "* error opening output file at " + outputPath
			}
			//_, err = ioutil.Copy(outputWriter, snappy.NewReader(bytes.NewReader(outputBuffer.Buffer.Bytes())))
			//_, err = outputWriter.Write(outputBuffer.Buffer.Bytes())
			_, err = io.Copy(outputWriter, outputBuffer.Buffer)
			if err != nil {
				messages <- "* error writing buffer to output file at " + outputPath
			}
			err = outputWriter.Close()
			if err != nil {
				messages <- "* error closing output file at " + outputPath
			}
			// delete output buffer and writer, since done writing to file
			delete(outputPathBuffers, outputPath)
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
			outputPath, err := processObject(object, outputNode, outputVars)
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

			line, err := formatObject(object, output.Format, output.Header)
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
		str, err := gss.SerializeString(object, "json", header, gss.NoLimit)
		if err != nil {
			return "", errors.Wrap(err, "error serializing object")
		}
		return str, nil
	}
	str, err := gss.SerializeString(object, format, header, gss.NoLimit)
	if err != nil {
		return "", errors.Wrap(err, "error serializing object")
	}
	return str, nil
}

func processAthenaInput(inputUri string, inputLimit int, tempUri string, outputFormat string, athenaClient *athena.Athena, logger *rlogger.Logger, verbose bool) (*athenaiterator.AthenaIterator, error) {

	if !strings.HasPrefix(tempUri, "s3://") {
		return nil, errors.New("temporary uri must be an S3 object")
	}

	_, queryName := grw.SplitUri(inputUri)
	getNamedQueryOutput, err := athenaClient.GetNamedQuery(&athena.GetNamedQueryInput{
		NamedQueryId: aws.String(queryName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting named athena query")
	}
	startQueryExecutionOutput, err := athenaClient.StartQueryExecution(&athena.StartQueryExecutionInput{
		QueryExecutionContext: &athena.QueryExecutionContext{
			Database: getNamedQueryOutput.NamedQuery.Database,
		},
		QueryString: getNamedQueryOutput.NamedQuery.QueryString,
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String(tempUri),
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "error starting athena query")
	}
	queryExecutionId := startQueryExecutionOutput.QueryExecutionId

	if verbose {
		logger.Info(fmt.Sprintf("* waiting for athena query %s to complete", *startQueryExecutionOutput.QueryExecutionId))
	}

	queryExecutionState := ""
	queryExecutionStateReason := ""
	for i := 0; i < 36; i++ {
		getQueryExecutionOutput, err := athenaClient.GetQueryExecution(&athena.GetQueryExecutionInput{
			QueryExecutionId: queryExecutionId,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error waiting on athena query")
		}
		queryExecutionState = *getQueryExecutionOutput.QueryExecution.Status.State
		if getQueryExecutionOutput.QueryExecution.Status.StateChangeReason != nil {
			queryExecutionStateReason = *getQueryExecutionOutput.QueryExecution.Status.StateChangeReason
		}
		wait := true
		switch queryExecutionState {
		case "QUEUED":
		case "RUNNING":
		case "FAILED", "SUCCEEDED", "CANCELLED":
			wait = false
		}
		if !wait {
			break
		}
		if verbose {
			logger.Info("* sleeping for 5 seconds")
			logger.Flush()
		}
		time.Sleep(5 * time.Second)
	}

	switch queryExecutionState {
	case "QUEUED":
	case "RUNNING":
	case "FAILED", "SUCCEEDED", "CANCELLED":
		logger.Info(fmt.Sprintf("* athena query result: %s: %s", queryExecutionState, queryExecutionStateReason))
	}

	if len(outputFormat) == 0 {
		return nil, nil // just exit if no output-format is given
	}

	athenaIterator, err := athenaiterator.New(athenaClient, queryExecutionId, inputLimit)
	if err != nil {
		return nil, errors.Wrap(err, "error creating athena iterator")
	}
	return athenaIterator, nil
}

func processFunction(cmd *cobra.Command, args []string) {

	v := processViper

	v.BindPFlags(cmd.PersistentFlags())
	v.BindPFlags(cmd.Flags())
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))

	verbose := v.GetBool("verbose")

	if verbose {
		fmt.Println("=================================================")
		fmt.Println("Viper:")
		fmt.Println("-------------------------------------------------")
		str, err := gss.SerializeString(v.AllSettings(), "properties", []string{}, gss.NoLimit)
		if err != nil {
			fmt.Println("error getting all settings")
			os.Exit(1)
		}
		fmt.Println(str)
		fmt.Println("=================================================")
	}

	fileDescriptorLimit := v.GetInt("file-descriptor-limit")
	stream := v.GetBool("stream")

	processConfig := &config.Process{
		AWS:              &config.AWS{},
		Input:            &config.Input{},
		Output:           &config.Output{},
		Temp:             &config.Temp{},
		Dfl:              &config.Dfl{},
		InfoDestination:  "",
		InfoCompression:  "",
		InfoFormat:       "",
		ErrorDestination: "",
		ErrorCompression: "",
		ErrorFormat:      "",
	}
	config.LoadConfigFromViper(processConfig, v)

	if verbose {
		fmt.Println("=================================================")
		fmt.Println("Configuration:")
		fmt.Println("-------------------------------------------------")
		str, err := gss.SerializeString(processConfig.Map(), "yaml", gss.NoHeader, gss.NoLimit)
		if err != nil {
			fmt.Println("error getting all settings")
			os.Exit(1)
		}
		fmt.Println(str)
		fmt.Println("=================================================")
	}

	var athenaClient *athena.Athena
	var s3_client *s3.S3

	if processConfig.HasAWSResource() {
		awsSession, err := session.NewSessionWithOptions(processConfig.AWSSessionOptions())
		if err != nil {
			fmt.Println(errors.Wrap(err, "error connecting to AWS"))
			os.Exit(1)
		}

		if processConfig.HasAthenaStoredQuery() {
			athenaClient = athena.New(awsSession)
		}

		if processConfig.HasS3Bucket() {
			s3_client = s3.New(awsSession)
		}
	}

	errorWriter, err := grw.WriteToResource(processConfig.ErrorDestination, processConfig.ErrorCompression, true, s3_client)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error creating error writer"))
		os.Exit(1)
	}

	infoWriter, err := grw.WriteToResource(processConfig.InfoDestination, processConfig.InfoCompression, true, s3_client)
	if err != nil {
		errorWriter.WriteError(errors.Wrap(err, "error creating log writer"))
		errorWriter.Close()
		os.Exit(1)
	}

	logger := rlogger.New(infoWriter, processConfig.InfoFormat, errorWriter, processConfig.ErrorFormat)

	errorsChannel := make(chan interface{}, 1000)
	messages := make(chan interface{}, 1000)
	logger.ListenInfo(messages, nil)

	if processConfig.ErrorDestination == processConfig.InfoDestination {
		go func(errorsChannel chan interface{}) {
			for err := range errorsChannel {
				messages <- err
			}
		}(errorsChannel)
	} else {
		logger.ListenError(errorsChannel, nil)
	}

	processConfig.Input.Init()

	var inputReader grw.ByteReadCloser
	if !processConfig.Input.IsAthenaStoredQuery() {
		r, inputMetadata, err := grw.ReadFromResource(
			processConfig.Input.Uri,
			processConfig.Input.Compression,
			processConfig.Input.ReaderBufferSize,
			false,
			s3_client)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error opening resource from uri "+processConfig.Input.Uri))
		}
		inputReader = r

		processConfig.Output.Init()

		if len(processConfig.Input.Format) == 0 {
			if inputMetadata != nil {
				if len(inputMetadata.ContentType) > 0 {
					switch inputMetadata.ContentType {
					case "application/json":
						processConfig.Input.Format = "json"
					case "application/vnd.geo+json":
						processConfig.Input.Format = "json"
					case "application/toml":
						processConfig.Input.Format = "toml"
					}
				}
			}
			if len(processConfig.Input.Format) == 0 && len(processConfig.Output.Format) > 0 {
				logger.Fatal("Error: Provided no --input-format and could not infer from resource.")
			}
		}
	}

	if len(processConfig.Input.Format) > 0 && len(processConfig.Output.Format) == 0 {
		logger.Fatal("Error: Provided input format but no output format.")
	}

	//fmt.Println("Output Format:", outputFormat)
	//fmt.Println("Output Compression:", outputCompression)

	if stream {

		if !(processConfig.Output.CanStream()) {
			logger.Fatal("output format " + processConfig.Output.Format + " is not compatible with streaming")
		}

		if processConfig.Output.IsEncrypted() {
			logger.Fatal("output passphrase is not compatible with streaming because it uses a block cipher")
		}

		// Stream Processing with Batch Input
		if processConfig.Input.IsEncrypted() || !(processConfig.Input.CanStream()) {

			inputBytes, err := util.DecryptReader(inputReader, processConfig.Input.Passphrase, processConfig.Input.Salt)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error decoding input"))
			}

			inputType, err := gss.GetType(inputBytes, processConfig.Input.Format)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error getting type for input"))
			}

			if !(inputType.Kind() == reflect.Array || inputType.Kind() == reflect.Slice) {
				logger.Fatal("input type cannot be streamed as it is not an array or slice but " + fmt.Sprint(inputType))
			}

			options := processConfig.InputOptions()
			options.Type = inputType
			inputObjects, err := options.DeserializeBytes(inputBytes, verbose)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error deserializing input using format "+processConfig.Input.Format))
			}

			dflNode, err := processConfig.Dfl.Node()
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error parsing"))
			}

			dflVars, err := processConfig.Dfl.Variables()
			if err != nil {
				logger.Fatal(err)
			}

			var wgObjects sync.WaitGroup
			var wgMessages sync.WaitGroup
			fatals := make(chan error, 1000)
			outputObjects := make(chan interface{}, 1000)
			//messages := make(chan interface, 1000)

			wgObjects.Add(1)
			wgMessages.Add(1)
			logger.ListenFatal(fatals)
			logger.ListenInfo(messages, &wgMessages)
			handleOutput(
				processConfig.Output,
				dflVars,
				outputObjects,
				fatals,
				messages,
				fileDescriptorLimit,
				&wgObjects,
				s3_client,
				verbose)

			inputObjectsValue := reflect.ValueOf(inputObjects)
			inputObjectsLength := inputObjectsValue.Len()
			for i := 0; i < inputObjectsLength; i++ {
				output, err := processObject(inputObjectsValue.Index(i).Interface(), dflNode, dflVars)
				if err != nil {
					logger.Fatal(errors.Wrap(err, "error processing object"))
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
			logger.Close()
			return
		}

		//dflNode, err := util.ParseDfl(dflUri, dflExpression)
		dflNode, err := processConfig.Dfl.Node()
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error parsing"))
		}

		dflVars, err := processConfig.Dfl.Variables()
		if err != nil {
			logger.Fatal(err)
		}

		var wgObjects sync.WaitGroup
		var wgMessages sync.WaitGroup
		errorsChannel := make(chan error, 1000)
		messages := make(chan interface{}, 1000)
		outputObjects := make(chan interface{}, 1000)

		wgObjects.Add(1)
		wgMessages.Add(1)
		logger.ListenFatal(errorsChannel)
		logger.ListenInfo(messages, &wgMessages)

		if processConfig.Input.IsAthenaStoredQuery() {

			athenaIterator, err := processAthenaInput(
				processConfig.Input.Uri,
				processConfig.Input.Limit,
				processConfig.Temp.Uri,
				processConfig.Output.Format,
				athenaClient,
				logger,
				verbose)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error processing athena input"))
			}

			handleOutput(
				processConfig.Output,
				dflVars,
				outputObjects,
				errorsChannel,
				messages,
				fileDescriptorLimit,
				&wgObjects,
				s3_client,
				verbose)

			inputCount := 0
			for {

				line, err := athenaIterator.Next()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						logger.Fatal(errors.Wrap(err, "error from athena iterator"))
					}
				}

				//messages <- "processing line: " + string(line)

				inputObject := map[string]interface{}{}
				err = json.Unmarshal(line, &inputObject)
				if err != nil {
					logger.Fatal(errors.Wrap(err, "error unmarshalling value from athena results: "+string(line)))
				}

				outputObject, err := processObject(inputObject, dflNode, dflVars)
				if err != nil {
					switch err.(type) {
					case *gss.ErrEmptyRow:
					default:
						logger.Fatal(errors.Wrap(err, "error processing object"))
					}
				} else {
					switch outputObject.(type) {
					case dfl.Null:
					default:
						outputObjects <- outputObject
					}
				}

				inputCount += 1
				if processConfig.Input.Limit > 0 && inputCount >= processConfig.Input.Limit {
					break
				}
			}
			messages <- "closing outputObjects"
			close(outputObjects)
			messages <- "waiting for wgObjects"
			wgObjects.Wait()
			messages <- "waiting for wgMessages"
			wgMessages.Wait()
			logger.Close()
			return

		}

		if len(processConfig.Input.Header) == 0 && (processConfig.Input.Format == "csv" || processConfig.Input.Format == "tsv") {
			inputBytes, err := inputReader.ReadBytes('\n')
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error reading header from resource"))
			}
			csvReader := csv.NewReader(bytes.NewReader(inputBytes))
			if processConfig.Input.Format == "tsv" {
				csvReader.Comma = '\t'
			}
			csvReader.LazyQuotes = processConfig.Input.LazyQuotes
			if len(processConfig.Input.Comment) > 1 {
				logger.Fatal("go's encoding/csv package only supports single character comment characters")
			} else if len(processConfig.Input.Comment) == 1 {
				csvReader.Comment = []rune(processConfig.Input.Comment)[0]
			}
			h, err := csvReader.Read()
			if err != nil {
				if err != io.EOF {
					logger.Fatal(errors.Wrap(err, "Error reading header from input with format csv"))
				}
			}
			processConfig.Input.Header = h
		}

		inputLines := make(chan []byte, 1000)

		handleInput(
			inputLines,
			processConfig.Input,
			dflNode,
			dflVars,
			outputObjects,
			processConfig.Output.Format,
			errorsChannel,
			verbose)

		handleOutput(
			processConfig.Output,
			dflVars,
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
					logger.Fatal("error reading line from resource")
				}
			}
			inputLines <- inputBytes
			inputCount += 1
			if processConfig.Input.Limit > 0 && inputCount >= processConfig.Input.Limit {
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
		logger.Close()
		return

	}

	// Batch Processing

	outputString := ""
	if processConfig.Input.IsAthenaStoredQuery() {

		athenaIterator, err := processAthenaInput(
			processConfig.Input.Uri,
			processConfig.Input.Limit,
			processConfig.Temp.Uri,
			processConfig.Output.Format,
			athenaClient,
			logger, verbose)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error processing athena input"))
		}

		outputObjects := make([]map[string]interface{}, 0)
		for {
			line, err := athenaIterator.Next()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					logger.Fatal(errors.Wrap(err, "error from athena iterator"))
				}
			}
			object := map[string]interface{}{}
			err = json.Unmarshal(line, &object)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error unmarshalling value from athena results: "+string(line)))
			}
			outputObjects = append(outputObjects, object)
		}

		if len(processConfig.Output.Format) == 0 {
			return // just exit if no output-format is given
		}

		str, err := gss.SerializeString(
			outputObjects,
			processConfig.Output.Format,
			processConfig.Output.Header,
			processConfig.Output.Limit)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error converting input"))
		}

		outputString = str

	} else {

		dflVars, err := processConfig.Dfl.Variables()
		if err != nil {
			logger.Fatal(err)
		}

		inputBytes, err := util.DecryptReader(inputReader, processConfig.Input.Passphrase, processConfig.Input.Salt)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error decrypting input"))
		}

		if len(processConfig.Output.Format) > 0 {

			dflNode, err := processConfig.Dfl.Node()
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error parsing"))
			}

			inputType, err := gss.GetType(inputBytes, processConfig.Input.Format)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error getting type for input"))
			}

			options := processConfig.Input.Options()
			options.Type = inputType
			inputObject, err := options.DeserializeBytes(inputBytes, verbose)
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error deserializing input using format "+processConfig.Input.Format))
			}

			var outputObject interface{}
			if dflNode != nil {
				_, filterObject, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, []string{"'", "\"", "`"})
				if err != nil {
					logger.Fatal(errors.Wrap(err, "error evaluating filter"))
				}
				outputObject = filterObject
			} else {
				outputObject = inputObject
			}

			str, err := processConfig.OutputOptions().SerializeString(gss.StringifyMapKeys(outputObject))
			if err != nil {
				logger.Fatal(errors.Wrap(err, "error converting input"))
			}

			outputString = str

		} else {
			outputString = string(inputBytes)
		}
	}

	err = processOutput(outputString, processConfig.Output, s3_client)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "error processing output"))
	}
	logger.Close()

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
	processCmd.Flags().String("input-format", "", "the input format: "+strings.Join(gss.Formats, ", "))
	processCmd.Flags().StringSlice("input-header", []string{}, "the input header, if the stdin input has no header.")
	processCmd.Flags().StringP("input-comment", "c", "", "the comment character for the input, e.g, #")
	processCmd.Flags().Bool("input-lazy-quotes", false, "allows lazy quotes for CSV and TSV")
	processCmd.Flags().String("input-passphrase", "", "input passphrase for AES-256 encryption")
	processCmd.Flags().String("input-salt", GO_RAILGUN_DEFAULT_SALT, "input salt for AES-256 encryption")
	processCmd.Flags().Int("input-reader-buffer-size", 4096, "the buffer size for the input reader")
	processCmd.Flags().Int("input-skip-lines", gss.NoSkip, "the number of lines to skip before processing")
	processCmd.Flags().Int("input-limit", gss.NoLimit, "maximum number of objects to read from input")

	processCmd.Flags().String("temp-uri", "", "the temporary uri for storing results")

	// Output Flags
	processCmd.Flags().StringP("output-uri", "o", "stdout", "the output uri (a dfl expression itself)")
	processCmd.Flags().StringP("output-compression", "", "", "the output compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	processCmd.Flags().StringP("output-format", "", "", "the output format: "+strings.Join(gss.Formats, ", "))
	processCmd.Flags().StringSliceP("output-header", "", []string{}, "the output header")
	processCmd.Flags().StringP("output-passphrase", "", "", "output passphrase for AES-256 encryption")
	processCmd.Flags().StringP("output-salt", "", "", "output salt for AES-256 encryption")
	processCmd.Flags().IntP("output-limit", "", gss.NoLimit, "maximum number of objects to send to output")
	processCmd.Flags().BoolP("output-append", "", false, "append to output files")
	processCmd.Flags().Bool("output-buffer-memory", false, "buffer output in memory")
	processCmd.Flags().Bool("output-mkdirs", false, "make directories if missing for output files")

	// DFL Flags
	processCmd.Flags().StringP("dfl-expression", "", "", "DFL expression to use")
	processCmd.Flags().StringP("dfl-uri", "", "", "URI to DFL file to use")
	processCmd.Flags().StringP("dfl-vars", "", "", "initial variables to use when evaluating DFL expression")

}
