package errors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

func IsTimeoutError(err error) bool {
	switch {
	case
		IsOpError(
			err,
			StatusTimeout,
			StatusCancelled,
		),
		IsTransportError(err, TransportErrorCanceled, TransportErrorDeadlineExceeded),
		Is(
			err,
			context.DeadlineExceeded,
			context.Canceled,
		):
		return true
	default:
		return false
	}
}

func ErrIf(cond bool, err error) error {
	if cond {
		return err
	}
	return nil
}

func HideEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// As is a proxy to errors.As
// This need to single import errors
func As(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	return errors.As(err, target)
}

// Is is a improved proxy to errors.Is
// This need to single import errors
func Is(err error, targets ...error) bool {
	if len(targets) == 0 {
		panic("empty targets")
	}
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// New is a proxy to errors.New
// This need to single import errors
func New(text string) error {
	return ErrorfSkip(1, "%w", errors.New(text))
}

// NewWithIssues returns error which contains child issues
func NewWithIssues(text string, issues ...error) error {
	return ErrorfSkip(1, "%w", &errorWithIssues{
		reason: text,
		issues: issues,
	})
}

type errorWithIssues struct {
	reason string
	issues []error
}

func (e *errorWithIssues) Error() string {
	var b bytes.Buffer
	b.WriteString(e.reason)
	b.WriteString(", issues: [")
	for i, issue := range e.issues {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(issue.Error())
	}
	b.WriteString("]")
	return b.String()
}

// Error is alias to fmt.Errorf() with prepend file:line prefix
func Error(err error) error {
	return &stackError{
		stackRecord: stackRecord(1),
		err:         err,
	}
}

// Errorf is alias to fmt.Errorf() with prepend file:line prefix
func Errorf(format string, args ...interface{}) error {
	return ErrorfSkip(1, format, args...)
}

// ErrorfSkip is alias to fmt.Errorf() with prepend file:line prefix
func ErrorfSkip(depth int, format string, args ...interface{}) error {
	return &stackError{
		stackRecord: stackRecord(depth + 1),
		err:         fmt.Errorf(format, args...),
	}
}

func stackRecord(depth int) string {
	function, file, line, _ := runtime.Caller(depth + 1)
	name := runtime.FuncForPC(function).Name()
	return funcName(name) + "(" + packageName(name) + "/" + fileName(file) + ":" + strconv.Itoa(line) + ")"
}

func fileName(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	}
	return original[i+1:]
}

func packageName(original string) string {
	i := strings.LastIndex(original, ".")
	if i == -1 {
		return original
	}
	return original[:i]
}

func funcName(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	}
	return original[i+1:]
}

type stackError struct {
	stackRecord string
	err         error
}

func (e *stackError) Error() string {
	return e.err.Error() + " at `" + e.stackRecord + "`"
}

func (e *stackError) Unwrap() error {
	return e.err
}
