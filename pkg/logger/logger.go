package logger

import (
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2/klogr"
)

type Logger interface {
	// Debugf logs a formatted debugging message.
	Debugf(format string, args ...interface{})

	// Infof logs a formatted informational message.
	Infof(format string, args ...interface{})

	// Warnf logs a formatted warning message.
	Warnf(format string, args ...interface{})

	// Errorf logs a formatted error message.
	Errorf(format string, args ...interface{})

	// Error logs an error
	Error(err error, format string, args ...interface{})

	// WithValues annotate klog
	WithValues(withKeysAndValues ...interface{}) Logger

	GetLogger() logr.Logger
}

type logger struct {
	name string
	logr logr.Logger
}

// New creates a new logger.
func New(name string) Logger {
	return &logger{
		name: name,
		logr: klogr.NewWithOptions(
			klogr.WithFormat(klogr.FormatKlog),
		).WithName(name),
	}
}

func (l *logger) WithValues(withKeysAndValues ...interface{}) Logger {
	return &logger{
		name: l.name,
		logr: l.logr.WithValues(withKeysAndValues...),
	}
}

func (l *logger) GetLogger() logr.Logger {
	return l.logr
}

func sanitize(in []interface{}) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, v := range in {
		// We have to check if the type is hashable; otherwise we risk a panic at runtime
		kind := reflect.TypeOf(v).Kind()
		isHashable := kind < reflect.Array || kind == reflect.Ptr || kind == reflect.UnsafePointer
		if isHashable {
			out = append(out, fmt.Sprintf("%v", v))
		} else {
			out = append(out, "__non_hashable_type__")
		}
	}
	return out
}

// Verbosity levels following https://kubernetes.io/docs/concepts/cluster-administration/system-logs/
func (l *logger) Debugf(format string, args ...interface{}) {
	l.logr.V(4).Info(format, sanitize(args)...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.logr.V(2).Info(format, sanitize(args)...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.logr.V(1).Info(format, sanitize(args))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.logr.V(0).Info(format, sanitize(args)...)
}

func (l *logger) Error(err error, format string, args ...interface{}) {
	l.logr.V(0).Error(err, format, sanitize(args)...)
}
