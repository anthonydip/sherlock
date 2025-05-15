package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

var GlobalLogger *Logger

const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconVerbose = "…"
	IconDebug   = "»"
)

type Logger struct {
	verbose      bool
	debug        bool
	colors       bool
	output       io.Writer
	verboseColor *color.Color
	errorColor   *color.Color
	warnColor    *color.Color
	successColor *color.Color
	debugColor   *color.Color
	fileColor    *color.Color
	lineColor    *color.Color
	mutex        sync.Mutex
}

// New creates a configured logger instance
func New(verbose bool, debug bool, useColors bool) *Logger {
	if noColor := os.Getenv("NO_COLOR") != ""; noColor {
		useColors = false
	}

	return &Logger{
		verbose:      verbose,
		debug:        debug,
		colors:       useColors,
		output:       os.Stdout,
		verboseColor: color.New(color.FgCyan),
		errorColor:   color.New(color.FgRed, color.Bold),
		warnColor:    color.New(color.FgYellow, color.Bold),
		successColor: color.New(color.FgGreen, color.Bold),
		debugColor:   color.New(color.Faint, color.FgBlue),
		fileColor:    color.New(color.Faint, color.FgWhite),
		lineColor:    color.New(color.FgHiBlue),
	}
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
}

func (l *Logger) SetVerbose(v bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.verbose = v
}

// SetOutput changes the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.output = w

	// Only check terminal if colors were enabled
	if l.colors {
		l.colors = isTerminal(w)
	}
}

// Debugf prints formatted debug message when debug enabled
func (l *Logger) Debugf(format string, v ...interface{}) {
	if !l.debug {
		return
	}
	msg := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("15:04:05.000")
	if l.colors {
		l.debugColor.Printf("%s %s [DEBUG] %s\n", IconDebug, timestamp, msg)
	} else {
		fmt.Fprintf(l.output, "%s [DEBUG] %s\n", timestamp, msg)
	}
}

// Verbosef prints formatted verbose message when verbose enabled
func (l *Logger) Verbosef(format string, v ...interface{}) {
	if !l.verbose {
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	msg := fmt.Sprintf(format, v...)
	if l.colors {
		l.verboseColor.Printf("%s [INFO] %s\n", IconVerbose, msg)
	} else {
		fmt.Fprintf(l.output, "%s [INFO] %s\n", IconVerbose, msg)
	}
}

// Errorf prints error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	msg := fmt.Sprintf(format, v...)
	writer := l.output

	if l.output == os.Stdout {
		writer = os.Stderr
	}

	if l.colors {
		l.errorColor.Fprintf(writer, "%s [ERROR] %s\n", IconError, msg)
	} else {
		fmt.Fprintf(writer, "[ERROR] %s\n", msg)
	}
}

// Successf prints success message
func (l *Logger) Successf(format string, v ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	msg := fmt.Sprintf(format, v...)
	if l.colors {
		l.successColor.Printf("%s [SUCCESS] %s\n", IconSuccess, msg)
	} else {
		fmt.Fprintf(l.output, "%s [SUCCESS] %s\n", IconSuccess, msg)
	}
}

// Warnf prints warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	msg := fmt.Sprintf(format, v...)
	if l.colors {
		l.warnColor.Printf("%s [WARN] %s\n", IconWarning, msg)
	} else {
		fmt.Fprintf(l.output, "%s [WARN] %s\n", IconWarning, msg)
	}
}

// Printf prints normal formatted message
func (l *Logger) Printf(format string, v ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	fmt.Fprintf(l.output, format, v...)
}

// PrintFileLocation formats file paths with optional line numbers
func (l *Logger) PrintFileLocation(file string, line int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.colors {
		l.fileColor.Print(file)
		if line > 0 {
			l.lineColor.Printf(":%d", line)
		}
	} else {
		if line > 0 {
			fmt.Fprintf(l.output, "%s:%d", file, line)
		} else {
			fmt.Fprint(l.output, file)
		}
	}
	fmt.Fprintln(l.output)
}

// PrintTestFailure formats a test failure with colors
func (l *Logger) PrintTestFailure(testName, errMsg, file string, line int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.colors {
		l.errorColor.Print("FAIL ")
		fmt.Fprint(l.output, testName)
		l.errorColor.Print(": ")
		fmt.Fprint(l.output, errMsg)
		fmt.Fprint(l.output, " (")
		l.fileColor.Print(file)
		if line > 0 {
			l.lineColor.Printf(":%d", line)
		}
		fmt.Fprintln(l.output, ")")
	} else {
		fmt.Fprintf(l.output, "FAIL %s: %s (%s", testName, errMsg, file)
		if line > 0 {
			fmt.Fprintf(l.output, ":%d", line)
		}
		fmt.Fprintln(l.output, ")")
	}
}
