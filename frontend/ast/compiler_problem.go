package ast

import "github.com/stackoverflow/novah-go/frontend/lexer"

type Severity = int

const (
	WARN Severity = iota
	ERROR
	FATAL
)

type CompilerProblem struct {
	Msg      string
	Span     lexer.Span
	Filename string
	Module   *string
	Severity Severity
}

func (cp CompilerProblem) Error() string {
	return cp.Msg
}
