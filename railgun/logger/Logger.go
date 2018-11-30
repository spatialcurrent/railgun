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
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

type Logger struct {
	infoWriter  grw.ByteWriteCloser
	infoFormat  string
	errorWriter grw.ByteWriteCloser
	errorFormat string
}

func New(infoWriter grw.ByteWriteCloser, infoFormat string, errorWriter grw.ByteWriteCloser, errorFormat string) *Logger {
	return &Logger{
		infoWriter:  infoWriter,
		infoFormat:  infoFormat,
		errorWriter: errorWriter,
		errorFormat: errorFormat,
	}
}

func NewLoggerFromConfig(infoDestination string, infoCompression string, infoFormat string, errorDestination string, errorCompression string, errorFormat string, s3_client *s3.S3) *Logger {

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

	return New(infoWriter, infoFormat, errorWriter, errorFormat)
}

func (l *Logger) Info(obj interface{}) {
	if err, ok := obj.(error); ok {
		l.infoWriter.WriteLine(gss.MustSerializeString(map[string]interface{}{"error": err.Error()}, l.infoFormat, gss.NoHeader, gss.NoLimit))
	} else if line, ok := obj.(string); ok {
		l.infoWriter.WriteLine(line)
	} else {
		l.infoWriter.WriteLine(gss.MustSerializeString(obj, l.infoFormat, gss.NoHeader, gss.NoLimit))
	}
}

func (l *Logger) Error(obj interface{}) {
	if err, ok := obj.(error); ok {
		l.errorWriter.WriteLine(gss.MustSerializeString(map[string]interface{}{"error": err.Error()}, l.errorFormat, gss.NoHeader, gss.NoLimit))
	} else if line, ok := obj.(string); ok {
		l.errorWriter.WriteLine(line)
	} else {
		l.errorWriter.WriteLine(gss.MustSerializeString(obj, l.infoFormat, gss.NoHeader, gss.NoLimit))
	}
}

func (l *Logger) Fatal(obj interface{}) {
	l.infoWriter.Flush()
	l.infoWriter.Close()
	l.Error(obj)
	l.errorWriter.Flush()
	l.errorWriter.Close()
	os.Exit(1)
}

func (l *Logger) Flush() {
	l.infoWriter.Flush()
	l.errorWriter.Flush()
}

func (l *Logger) Close() {
	l.Flush()
	l.infoWriter.Close()
	l.errorWriter.Close()
}

func (l *Logger) ListenInfo(messages chan interface{}, wg *sync.WaitGroup) {
	go func(messages chan interface{}) {
		for message := range messages {
			l.Info(message)
			l.infoWriter.Flush()
		}
		if wg != nil {
			wg.Done()
		}
	}(messages)
}

func (l *Logger) ListenError(messages chan interface{}, wg *sync.WaitGroup) {
	go func(messages chan interface{}) {
		for message := range messages {
			l.Error(message)
			l.errorWriter.Flush()
		}
		if wg != nil {
			wg.Done()
		}
	}(messages)
}

func (l *Logger) ListenFatal(c chan error) {
	go func() {
		for err := range c {
			l.Fatal(err)
		}
	}()
}
