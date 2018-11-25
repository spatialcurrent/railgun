package logger

import (
	"fmt"
	"os"
	"sync"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader-writer/grw"
)

type Logger struct {
	infoWriter  grw.ByteWriteCloser
	errorWriter grw.ByteWriteCloser
}

func New(infoWriter grw.ByteWriteCloser, errorWriter grw.ByteWriteCloser) *Logger {
	return &Logger{
		infoWriter:  infoWriter,
		errorWriter: errorWriter,
	}
}

func NewLoggerFromConfig(infoDestination string, infoCompression string, errorDestination string, errorCompression string, s3_client *s3.S3) *Logger {

	errorWriter, err := grw.WriteToResource(errorDestination, errorCompression, true, s3_client)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error creating error writer"))
		os.Exit(1)
	}

	infoWriter, err := grw.WriteToResource(infoDestination, infoCompression, true, s3_client)
	if err != nil {
		errorWriter.WriteError(errors.Wrap(err, "error creating info writer"))
		errorWriter.Close()
		os.Exit(1)
	}

	return New(infoWriter, errorWriter)
}

func (l *Logger) Info(obj interface{}) {
	if err, ok := obj.(error); ok {
		l.infoWriter.WriteError(err)
	} else if line, ok := obj.(string); ok {
		l.infoWriter.WriteLine(line)
	} else {
		l.infoWriter.WriteLine(fmt.Sprint(obj))
	}
}

func (l *Logger) Error(obj interface{}) {
	if err, ok := obj.(error); ok {
		l.errorWriter.WriteError(err)
	} else if line, ok := obj.(string); ok {
		l.errorWriter.WriteLine(line)
	} else {
		l.errorWriter.WriteLine(fmt.Sprint(obj))
	}
}

func (l *Logger) Fatal(obj interface{}) {
	l.infoWriter.Flush()
	l.Error(obj)
	l.errorWriter.Flush()
	os.Exit(1)
}

func (l *Logger) Flush() {
	l.infoWriter.Flush()
	l.errorWriter.Flush()
}

func (l *Logger) Close() {
	l.Flush()
}

func (l *Logger) ListenInfo(messages chan interface{}, wg *sync.WaitGroup) {
	go func(messages chan interface{}) {
		for message := range messages {
			l.infoWriter.WriteLine(fmt.Sprint(message))
			l.infoWriter.Flush()
		}
		if wg != nil {
			wg.Done()
		}
	}(messages)
}

func (l *Logger) ListenError(c chan error) {
	go func() {
		for err := range c {
			l.Error(err)
		}
	}()
}

func (l *Logger) ListenFatal(c chan error) {
	go func() {
		for err := range c {
			l.Fatal(err)
		}
	}()
}
