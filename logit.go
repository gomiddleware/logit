package logit

import (
	"io"
	"sync"
	"time"
)

// A Logger represents an active logging object that generates lines of output to an io.Writer. Each logging operation
// makes a single call to the Writer's Write method. A Logger can be used simultaneously from multiple goroutines; it
// guarantees to serialize access to the Writer.
type Logger struct {
	mu     *sync.Mutex            // ensures atomic writes; protects the following fields
	out    io.Writer              // destination for output
	sys    string                 // the sub-system to write at beginning of each line
	fields map[string]interface{} // the fields to also log
}

func New(out io.Writer, sys string) *Logger {
	mu := sync.Mutex{}
	return &Logger{mu: &mu, out: out, sys: sys, fields: make(map[string]interface{})}
}

// Clone returns a new Logger which uses the same output but which has the system set to "<current>.system". e.g. if
// you have a logger where has the "main" system, cloning it with the system "datastore" will result in the system
// field being "main.datastore".
//
// Since the cloned logger uses the same lock as the original, use of this lock also garantees serial access to the
// Writer and therefore can also be used in multiple goroutines.
func (l *Logger) Clone(sys string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	// clone the l.fields here
	m := make(map[string]interface{})
	for k, v := range l.fields {
		m[k] = v
	}

	return &Logger{mu: l.mu, out: l.out, sys: l.sys + "." + sys, fields: m}
}

func (l *Logger) WithField(key string, value interface{}) {
	if key == "time" {
		panic("logit: key=time is not allowed")
	}
	if key == "sys" {
		panic("logit: key=sys is not allowed")
	}
	if key == "msg" {
		panic("logit: key=msg is not allowed")
	}
	if key == "err" {
		panic("logit: key=err is not allowed")
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.fields[key] = value
}

// Log just logs a message to the output. It doesn't do anything special.
func (l *Logger) Log(msg string) error {
	return l.Output(msg)
}

// Output writes the output for a logging event.
func (l *Logger) Output(msg string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// time
	str := "time="
	str += time.Now().UTC().Format("20060102-150405.000000000")
	str += " "

	// sys
	str += "sys=" + l.sys + " "

	// now do all of the fields
	for k, v := range l.fields {
		switch vv := v.(type) {
		case string:
			// ToDo: currently presuming everything is a string
			str += k + "=" + vv + " "
		default:
			str += k + "=" + "(unknown type)"
		}
	}

	// message
	str += "msg=" + msg

	// newline
	str += "\n"

	_, err := l.out.Write([]byte(str))
	return err
}

// // SetOutput sets the output destination for the logger.
// func (l *Logger) SetOutput(w io.Writer) {
// 	l.mu.Lock()
// 	defer l.mu.Unlock()
// 	l.out = w
// }
