// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	//"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	//"unicode"
)

import (
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
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
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/splitter"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-simple-serializer/pkg/iterator"
	"github.com/spatialcurrent/go-simple-serializer/pkg/jsonl"
	//"github.com/spatialcurrent/go-simple-serializer/pkg/properties"
	"github.com/spatialcurrent/go-simple-serializer/pkg/sv"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
)

import (
	//rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/athenaiterator"
	"github.com/spatialcurrent/railgun/pkg/config"
	"github.com/spatialcurrent/railgun/pkg/util"
)

import (
	"github.com/spatialcurrent/go-pipe/pkg/pipe"
)

var GO_RAILGUN_COMPRESSION_ALGORITHMS = []string{"none", "bzip2", "gzip", "snappy"}
var GO_RAILGUN_DEFAULT_SALT = "4F56C8C88B38CD8CD96BF8A9724F4BFE"

var processViper = viper.New()

//outputUri string, outputCompression string, outputAppend bool, outputPassphrase string, outputSalt string,
func processOutput(content string, output *config.Output, s3Client *s3.S3) error {
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

		outputWriter, err := grw.WriteToResource(output.Uri, output.Compression, output.Append, s3Client)
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

		err = outputWriter.Flush()
		if err != nil {
			return errors.Wrap(err, "error flushing output file.")
		}

		err = outputWriter.Close()
		if err != nil {
			return errors.Wrap(err, "error closing output file.")
		}

	}
	return nil
}

func processObject(object interface{}, node dfl.Node, vars map[string]interface{}, stringer stringify.Stringer) (interface{}, error) {
	if node != nil {
		_, o, err := node.Evaluate(
			vars,
			object,
			dfl.DefaultFunctionMap,
			[]string{"'", "\"", "`"})
		if err != nil {
			return "", errors.Wrap(err, "error evaluating filter")
		}
		return stringify.StringifyMapKeys(o, stringer)
	}
	return object, nil
}

/* handleInputFromSV reads input from a CSV/TSV separated values format
func handleInputFromSV(inputLines chan []byte, input *config.Input, node dfl.Node, vars map[string]interface{}, outputObjects chan interface{}, stringer stringify.Stringer, logger *gsl.Logger, verbose bool) {
	inputType := reflect.TypeOf(map[string]string{})
	for inputLine := range inputLines {
		inputObject, err := gss.DeserializeBytes(&gss.DeserializeInput{
			Bytes:      inputLine,
			Format:     input.Format,
			Header:     input.Header,
			Comment:    input.Comment,
			LazyQuotes: input.LazyQuotes,
			SkipLines:  gss.NoSkip,
			Limit:      1,
			Type:       inputType,
			Async:      false,
			Verbose:    verbose,
		})
		if err != nil {
			logger.Error(errors.Wrap(err, "error deserializing input from lines of bytes "+fmt.Sprint(inputLine)+""))
			continue
		}
		outputObject, err := processObject(inputObject, node, vars, stringer)
		if err != nil {
			switch err.(type) {
			case *gss.ErrEmptyRow:
			default:
				logger.Error(errors.Wrap(err, "error processing object"))
			}
		} else {
			switch outputObject.(type) {
			case dfl.Null:
			default:
				outputObjects <- outputObject
			}
		}
	}
	logger.Debug("input lines channel was closed")
	logger.Flush()
	close(outputObjects)
}

func handleInputFromJSONL(inputLines chan []byte, input *config.Input, node dfl.Node, vars map[string]interface{}, outputObjects chan interface{}, logger *gsl.Logger, verbose bool) {
	for inputLine := range inputLines {
		var inputType reflect.Type
		if str := strings.TrimLeftFunc(string(inputLine), unicode.IsSpace); len(str) > 0 && str[0] == '[' {
			inputType = reflect.TypeOf([]interface{}{})
		} else {
			inputType = reflect.TypeOf(map[string]interface{}{})
		}
		inputObject, err := gss.DeserializeBytes(&gss.DeserializeInput{
			Bytes:      inputLine,
			Format:     "json",
			Header:     input.Header,
			Comment:    input.Comment,
			LazyQuotes: input.LazyQuotes,
			SkipLines:  gss.NoSkip,
			Limit:      gss.NoLimit,
			Type:       inputType,
			Async:      false,
			Verbose:    verbose,
		})
		if err != nil {
			logger.Error(errors.Wrap(err, "error deserializing input"))
			continue
		}
		outputObject, err := processObject(inputObject, node, vars)
		if err != nil {
			switch err.(type) {
			case *gss.ErrEmptyRow:
			default:
				logger.Error(errors.Wrap(err, "error processing object"))
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
}
*/

