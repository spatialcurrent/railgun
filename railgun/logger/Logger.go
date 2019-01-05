package logger

import (
	"fmt"
	"os"
	"strings"
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
		errorWriter.WriteError(errors.Wrap(err, "error creating info writer")) // #nosec
		errorWriter.Close()                                                    // #nosec
		os.Exit(1)
	}

	return New(infoWriter, infoFormat, errorWriter, errorFormat)
}

func (l *Logger) Debug(obj interface{}) {
	if err, ok := obj.(error); ok {
		m := map[string]interface{}{"level": "debug", "error": strings.Replace(err.Error(), "\n", ": ", -1)}
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if line, ok := obj.(string); ok {
		m := map[string]interface{}{"level": "debug", "message": line}
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]string); ok {
		m["level"] = "debug"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]interface{}); ok {
		m["level"] = "debug"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else {
		l.infoWriter.WriteLine(gss.MustSerializeString(obj, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	}
}

func (l *Logger) Info(obj interface{}) {
	if err, ok := obj.(error); ok {
		m := map[string]interface{}{"level": "info", "error": strings.Replace(err.Error(), "\n", ": ", -1)}
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if line, ok := obj.(string); ok {
		m := map[string]interface{}{"level": "info", "message": line}
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]string); ok {
		m["level"] = "info"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]interface{}); ok {
		m["level"] = "info"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else {
		l.infoWriter.WriteLine(gss.MustSerializeString(obj, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	}
}

func (l *Logger) Warn(obj interface{}) {
	if err, ok := obj.(error); ok {
		m := map[string]interface{}{"level": "warn", "error": strings.Replace(err.Error(), "\n", ": ", -1)}
		l.errorWriter.WriteLine(gss.MustSerializeString(m, l.errorFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if line, ok := obj.(string); ok {
		m := map[string]interface{}{"level": "warn", "message": line}
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]string); ok {
		m["level"] = "warn"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]interface{}); ok {
		m["level"] = "warn"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else {
		l.errorWriter.WriteLine(gss.MustSerializeString(obj, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	}
}

func (l *Logger) Error(obj interface{}) {
	if err, ok := obj.(error); ok {
		m := map[string]interface{}{"level": "error", "error": strings.Replace(err.Error(), "\n", ": ", -1)}
		l.errorWriter.WriteLine(gss.MustSerializeString(m, l.errorFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if line, ok := obj.(string); ok {
		m := map[string]interface{}{"level": "error", "message": line}
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]string); ok {
		m["level"] = "error"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else if m, ok := obj.(map[string]interface{}); ok {
		m["level"] = "error"
		l.infoWriter.WriteLine(gss.MustSerializeString(m, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	} else {
		l.errorWriter.WriteLine(gss.MustSerializeString(obj, l.infoFormat, gss.NoHeader, gss.NoLimit)) // #nosec
	}
}

func (l *Logger) Fatal(obj interface{}) {
	l.infoWriter.Flush()  // #nosec
	l.infoWriter.Close()  // #nosec
	l.Error(obj)          // #nosec
	l.errorWriter.Flush() // #nosec
	l.errorWriter.Close() // #nosec
	os.Exit(1)
}

func (l *Logger) Flush() {
	l.infoWriter.Flush()  // #nosec
	l.errorWriter.Flush() // #nosec
}

func (l *Logger) Close() {
	l.Flush()
	l.infoWriter.Close()  // #nosec
	l.errorWriter.Close() // #nosec
}

func (l *Logger) ListenInfo(messages chan interface{}, wg *sync.WaitGroup) {
	go func(messages chan interface{}) {
		for message := range messages {
			l.Info(message)
			l.infoWriter.Flush() // #nosec
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
			l.errorWriter.Flush() // #nosec
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
