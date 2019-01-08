package logger

import (
	//"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

import (
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

type Logger struct {
	levels  map[string]int        // level --> position in writers
	writers []grw.ByteWriteCloser // list of writers
	formats []string              // list of formats for each writer
}

func New(levels map[string]int, writers []grw.ByteWriteCloser, formats []string) *Logger {
	return &Logger{
		levels:  levels,
		writers: writers,
		formats: formats,
	}
}

func (l *Logger) Debug(obj interface{}) {
	level := "debug"
	position, ok := l.levels[level]
	if !ok {
		return
	}
	l.Line("debug", obj, l.writers[position], l.formats[position])
}

func (l *Logger) Info(obj interface{}) {
	level := "info"
	position, ok := l.levels[level]
	if !ok {
		return
	}
	l.Line("info", obj, l.writers[position], l.formats[position])
}

func (l *Logger) Warn(obj interface{}) {
	level := "warn"
	position, ok := l.levels[level]
	if !ok {
		return
	}
	l.Line("warn", obj, l.writers[position], l.formats[position])
}

func (l *Logger) Error(obj interface{}) {
	level := "error"
	position, ok := l.levels[level]
	if !ok {
		return
	}
	l.Line("error", obj, l.writers[position], l.formats[position])
}

func (l *Logger) Line(level string, obj interface{}, writer grw.ByteWriteCloser, format string) {
	line := ""

	if err, ok := obj.(error); ok {
		m := map[string]interface{}{
			"level": level,
			"msg":   strings.Replace(err.Error(), "\n", ": ", -1),
			"ts":    time.Now().Format(time.RFC3339),
		}
		line = gss.MustSerializeString(m, format, gss.NoHeader, gss.NoLimit) // #nosec
	} else if msg, ok := obj.(string); ok {
		m := map[string]interface{}{
			"level": level,
			"msg":   msg,
			"ts":    time.Now().Format(time.RFC3339),
		}
		line = gss.MustSerializeString(m, format, gss.NoHeader, gss.NoLimit) // #nosec
	} else if m, ok := obj.(map[string]string); ok {
		m["level"] = level
		m["ts"] = time.Now().Format(time.RFC3339)
		line = gss.MustSerializeString(m, format, gss.NoHeader, gss.NoLimit) // #nosec
	} else if m, ok := obj.(map[string]interface{}); ok {
		m["level"] = level
		m["ts"] = time.Now().Format(time.RFC3339)
		line = gss.MustSerializeString(m, format, gss.NoHeader, gss.NoLimit) // #nosec
	} else {
		line = gss.MustSerializeString(obj, format, gss.NoHeader, gss.NoLimit) // #nosec
	}

	if len(line) > 0 {
		writer.WriteLineSafe(line) // #nosec
	}
}

func (l *Logger) Fatal(obj interface{}) {
	for _, w := range l.writers {
		w.Lock()
	}
	for _, w := range l.writers {
		w.Flush() // #nosec
	}
	l.Error(obj) // #nosec
	for _, w := range l.writers {
		w.Flush() // #nosec
	}
	for _, w := range l.writers {
		w.Close() // #nosec
	}
	os.Exit(1)
}

func (l *Logger) Flush() {
	for _, w := range l.writers {
		w.FlushSafe() // #nosec
	}
}

func (l *Logger) Close() {
	for _, w := range l.writers {
		w.FlushSafe() // #nosec
	}
	for _, w := range l.writers {
		w.CloseSafe() // #nosec
	}
}

func (l *Logger) ListenInfo(messages chan interface{}, wg *sync.WaitGroup) {
	go func(messages chan interface{}) {
		for message := range messages {
			l.Info(message)
			l.Flush()
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
			l.Flush()
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