func writeBuffersToFiles(buffers map[string]struct {
	Writer grw.ByteWriteCloser
	Buffer grw.Buffer
}, mkdirs bool, append bool, s3Client *s3.S3, logger *gsl.Logger) {
	logger.Info(map[string]interface{}{
		"msg":     "Writing buffers to files",
		"buffers": len(buffers),
	})
	logger.Flush()
	for outputPath, outputBuffer := range buffers {
		err := outputBuffer.Writer.Flush()
		if err != nil {
			logger.Error("error flushing output buffer for " + outputPath)
			logger.Flush()
		}
		err = outputBuffer.Writer.Close()
		if err != nil {
			logger.Error("error closing output buffer for " + outputPath)
			logger.Flush()
		}
		if mkdirs {
			err := grw.Mkdirs(filepath.Dir(outputPath))
			if err != nil {
				logger.FatalF("error creating parent directories for file at %q", outputPath)
			}
		}
		outputWriter, err := grw.WriteToResource(outputPath, "", append, s3Client)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error opening output file to "+outputPath))
		}
		//_, err = ioutil.Copy(outputWriter, snappy.NewReader(bytes.NewReader(outputBuffer.Buffer.Bytes())))
		//_, err = outputWriter.Write(outputBuffer.Buffer.Bytes())
		_, err = io.Copy(outputWriter, outputBuffer.Buffer)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "error writing buffer to output file to "+outputPath))
		}
		err = outputWriter.Flush()
		if err != nil {
			logger.Fatal(errors.Wrapf(err, "error flushing output file at %q", outputPath))
		}
		err = outputWriter.Close()
		if err != nil {
			logger.Fatal(errors.Wrapf(err, "error closing output file at %q, ", outputPath))
		}
		// delete output buffer and writer, since done writing to file
		delete(buffers, outputPath)
	}
	logger.Debug("Done writing buffers to files")
	logger.Flush()
}

func handleOutputWithMemoryBuffer(output *config.Output, outputNode dfl.Node, outputVars map[string]interface{}, objects chan interface{}, fileDescriptorLimit int, wg *sync.WaitGroup, s3Client *s3.S3, stringer stringify.Stringer, logger *gsl.Logger, verbose bool) error {

	if verbose {
		logger.Debug("handleOutputWithMemoryBuffer")
		logger.Flush()
	}

	outputLines := make(chan struct {
		Path string
		Line string
	}, 1000)

	outputPathBuffersMutex := &sync.RWMutex{}
	outputPathBuffers := map[string]struct {
		Writer grw.ByteWriteCloser
		Buffer grw.Buffer
	}{}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Fatal(map[string]interface{}{
					"msg":   "Recovered from panic",
					"value": r,
				})
			}
		}()
		for line := range outputLines {

			/*
				logger.Debug(map[string]string{
					"msg":  "Writing output line",
					"path": line.Path,
					"line": line.Line,
				})
				logger.Flush()
			*/

			outputPathBuffersMutex.Lock()
			if _, ok := outputPathBuffers[line.Path]; !ok {
				//outputWriter, outputBuffer, err := grw.WriteSnappyBytes(output.Compression)
				outputWriter, outputBuffer, err := grw.WriteBytes(output.Compression)
				if err != nil {
					panic(err)
				}

				outputPathBuffers[line.Path] = struct {
					Writer grw.ByteWriteCloser
					Buffer grw.Buffer
				}{Writer: outputWriter, Buffer: outputBuffer}

				logger.Debug(map[string]string{
					"msg":  "Created buffer for path",
					"path": line.Path,
				})
				logger.Flush()
			}
			outputPathBuffersMutex.Unlock()

			outputPathBuffersMutex.RLock()
			_, err := outputPathBuffers[line.Path].Writer.WriteLineSafe(line.Line)
			if err != nil {
				panic(err)
			}
			outputPathBuffersMutex.RUnlock()
		}

		logger.Debug("Done processing output lines")
		logger.Flush()

		if len(outputPathBuffers) > 0 {
			logger.Debug("writing buffers to files")
			logger.Flush()
			writeBuffersToFiles(outputPathBuffers, output.Mkdirs, output.Append, s3Client, logger)
		}
		wg.Done()
	}()

	if verbose {
		logger.Debug("starting to process objects")
		logger.Flush()
	}

	go func() {
		for object := range objects {

			/*
				logger.Debug(map[string]interface{}{
					"msg":  "Processing Object",
					"obj": object,
				})
				logger.Flush()
			*/

			outputPath, err := processObject(object, outputNode, outputVars, stringer)
			if err != nil {
				logger.Error(errors.Wrap(err, "Error writing string to output file"))
				logger.Flush()
				break
			}

			if reflect.TypeOf(outputPath).Kind() != reflect.String {
				logger.Error(errors.Wrap(err, "output path is not a string"))
				logger.Flush()
				break
			}

			outputPathString, err := homedir.Expand(outputPath.(string))
			if err != nil {
				logger.Error(errors.Wrap(err, "output path cannot be expanded"))
				logger.Flush()
				break
			}

			line, err := formatObject(object, output.Format, output.Header, output.LineSeparator)
			if err != nil {
				logger.Error(errors.Wrap(err, "error formatting object"))
				logger.Flush()
				break
			}

			outputLines <- struct {
				Path string
				Line string
			}{Path: outputPathString, Line: string(line)}
		}
		close(outputLines)
		logger.Debug("output lines closed")
		logger.Flush()
	}()

	return nil
}

