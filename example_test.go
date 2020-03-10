package log_test

import (
	"github.com/cobaltspeech/log"
	"github.com/cobaltspeech/log/pkg/level"
)

func Example() {
	// Create a leveled logger
	l := log.NewLeveledLogger()

	// Provide the logger to library functions
	Divide(l, 5, 0)

	// Change the logging level at runtime
	l.SetFilterLevel(level.Debug | level.Info | level.Error)

	// Provide the logger to constructors that support the Logger interface
	e := NewEngine(l)
	e.Run()
}

// Divide supports the Logger interface and uses it to report events when
// performing the division of given arguments.  It uses the DiscardLogger if a
// valid logger was not provided.  Library functions can use such a signature to
// support the logger.
func Divide(l log.Logger, a, b int) int {
	if l == nil {
		l = log.NewDiscardLogger()
	}

	l.Trace("msg", "entering Divide()")
	defer l.Trace("msg", "exiting Divide()")

	if b == 0 {
		l.Error(
			"msg", "attempt to divide by zero",
			"a", a,
			"b", b)

		return 0
	}

	return a / b
}

// Engine is an example type that supports the logging interface for use it its methods.
type Engine struct {
	log log.Logger
}

// NewEngine returns an initialized Engine configured with the provided logger.
// If a nil logger is provided, it uses the DiscardLogger.
func NewEngine(l log.Logger) *Engine {
	if l == nil {
		l = log.NewDiscardLogger()
	}

	return &Engine{l}
}

// Run uses the configured logger to report events.
func (e *Engine) Run() {
	e.log.Debug("msg", "running engine")
}
