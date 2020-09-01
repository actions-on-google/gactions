// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package log provides common log functions to the cli infrastructure.
package log

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
)

// Level defines the supported log levels.
type Level uint8

// Levels are the currently defined log levels for the CLI.
const (
	DebugLevel Level = iota // DebugLevel will be set if the debug version of the binary gets built.
	InfoLevel               // InfoLevel will be set if developer passes a verbose flag.
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

var (
	// DebugLogger will reveal debug info which can be internal; will not be part of public binary
	DebugLogger = log.New(os.Stdout, colorMaybe("[DEBUG] ", color.HiBlueString), log.Ldate|log.Ltime|log.Llongfile)
	// InfoLogger sends useful but verbose information. Only sends if severity is >= InfoLevel.
	InfoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime)
	// OutLogger sends an important output from execution of the command, intended for a user to read.
	OutLogger = log.New(os.Stdout, "", 0)
	// WarnLogger sends warnings to stderr.
	WarnLogger = log.New(os.Stderr, colorMaybe("[WARNING] ", color.YellowString), 0)
	// ErrorLogger sends errors to stderr.
	ErrorLogger = log.New(os.Stderr, colorMaybe("[ERROR] ", color.RedString), 0)
	// Severity can be set to restrict level of log messages.
	Severity = WarnLevel
)

func colorMaybe(s string, f func(format string, a ...interface{}) string) string {
	if runtime.GOOS == "windows" {
		return s
	}
	return f(s)
}

// DoneMsgln surrounds msg with helpful visual cues for the user to indicate completion of a task.
func DoneMsgln(msg string) {
	// Windows doesn't print special characters and colors nicely.
	if runtime.GOOS == "windows" {
		Outf("Done. %s\n", msg)
		return
	}
	Outf("%v Done. %s\n", color.GreenString("✔"), msg)
}

// Debugf calls Output to print to the DebugLogger.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	if Severity > DebugLevel {
		return
	}
	DebugLogger.Output(2, fmt.Sprintf(format, v...))
}

// Debugln calls Output to print to the DebugLogger.
// Arguments are handled in the manner of fmt.Println.
func Debugln(v ...interface{}) {
	if Severity > DebugLevel {
		return
	}
	DebugLogger.Output(2, fmt.Sprintln(v...))
}

// Out calls Output to print to the OutLogger.
// Arguments are handled in the manner of fmt.Print.
func Out(v ...interface{}) {
	OutLogger.Output(2, fmt.Sprint(v...))
}

// Outf calls Output to print to the OutLogger.
// Arguments are handled in the manner of fmt.Printf.
func Outf(format string, v ...interface{}) {
	OutLogger.Output(2, fmt.Sprintf(format, v...))
}

// Outln calls Output to print to the OutLogger.
// Arguments are handled in the manner of fmt.Println.
func Outln(v ...interface{}) {
	OutLogger.Output(2, fmt.Sprintln(v...))
}

// Infoln calls Output to print to the InfoLogger.
// Arguments are handled in the manner of fmt.Println.
func Infoln(v ...interface{}) {
	if Severity > InfoLevel {
		return
	}
	InfoLogger.Output(2, fmt.Sprintln(v...))
}

// Infof calls Output to print to the InfoLogger.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	if Severity > InfoLevel {
		return
	}
	InfoLogger.Output(2, fmt.Sprintf(format, v...))
}

// Error calls Output to print to the ErrorLogger.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	if Severity > ErrorLevel {
		return
	}
	ErrorLogger.Output(2, fmt.Sprint(v...))
}

// Errorf calls Output to print to the ErrorLogger.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	if Severity > ErrorLevel {
		return
	}
	ErrorLogger.Output(2, fmt.Sprintf(format, v...))
}

// Warnf calls Output to print to the WarnLogger.
// Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, v ...interface{}) {
	if Severity > WarnLevel {
		return
	}
	WarnLogger.Output(2, fmt.Sprintf(format, v...))
}

// Warnln calls Output to print to the WarnLogger.
// Arguments are handled in the manner of fmt.Println.
func Warnln(v ...interface{}) {
	if Severity > WarnLevel {
		return
	}
	WarnLogger.Output(2, fmt.Sprintln(v...))
}