func handleOutput(output *config.Output, outputVars map[string]interface{}, objects chan interface{}, fileDescriptorLimit int, wg *sync.WaitGroup, s3Client *s3.S3, stringer stringify.Stringer, logger *gsl.Logger, verbose bool) error {

	if verbose {
		logger.Debug("handleOutput")
		logger.Flush()
	}

	if output.Uri == "stdout" || output.Uri == "stderr" {
		it, err := pipe.NewChannelIterator(objects)
		if err != nil {
			return errors.Wrap(err, "error creating channel iterator")
		}
		b := pipe.NewBuilder().
			Input(it).
			Transform(func(inputObject interface{}) (interface{}, error) {
				b, err := formatObject(inputObject, output.Format, output.Header, output.LineSeparator)
				if err != nil {
					return "", err
				}
				return string(b), nil
			}).
			Error(func(err error) error {
				return errors.Wrap(err, "error formatting object")
			})

		switch output.Uri {
		case "stdout":
			b = b.Output(pipe.WriterStdout)
		case "stderr":
			b = b.Output(pipe.WriterStderr)
		}

		go func() {
			err := b.Run()
			if err != nil {
				logger.Error(err)
			}
			wg.Done()
		}()
	}

	n, err := dfl.ParseCompile(output.Uri)
	if err != nil {
		return errors.Wrap(err, "error parsing output uri: "+output.Uri)
	}
	outputNode := n

	if output.BufferMemory {
		err := handleOutputWithMemoryBuffer(output, outputNode, outputVars, objects, fileDescriptorLimit, wg, s3Client, stringer, logger, verbose)
		if err != nil {
			return errors.Wrap(err, "error processing output using memory buffers")
		}
		return nil
	}

	outputLines := make(chan struct {
		Path string
		Line string
	}, 1000)

	outputPathSemaphores := map[string]chan struct{}{}
	outputFileDescriptorSemaphore := make(chan struct{}, fileDescriptorLimit)

	var outputPathMutex = &sync.Mutex{}
	getOutputPathSemaphore := func(outputPathString string) chan struct{} {
		outputPathMutex.Lock()
		if _, ok := outputPathSemaphores[outputPathString]; !ok {
			outputPathSemaphores[outputPathString] = make(chan struct{}, 1)
		}
		outputPathMutex.Unlock()
		return outputPathSemaphores[outputPathString]
	}

	if verbose {
		logger.Debug("created semaphores using file descriptor limit " + fmt.Sprint(fileDescriptorLimit))
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

				if output.Mkdirs {
					grw.Mkdirs(filepath.Dir(line.Path))
					if err != nil {
						logger.Fatal(errors.Wrap(err, "error creating parent directories for "+line.Path))
					}
				}

				outputWriter, err := grw.WriteToResource(line.Path, output.Compression, true, s3Client)
				if err != nil {
					<-outputPathSemaphore
					<-outputFileDescriptorSemaphore
					logger.Fatal(errors.Wrap(err, "error opening file at path "+line.Path))
				}

				_, err = outputWriter.WriteLine(line.Line)
				if err != nil {
					logger.Fatal(errors.Wrap(err, "Error writing string to output file"))
				}

				err = outputWriter.Flush()
				if err != nil {
					logger.Fatal(errors.Wrap(err, "error flushing output file."))
				}

				err = outputWriter.Close()
				if err != nil {
					logger.Fatal(errors.Wrap(err, "error closing output file."))
				}

				<-outputPathSemaphore
				<-outputFileDescriptorSemaphore

			}(&wgLines, line)
		}
		logger.Info("* closing file descriptor semaphore")
		close(outputFileDescriptorSemaphore)
		logger.Info("* waiting for wgLines to be done")
		wgLines.Wait()
	}()

	if verbose {
		logger.Debug("starting to process objects")
		logger.Flush()
	}

	go func() {
		for object := range objects {
			outputPath, err := processObject(object, outputNode, outputVars, stringer)
			if err != nil {
				logger.Error(errors.Wrap(err, "Error writing string to output file"))
				break
			}

			if reflect.TypeOf(outputPath).Kind() != reflect.String {
				logger.Error(errors.Wrap(err, "output path is not a string"))
				break
			}

			outputPathString, err := homedir.Expand(outputPath.(string))
			if err != nil {
				logger.Error(errors.Wrap(err, "output path cannot be expanded"))
				break
			}

			line, err := formatObject(object, output.Format, output.Header, output.LineSeparator)
			if err != nil {
				logger.Error(errors.Wrap(err, "error formatting object"))
				break
			}

			outputLines <- struct {
				Path string
				Line string
			}{Path: outputPathString, Line: string(line)}

		}
		logger.Info("closing output lines")
		close(outputLines)
	}()

	return nil
}

func formatObject(object interface{}, format string, header []interface{}, lineSeparator string) ([]byte, error) {
	if format == "jsonl" {
		b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
			Object:            object,
			Format:            "json",
			Header:            header,
			Limit:             gss.NoLimit,
			Pretty:            false,
			LineSeparator:     lineSeparator,
			KeyValueSeparator: "=",
		})
		if err != nil {
			return make([]byte, 0), errors.Wrap(err, "error serializing object")
		}
		return b, nil
	}
	b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            object,
		Format:            format,
		Header:            header,
		Limit:             gss.NoLimit,
		Pretty:            false,
		LineSeparator:     lineSeparator,
		KeyValueSeparator: "=",
	})
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error serializing object")
	}
	return b, nil
}

