// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package logging

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
)

// Writer implements LogSink interface using zerolog.Logger

type Writer struct {
	Log       *zerolog.Logger
	Verbosity zerolog.Level
	Caller    string
}

// ignoring CallDepth
func (w *Writer) Init(ri logr.RuntimeInfo) {
}

func (w *Writer) Enabled(lvl int) bool {
	return lvl >= int(w.Verbosity)
}

func (w *Writer) WithValues(keysAndValues ...interface{}) logr.LogSink {
	values, err := handleFields(keysAndValues)
	if err != nil {
		w.Log.Error().Msg(err.Error())
		return w
	}
	l := w.Log.With()
	for key, val := range values {
		l = l.Str(key, val)
	}
	newLog := l.Logger()
	return &Writer{Log: &newLog, Verbosity: w.Verbosity, Caller: w.Caller}
}

func (w *Writer) Info(lvl int, msg string, keysAndVals ...interface{}) {
	if !w.Enabled(lvl) {
		return
	}
	values, err := handleFields(keysAndVals)
	if err != nil {
		w.Log.Error().Msg(err.Error())
	}
	l := w.Log.Info()
	if w.Caller != "" {
		l = l.Str(SETUP, w.Caller)
	}
	for key, val := range values {
		l = l.Str(key, val)
	}
	l.Msg(msg)
}

func (w *Writer) Error(err error, msg string, keysAndVals ...interface{}) {
	values, inputErr := handleFields(keysAndVals)
	if inputErr != nil {
		w.Log.Error().Msg(inputErr.Error())
	}
	l := w.Log.Error().Err(err)
	if w.Caller != "" {
		l = l.Str(SETUP, w.Caller)
	}
	for key, val := range values {
		l = l.Str(key, val)
	}
	l.Msg(msg)
}

func (w *Writer) WithName(name string) logr.LogSink {
	newLogger := *w
	if w.Caller == "" {
		newLogger.Caller = name
	} else {
		newLogger.Caller += "." + name
	}
	return &newLogger
}

func (w *Writer) WithCallDepth(depth int) logr.LogSink {
	return w
}

func handleFields(args []interface{}) (map[string]string, error) {
	if len(args)%2 == 1 {
		return nil, errors.New("odd number of arguments passed as key-value pairs to the logger")
	}
	res := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		keyStr, isString := args[i].(string)
		if !isString {
			return nil, errors.New("non-string type passed as a key to the logger")
		}
		valStr := fmt.Sprint(args[i+1])
		res[keyStr] = valStr
	}
	return res, nil
}
