package data

import (
	"fmt"
)

type Severity = int

const (
	WARN Severity = iota
	ERROR
	FATAL
)

const (
	red    = "\u001b[31m"
	yellow = "\u001b[33m"
	reset  = "\u001b[0m"
)

type CompilerProblem struct {
	Msg           string
	Span          Span
	Filename      string
	Module        string
	Severity      Severity
	TypingContext string
}

func (cp CompilerProblem) Error() string {
	return fmt.Sprintf("%s at %s", cp.Msg, cp.Span.String())
}

func (err CompilerProblem) FormatToConsole() string {
	var mod string
	if err.Module != "" {
		mod = fmt.Sprintf("module %s%s%s", yellow, err.Module, reset)
	}
	at := fmt.Sprintf("at %s:%s\n\n", err.Filename, err.Span.String())

	return fmt.Sprintf("%s%s%s\n\n%s", mod, at, PrependIdent(err.Msg, "  "), err.TypingContext)
}