func processAthenaInput(inputUri string, inputLimit int, tempUri string, outputFormat string, athenaClient *athena.Athena, logger *gsl.Logger, verbose bool) (*athenaiterator.AthenaIterator, error) {

	if !strings.HasPrefix(tempUri, "s3://") {
		return nil, errors.New("temporary uri must be an S3 object")
	}

	_, queryName := splitter.SplitUri(inputUri)
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

// processSinkToStream processes reads from a batch input and writes to a stream.
func processSinkToStream(inputReader grw.ByteReadCloser, processConfig *config.Process, s3Client *s3.S3, stringer stringify.Stringer, logger *gsl.Logger) error {
	if processConfig.Verbose {
		logger.Info("Processing as sink to stream.")
		logger.Flush()
	}
	inputBytes, err := util.DecryptReader(inputReader, processConfig.Input.Passphrase, processConfig.Input.Salt)
	if err != nil {
		return errors.Wrap(err, "error decoding input")
	}

	inputType, err := gss.GetType(inputBytes, processConfig.Input.Format)
	if err != nil {
		return errors.Wrap(err, "error getting type for input")
	}

	if !(inputType.Kind() == reflect.Array || inputType.Kind() == reflect.Slice) {
		return errors.New("input type cannot be streamed as it is not an array or slice but " + fmt.Sprint(inputType))
	}

	inputObjects, err := gss.DeserializeBytes(&gss.DeserializeBytesInput{
		Bytes:         inputBytes,
		Format:        processConfig.Input.Format,
		Header:        processConfig.Input.Header,
		Comment:       processConfig.Input.Comment,
		LazyQuotes:    processConfig.Input.LazyQuotes,
		SkipLines:     processConfig.Input.SkipLines,
		Limit:         processConfig.Input.Limit,
		Type:          inputType,
		LineSeparator: processConfig.Input.LineSeparator,
	})
	if err != nil {
		return errors.Wrap(err, "error deserializing input using format "+processConfig.Input.Format)
	}

	dflNode, err := processConfig.Dfl.Node()
	if err != nil {
		return errors.Wrap(err, "error parsing")
	}

	dflVars, err := processConfig.Dfl.Variables()
	if err != nil {
		return errors.Wrap(err, "error getting variable from process config")
	}

	var wgObjects sync.WaitGroup
	outputObjects := make(chan interface{}, 1000)

	wgObjects.Add(1)
	err = handleOutput(
		processConfig.Output,
		dflVars,
		outputObjects,
		processConfig.FileDescriptorLimit,
		&wgObjects,
		s3Client,
		stringer,
		logger,
		processConfig.Verbose)
	if err != nil {
		return errors.Wrap(err, "error procssing output")
	}

	inputObjectsValue := reflect.ValueOf(inputObjects)
	inputObjectsLength := inputObjectsValue.Len()
	for i := 0; i < inputObjectsLength; i++ {
		output, err := processObject(inputObjectsValue.Index(i).Interface(), dflNode, dflVars, stringer)
		if err != nil {
			return errors.Wrap(err, "error processing object")
		}
		switch output.(type) {
		case dfl.Null:
		default:
			outputObjects <- output
		}
	}
	close(outputObjects)
	wgObjects.Wait()

	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "ended",
		})
	}
	return nil // exits function
}

func processAthenaToStream(processConfig *config.Process, dflVars map[string]interface{}, dflNode dfl.Node, athenaClient *athena.Athena, s3Client *s3.S3, stringer stringify.Stringer, logger *gsl.Logger) error {
	var wgObjects sync.WaitGroup
	outputObjects := make(chan interface{}, 1000)
	wgObjects.Add(1)

	if processConfig.Verbose {
		logger.Info("Processing as athena to stream.")
		logger.Flush()
	}
	athenaIterator, err := processAthenaInput(
		processConfig.Input.Uri,
		processConfig.Input.Limit,
		processConfig.Temp.Uri,
		processConfig.Output.Format,
		athenaClient,
		logger,
		processConfig.Verbose)
	if err != nil {
		return errors.Wrap(err, "error processing athena input")
	}

	err = handleOutput(
		processConfig.Output,
		dflVars,
		outputObjects,
		processConfig.FileDescriptorLimit,
		&wgObjects,
		s3Client,
		stringer,
		logger,
		processConfig.Verbose)
	if err != nil {
		return errors.Wrap(err, "error procssing output")
	}

	inputCount := 0
	for {

		line, err := athenaIterator.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return errors.Wrap(err, "error from athena iterator")
			}
		}

		//messages <- "processing line: " + string(line)

		inputObject := map[string]interface{}{}
		err = json.Unmarshal(line, &inputObject)
		if err != nil {
			return errors.Wrap(err, "error unmarshalling value from athena results: "+string(line))
		}

		outputObject, err := processObject(inputObject, dflNode, dflVars, stringer)
		if err != nil {
			if err != gss.ErrEmptyRow {
				return errors.Wrap(err, "error processing object")
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
	logger.Info("closing outputObjects")
	close(outputObjects)
	logger.Info("waiting for wgObjects")
	wgObjects.Wait()

	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "ended",
		})
	}
	return nil
}

func createTransformFunction(dflNode dfl.Node, dflVars map[string]interface{}) func(inputObject interface{}) (interface{}, error) {
	return func(inputObject interface{}) (interface{}, error) {
		_, outputObject, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
		if err != nil {
			return nil, errors.Wrap(err, "error evaluating filter")
		}
		if _, ok := outputObject.(dfl.Null); ok {
			return nil, nil
		}
		return outputObject, nil
	}
}

