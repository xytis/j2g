package main

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"strings"
)

type textFormatter struct {
}

// Based off logrus.TextFormatter, which behaves completely
// differently when you don't want colored output
func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	levelText := strings.ToUpper(entry.Level.String())[0:4]
	timeStamp := entry.Time.Format("2006/01/02 15:04:05.000000")
	if len(entry.Data) > 0 {
		fmt.Fprintf(b, "%s: %s %-44s ", levelText, timeStamp, entry.Message)
		for k, v := range entry.Data {
			fmt.Fprintf(b, " %s=%v", k, v)
		}
	} else {
		// No padding when there's no fields
		fmt.Fprintf(b, "%s: %s %s", levelText, timeStamp, entry.Message)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

var (
	standardTextFormatter = &textFormatter{}
)

var (
	Log *logrus.Logger
)

func init() {
	Log = logrus.New()
	Log.Formatter = standardTextFormatter
}

func SetLogLevel(levelname string) {
	level, err := logrus.ParseLevel(levelname)
	if err != nil {
		Log.Fatal(err)
	}
	Log.Level = level
}

func CheckFatal(e error) {
	if e != nil {
		Log.Fatal(e)
	}
}

func CheckWarn(e error) {
	if e != nil {
		Log.Warnln(e)
	}
}
