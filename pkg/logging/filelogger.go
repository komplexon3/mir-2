package logging

import (
	"fmt"
	"os"
)

type fileLogger struct {
	file  *os.File
	level LogLevel
}

func NewFileLogger(level LogLevel, path string) (*fileLogger, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &fileLogger{
		level: level,
		file:  file,
	}, nil
}

func (l *fileLogger) Stop() error {
	err := l.file.Sync()
	l.file.Close()
	return err
}

func (l *fileLogger) Log(level LogLevel, text string, args ...interface{}) {
	// NOTE: copy-pasted from console logger
	// TODO: cleanup
	if level < l.level {
		return
	}

	fmt.Fprint(l.file, text)
	for i := 0; i < len(args); i++ {
		if i+1 < len(args) {
			switch args[i+1].(type) {
			case []byte:
				// Print byte arrays in base 16 encoding.
				fmt.Fprintf(l.file, " %s=%x", args[i], args[i+1])
			default:
				// Print all other types using the Go default format.
				fmt.Fprintf(l.file, " %s=%v", args[i], args[i+1])
			}
			i++
		} else {
			fmt.Printf(" %s=%%MISSING%%", args[i])
		}
	}
	fmt.Fprintf(l.file, "\n")
}

func (l *fileLogger) IsConcurrent() bool {
	return false
}

func (l *fileLogger) MinLevel() LogLevel {
	return l.level
}