func createOutputFunction(outputNode dfl.Node, dflVars map[string]interface{}, outputFormat string, outputCompression string, outputHeader []interface{}, outputKeySerializer stringify.Stringer, outputValueSerializer stringify.Stringer, outputLineSeparator string, outputPathBuffersMutex *sync.RWMutex, outputPathWriters map[string]pipe.Writer, outputPathBuffers map[string]grw.Buffer) func(object interface{}) error {
	return func(object interface{}) error {

		_, outputPath, err := outputNode.Evaluate(dflVars, object, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
		if err != nil {
			return errors.Wrap(err, "error evaluating filter")
		}

		outputPathString, ok := outputPath.(string)
		if !ok {
			return errors.New("output path is not a string")
		}

		outputPathBuffersMutex.Lock()
		if _, ok := outputPathBuffers[outputPathString]; !ok {

			outputWriter, outputBuffer, err := grw.WriteBytes(outputCompression)
			if err != nil {
				return errors.Wrapf(err, "error writing to bytes for compression %q", outputCompression)
			}

			if outputFormat == "csv" || outputFormat == "tsv" {
				separator, err := sv.FormatToSeparator(outputFormat)
				if err != nil {
					return err
				}
				outputPathWriters[outputPathString] = sv.NewWriter(outputWriter, separator, outputHeader, outputKeySerializer, outputValueSerializer, true, false)
			} else if outputFormat == "jsonl" {
				outputPathWriters[outputPathString] = jsonl.NewWriter(outputWriter, outputLineSeparator, outputKeySerializer, false)
			} else {
				return fmt.Errorf("cannot create streaming writer for format %q", outputFormat)
			}

			outputPathBuffers[outputPathString] = outputBuffer
		}
		outputPathBuffersMutex.Unlock()

		err = outputPathWriters[outputPathString].WriteObject(object)
		if err != nil {
			return errors.Wrap(err, "error writing object to buffer")
		}

		err = outputPathWriters[outputPathString].Flush()
		if err != nil {
			return errors.Wrap(err, "error flushing object to buffer")
		}

		return nil
	}
}

func processStreamToStream(inputReader grw.ByteReadCloser, processConfig *config.Process, dflVars map[string]interface{}, dflNode dfl.Node, s3Client *s3.S3, logger *gsl.Logger) error {
	if processConfig.Verbose {
		logger.Info("Processing as stream to stream.")
		logger.Flush()
	}

	p := pipe.NewBuilder().OutputLimit(processConfig.Output.Limit).Filter(func(object interface{}) (bool, error) {
		return object != nil, nil
	})

	if len(processConfig.Input.LineSeparator) != 1 {
		return fmt.Errorf("invalid line separator %q with length %d", processConfig.Input.LineSeparator, len(processConfig.Input.LineSeparator))
	}

	it, err := iterator.NewIterator(&iterator.NewIteratorInput{
		Reader:            inputReader,
		Format:            processConfig.Input.Format,
		Header:            processConfig.Input.Header,
		SkipLines:         processConfig.Input.SkipLines,
		SkipBlanks:        true,
		SkipComments:      true,
		Comment:           processConfig.Input.Comment,
		Trim:              true,
		LazyQuotes:        processConfig.Input.LazyQuotes,
		Limit:             processConfig.Input.Limit,
		LineSeparator:     []byte(processConfig.Input.LineSeparator)[0],
		KeyValueSeparator: processConfig.Input.KeyValueSeparator,
		DropCR:            processConfig.Input.DropCR,
	})
	if err != nil {
		if err == io.EOF {
			return io.EOF
		}
		return errors.Wrap(err, "error creating iterator")
	}
	p = p.Input(it)

	if dflNode != nil {
		p = p.Transform(createTransformFunction(dflNode, dflVars))
	}

	outputUri := processConfig.Output.Uri
	outputFormat := processConfig.Output.Format
	outputCompression := processConfig.Output.Compression
	outputHeader := make([]interface{}, 0)
	for _, str := range processConfig.Output.Header {
		outputHeader = append(outputHeader, str)
	}
	outputAppend := processConfig.Output.Append
	outputOverwrite := processConfig.Output.Overwrite
	outputMkdirs := processConfig.Output.Mkdirs
	outputKeySerializer := processConfig.Output.KeySerializer()
	outputValueSerializer := processConfig.Output.ValueSerializer()

	outputLineSeparator := processConfig.Output.LineSeparator
	outputSorted := processConfig.Output.Sorted
	outputReversed := processConfig.Output.Reversed
	outputPretty := processConfig.Output.Pretty

	if outputUri == "stdout" || outputUri == "stderr" {
		var outputWriter io.Writer
		switch outputUri {
		case "stdout":
			outputWriter = os.Stdout
		case "stderr":
			outputWriter = os.Stderr
		}
		switch outputFormat {
		case "csv", "tsv":
			separator, err := sv.FormatToSeparator(outputFormat)
			if err != nil {
				return err
			}
			w := sv.NewWriter(outputWriter, separator, outputHeader, outputKeySerializer, outputValueSerializer, outputSorted, outputReversed)
			err = p.Output(w).Run()
			if err != nil {
				return errors.Wrap(err, "error processing iterator to writer")
			}
		case "jsonl":
			err = p.Output(jsonl.NewWriter(outputWriter, outputLineSeparator, outputKeySerializer, outputPretty)).Run()
			if err != nil {
				return errors.Wrap(err, "error processing iterator to writer")
			}
		}
	} else {

		outputNode, err := dfl.ParseCompile(outputUri)
		if err != nil {
			return errors.Wrap(err, "error parsing output uri: "+outputUri)
		}

		outputPathBuffersMutex := &sync.RWMutex{}
		outputPathWriters := map[string]pipe.Writer{}
		outputPathBuffers := map[string]grw.Buffer{}

		p = p.Output(
			pipe.NewFunctionWriter(
				createOutputFunction(
					outputNode,
					dflVars,
					outputFormat,
					outputCompression,
					outputHeader,
					outputKeySerializer,
					outputValueSerializer,
					outputLineSeparator,
					outputPathBuffersMutex,
					outputPathWriters,
					outputPathBuffers,
				),
			),
		)

		logger.Info(map[string]interface{}{
			"msg": "starting pipeline",
		})
		logger.Flush()

		err = p.Run()
		if err != nil {
			return errors.Wrap(err, "error processing iterator to writer")
		}

		logger.Info(map[string]interface{}{
			"msg":               "done running pipeline",
			"outputPathWriters": len(outputPathWriters),
			"outputPathBuffers": len(outputPathBuffers),
		})
		logger.Flush()

		for _, outputWriter := range outputPathWriters {

			err := outputWriter.Flush()
			if err != nil {
				return errors.Wrap(err, "error flushing output writer")
			}

			if closer, ok := outputWriter.(io.Closer); ok {
				err = closer.Close()
				if err != nil {
					return errors.Wrap(err, "error flushing output writer")
				}
			}

		}

		err = grw.WriteBuffers(&grw.WriteBuffersInput{
			Buffers:   outputPathBuffers,
			Algorithm: "none",
			Overwrite: outputOverwrite,
			Append:    outputAppend,
			Mkdirs:    outputMkdirs,
			S3Client:  s3Client,
		})
		if err != nil {
			return errors.Wrap(err, "error writing buffers to files")
		}
	}

	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "ended",
		})
	}

	return nil
}

