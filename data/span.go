package data

import "fmt"

type Pos struct {
	Line int
	Col  int
}

type Span struct {
	Start Pos
	End   Pos
}

func NewSpan(s1 Span, s2 Span) Span {
	return Span{s1.Start, s2.End}
}

func NewSpan2(ls int, cs int, le int, ce int) Span {
	return Span{Pos{ls, cs}, Pos{le, ce}}
}

func (s Span) String() string {
	return fmt.Sprintf("%d:%d - %d:%d", s.Start.Line, s.Start.Col, s.End.Line, s.End.Col)
}

// Returns true if there's no lines between these spans
func (s Span) Adjacent(other Span) bool {
	return s.End.Line+1 == other.Start.Line
}

// Returns true if this ends on the same line as other starts
func (s Span) SameLine(other Span) bool {
	return s.End.Line == other.Start.Line
}

func (s Span) IsEmpty() bool {
	return s.Start.Line == 0 && s.Start.Col == 0 && s.End.Line == 0 && s.End.Col == 0
}
