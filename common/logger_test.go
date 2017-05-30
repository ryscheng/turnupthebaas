package common

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	l := NewLogger("test")
	l.Disable()
	l.Enable()
	SilenceLoggers()
}