/*
func processStreamToStreamOld(inputReader grw.ByteReadCloser, processConfig *config.Process, dflVars map[string]interface{}, dflNode dfl.Node, s3Client *s3.S3, logger *gsl.Logger) error {
	if processConfig.Verbose {
		logger.Info("Processing as stream to stream.")
		logger.Flush()
	}

	if len(processConfig.Input.Header) == 0 && (processConfig.Input.Format == "csv" || processConfig.Input.Format == "tsv") {
		inputBytes, err := inputReader.ReadBytes('\n')
		if err != nil {
			return errors.Wrap(err, "error reading header from resource")
		}
		csvReader := csv.NewReader(bytes.NewReader(inputBytes))
		if processConfig.Input.Format == "tsv" {
			csvReader.Comma = '\t'
		}
		csvReader.LazyQuotes = processConfig.Input.LazyQuotes
		if len(processConfig.Input.Comment) > 1 {
			return errors.Wrap(&gss.ErrInvalidComment{Value: processConfig.Input.Comment}, "the standard go csv package only support single character comments")
		} else if len(processConfig.Input.Comment) == 1 {
			csvReader.Comment = []rune(processConfig.Input.Comment)[0]
		}
		h, err := csvReader.Read()
		if err != nil {
			if err != io.EOF {
				return errors.Wrap(err, "Error reading header from input with format csv")
			}
		}
		processConfig.Input.Header = h
	}

	outputObjects := make(chan interface{}, 1000)

	var wgObjects sync.WaitGroup
	wgObjects.Add(1)

	err := handleOutput(
		processConfig.Output,
		dflVars,
		outputObjects,
		processConfig.FileDescriptorLimit,
		&wgObjects,
		s3Client,
		logger,
		processConfig.Verbose)
	if err != nil {
		return errors.Wrap(err, "error procssing output")
	}

	inputLines := make(chan []byte, 1000)

	switch processConfig.Input.Format {
	case "jsonl":
		go handleInputFromJSONL(
			inputLines,
			processConfig.Input,
			dflNode,
			dflVars,
			outputObjects,
			logger,
			processConfig.Verbose)
	case "csv", "tsv":
		go handleInputFromSV(
			inputLines,
			processConfig.Input,
			dflNode,
			dflVars,
			outputObjects,
			logger,
			processConfig.Verbose)
	default:
		return errors.New("Invalid format for stream processing")
	}

	inputCount := 0
	for {
		inputBytes, err := inputReader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return errors.New("error reading line from resource")
			}
		}
		// If line is a blank line then continue to next.
		if len(inputBytes) == 0 || (len(inputBytes) == 1 && inputBytes[0] == '\n') || (len(inputBytes) == 2 && inputBytes[0] == '\r' && inputBytes[1] == '\n') {
			continue
		}
		inputLines <- inputBytes
		inputCount += 1
		logger.Debug(fmt.Sprintf("Lines Read: %d", inputCount))
		logger.Flush()
		if processConfig.Input.Limit > 0 && inputCount >= processConfig.Input.Limit {
			break
		}
	}
	err = inputReader.Close()
	if err != nil {
		return errors.Wrap(err, "error closing input")
	}
	close(inputLines)
	logger.Debug("Input closed.  Waiting for objects to be processed.")
	logger.Flush()
	wgObjects.Wait()

	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "ended",
		})
	}

	return nil
}
*/

func canStream(inputFormat string, outputFormat string, outputSorted bool) bool {
	if !outputSorted {
		if inputFormat == "csv" || inputFormat == "tsv" {
			if outputFormat == "csv" || outputFormat == "tsv" || outputFormat == "jsonl" {
				return true
			}
		} else if inputFormat == "jsonl" {
			if outputFormat == "jsonl" {
				return true
			}
		}
	}
	return false
}

func processAsStream(inputReader grw.ByteReadCloser, processConfig *config.Process, athenaClient *athena.Athena, s3Client *s3.S3, stringer stringify.Stringer, logger *gsl.Logger) error {

	if processConfig.Verbose {
		logger.Info("Processing as stream.")
		logger.Flush()
	}

	if !(processConfig.Output.CanStream()) {
		return errors.New("output format " + processConfig.Output.Format + " is not compatible with streaming")
	}

	if processConfig.Output.IsEncrypted() {
		return errors.New("output passphrase is not compatible with streaming because it uses a block cipher")
	}

	if !canStream(processConfig.Input.Format, processConfig.Output.Format, processConfig.Output.Sorted) {
		return errors.New("cannot stream with these inputs")
	}

	// Stream Processing with Batch Input
	if processConfig.Input.IsEncrypted() || !(processConfig.Input.CanStream()) {
		return processSinkToStream(inputReader, processConfig, s3Client, stringer, logger)
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

	if processConfig.Input.IsAthenaStoredQuery() {
		return processAthenaToStream(processConfig, dflVars, dflNode, athenaClient, s3Client, stringer, logger)
	}

	return processStreamToStream(inputReader, processConfig, dflVars, dflNode, s3Client, logger)
}

func processAsBatch(inputReader grw.ByteReadCloser, processConfig *config.Process, s3Client *s3.S3, logger *gsl.Logger) error {

	inputObject, err := gss.DeserializeReader(&gss.DeserializeReaderInput{
		Reader:          inputReader,
		Format:          processConfig.Input.Format,
		Header:          processConfig.Input.Header,
		Comment:         processConfig.Input.Comment,
		LazyQuotes:      processConfig.Input.LazyQuotes,
		SkipLines:       processConfig.Input.SkipLines,
		SkipBlanks:      processConfig.Input.SkipBlanks,
		SkipComments:    processConfig.Input.SkipComments,
		Trim:            processConfig.Input.Trim,
		Limit:           processConfig.Input.Limit,
		LineSeparator:   processConfig.Input.LineSeparator,
		DropCR:          processConfig.Input.DropCR,
		EscapePrefix:    processConfig.Input.EscapePrefix,
		UnescapeSpace:   processConfig.Input.UnescapeSpace,
		UnescapeNewLine: processConfig.Input.UnescapeNewLine,
		UnescapeColon:   processConfig.Input.UnescapeColon,
		UnescapeEqual:   processConfig.Input.UnescapeEqual,
	})
	if err != nil {
		return errors.Wrap(err, "error parsing input object")
	}

	dflNode, err := processConfig.Dfl.Node()
	if err != nil {
		return errors.Wrap(err, "error parsing dfl node")
	}

	var outputObject interface{}
	if dflNode != nil {

		dflVars, err := processConfig.Dfl.Variables()
		if err != nil {
			return errors.Wrap(err, "error getting dfl variables")
		}

		_, object, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, []string{"'", "\"", "`"})
		if err != nil {
			return errors.Wrap(err, "error evaluating filter")
		}

		outputObject = object
	} else {
		outputObject = inputObject
	}

	outputBytes, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            outputObject,
		Format:            processConfig.Output.Format,
		Header:            processConfig.Output.Header,
		Limit:             processConfig.Output.Limit,
		Pretty:            processConfig.Output.Pretty,
		Sorted:            processConfig.Output.Sorted,
		Reversed:          processConfig.Output.Reversed,
		LineSeparator:     processConfig.Output.LineSeparator,
		KeyValueSeparator: processConfig.Output.KeyValueSeparator,
		KeySerializer:     processConfig.Output.KeySerializer(),
		ValueSerializer:   processConfig.Output.ValueSerializer(),
		EscapePrefix:      processConfig.Output.EscapePrefix,
		EscapeSpace:       processConfig.Output.EscapeSpace,
		EscapeNewLine:     processConfig.Output.EscapeNewLine,
		EscapeEqual:       processConfig.Output.EscapeEqual,
		EscapeColon:       processConfig.Output.EscapeColon,
		ExpandHeader:      processConfig.Output.ExpandHeader,
	})

	if err != nil {
		return errors.Wrap(err, "error evaluating filter")
	}

	outputWriter, err := grw.WriteToResource(
		processConfig.Output.Uri,
		processConfig.Output.Compression,
		processConfig.Output.Append,
		s3Client)
	if err != nil {
		return errors.Wrap(err, "error opening output file")
	}

	_, err = outputWriter.Write(outputBytes)
	if err != nil {
		return errors.Wrap(err, "error writing to output file")
	}

	err = outputWriter.Flush()
	if err != nil {
		return errors.Wrap(err, "error flushing to output file")
	}

	err = outputWriter.Close()
	if err != nil {
		return errors.Wrap(err, "error writing to output file")
	}

	return nil
}

/*
func processAsBatchOld(inputReader grw.ByteReadCloser, processConfig *config.Process, athenaClient *athena.Athena, s3Client *s3.S3, logger *gsl.Logger) error {

	if processConfig.Verbose {
		logger.Info("Processing as batch.")
		logger.Flush()
	}

	outputString := ""
	if processConfig.Input.IsAthenaStoredQuery() {
		athenaIterator, err := processAthenaInput(
			processConfig.Input.Uri,
			processConfig.Input.Limit,
			processConfig.Temp.Uri,
			processConfig.Output.Format,
			athenaClient,
			logger,
			processConfig.Verbose)
		if err != nil {
			return errors.Wrap(err, "error processing athena input")
		}

		outputObjects := make([]map[string]interface{}, 0)
		for {
			line, err := athenaIterator.Next()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return errors.Wrap(err, "error from athena iterator")
				}
			}
			object := map[string]interface{}{}
			err = json.Unmarshal(line, &object)
			if err != nil {
				return errors.Wrap(err, "error unmarshalling value from athena results: "+string(line))
			}
			outputObjects = append(outputObjects, object)
		}

		if len(processConfig.Output.Format) == 0 {
			return nil // just exit if no output-format is given
		}

		outputBytes, err := gss.SerializeBytes(&gss.SerializeBytesInput{
			Object:  outputObjects,
			Format:  processConfig.Output.Format,
			Header:  processConfig.Output.Header,
			Limit:   processConfig.Output.Limit,
			Pretty:  processConfig.Output.Pretty,
			NewLine: processConfig.Output.NewLine,
		})
		if err != nil {
			return errors.Wrap(err, "error converting output")
		}

		outputString = string(outputBytes)

	} else {

		dflVars, err := processConfig.Dfl.Variables()
		if err != nil {
			return errors.Wrap(err, "error getting dfl variables")
		}

		inputBytes, err := util.DecryptReader(inputReader, processConfig.Input.Passphrase, processConfig.Input.Salt)
		if err != nil {
			return errors.Wrap(err, "error decrypting input")
		}

		if len(processConfig.Output.Format) > 0 {

			dflNode, err := processConfig.Dfl.Node()
			if err != nil {
				return errors.Wrap(err, "error parsing")
			}

			inputType, err := gss.GetType(inputBytes, processConfig.Input.Format)
			if err != nil {
				return errors.Wrap(err, "error getting type for input")
			}

			options := processConfig.Input.Options()
			options.Type = inputType
			inputObject, err := options.DeserializeBytes(inputBytes, processConfig.Verbose)
			if err != nil {
				return errors.Wrap(err, "error deserializing input using format "+processConfig.Input.Format)
			}

			var outputObject interface{}
			if dflNode != nil {
				_, filterObject, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, []string{"'", "\"", "`"})
				if err != nil {
					return errors.Wrap(err, "error evaluating filter")
				}
				outputObject = filterObject
			} else {
				outputObject = inputObject
			}

			str, err := processConfig.Output.Options().SerializeString(stringify.StringifyMapKeys(outputObject))
			if err != nil {
				return errors.Wrap(err, "error converting output")
			}

			outputString = str

		} else {
			outputString = string(inputBytes)
		}
	}

	err := processOutput(outputString, processConfig.Output, s3Client)
	if err != nil {
		return errors.Wrap(err, "error processing output")
	}

	return nil
}
*/

func processFunction(cmd *cobra.Command, args []string) {

	v := processViper

	//err := v.BindPFlags(cmd.PersistentFlags())
	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))

	verbose := v.GetBool("verbose")

	if verbose {
		printViperSettings(v)
	}

	errorInput := CheckInput(v)
	if errorInput != nil {
		panic(errorInput)
	}

	errorOutput := CheckOutput(v)
	if errorOutput != nil {
		panic(errorOutput)
	}

	processConfig := config.NewProcessConfig()
	config.LoadConfigFromViper(processConfig, v)

	if verbose {
		printConfig(processConfig)
	}

	var athenaClient *athena.Athena
	var s3Client *s3.S3

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
			s3Client = s3.New(awsSession)
		}
	}

	logger := gsl.CreateApplicationLogger(&gsl.CreateApplicationLoggerInput{
		ErrorDestination: v.GetString("error-destination"),
		ErrorCompression: processConfig.ErrorCompression,
		ErrorFormat:      processConfig.ErrorFormat,
		InfoDestination:  processConfig.InfoDestination,
		InfoCompression:  processConfig.InfoCompression,
		InfoFormat:       processConfig.InfoFormat,
		Verbose:          processConfig.Verbose,
	})

	start := time.Now()
	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "started",
			"ts":  start.Format(time.RFC3339),
		})
	}

	if processConfig.Timeout.Seconds() > 0 {
		deadline := time.Now().Add(processConfig.Timeout)
		logger.Debug(fmt.Sprintf("Deadline: %v", deadline))
		go func() {
			for {
				if time.Now().After(deadline) {
					logger.FatalF("program exceeded timeout %v", processConfig.Timeout)
				}
				time.Sleep(15 * time.Second)
			}
		}()
	}

	processConfig.Input.Init()

	var inputReader grw.ByteReadCloser
	if !processConfig.Input.IsAthenaStoredQuery() {
		r, inputMetadata, err := grw.ReadFromResource(
			processConfig.Input.Uri,
			processConfig.Input.Compression,
			processConfig.Input.ReaderBufferSize,
			s3Client)
		if err != nil {
			logger.Fatal(errors.Wrapf(err, "error opening resource from uri %q", processConfig.Input.Uri))
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

	stringer := stringify.NewDefaultStringer()

	if processConfig.Stream {
		err := processAsStream(inputReader, processConfig, athenaClient, s3Client, stringer, logger)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		err := processAsBatch(inputReader, processConfig, s3Client, logger)
		if err != nil {
			logger.Fatal(err)
		}
	}

	if processConfig.Time {
		end := time.Now()
		logger.Info(map[string]interface{}{
			"msg":      "ended",
			"ts":       end.Format(time.RFC3339),
			"duration": end.Sub(start).String(),
		})
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

	flag := processCmd.Flags()

	flag.BoolP("dry-run", "", false, "parse and compile expression, but do not evaluate against context")
	flag.BoolP("stream", "s", false, "stream process (context == row rather than encompassing array)")
	flag.Duration("timeout", 1*time.Minute, "If not zero, then sets the timeout for the program.")

	// Input Flags
	InitInputFlags(flag)

	processCmd.Flags().String("temp-uri", "", "the temporary uri for storing results")

	// Output Flags
	InitOutputFlags(flag, "")

	// DFL Flags
	InitDflFlags(flag)
}
