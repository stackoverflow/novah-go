package typechecker

import (
	"github.com/huandu/go-clone"
	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type KindType = int

const (
	STAR KindType = iota
	CTOR
)

type Kind struct {
	Type  KindType
	Arity int
}

func (k Kind) String() string {
	if k.Type == STAR {
		return "Type"
	}
	return data.JoinToStringFunc(data.Range(0, k.Arity), " -> ", func(_ int) string { return "Type" })
}

type TypeVarTag = int

const (
	UNBOUND TypeVarTag = iota
	LINK
	GENERIC
)

type TypeVar struct {
	Tag   TypeVarTag
	Id    int
	Level int
	Type  int
}

type Type interface {
	sType()
	Clone() Type
	GetSpan() lexer.Span
	WithSpan(lexer.Span) Type
}

type TConst struct {
	Name string
	Kind KindType
	Span lexer.Span
}

type TApp struct {
	Type  Type
	Types []Type
	Span  lexer.Span
}

type TArrow struct {
	Args []Type
	Ret  Type
	Span lexer.Span
}

type TImplicit struct {
	Type Type
	Span lexer.Span
}

type TRecord struct {
	Row  Type
	Span lexer.Span
}

type TRowEmpty struct {
	Span lexer.Span
}

type TRowExtend struct {
	Labels data.LabelMap[Type]
	Row    Type
	Span   lexer.Span
}

type TVar struct {
	Tvar TypeVar
	Span lexer.Span
}

func (_ TConst) sType()     {}
func (_ TApp) sType()       {}
func (_ TArrow) sType()     {}
func (_ TImplicit) sType()  {}
func (_ TRecord) sType()    {}
func (_ TRowEmpty) sType()  {}
func (_ TRowExtend) sType() {}
func (_ TVar) sType()       {}
func (t TConst) GetSpan() lexer.Span {
	return t.Span
}
func (t TApp) GetSpan() lexer.Span {
	return t.Span
}
func (t TArrow) GetSpan() lexer.Span {
	return t.Span
}
func (t TImplicit) GetSpan() lexer.Span {
	return t.Span
}
func (t TRecord) GetSpan() lexer.Span {
	return t.Span
}
func (t TRowEmpty) GetSpan() lexer.Span {
	return t.Span
}
func (t TRowExtend) GetSpan() lexer.Span {
	return t.Span
}
func (t TVar) GetSpan() lexer.Span {
	return t.Span
}
func (t TConst) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}
func (t TApp) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}
func (t TArrow) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}
func (t TImplicit) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}
func (t TRecord) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}
func (t TRowEmpty) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}
func (t TRowExtend) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}
func (t TVar) WithSpan(span lexer.Span) Type {
	t.Span = span
	return t
}

func (t TConst) Clone() Type {
	return clone.Clone(t).(TConst)
}
func (t TApp) Clone() Type {
	return clone.Clone(t).(TApp)
}
func (t TArrow) Clone() Type {
	return clone.Clone(t).(TArrow)
}
func (t TImplicit) Clone() Type {
	return clone.Clone(t).(TImplicit)
}
func (t TRecord) Clone() Type {
	return clone.Clone(t).(TRecord)
}
func (t TRowEmpty) Clone() Type {
	return clone.Clone(t).(TRowEmpty)
}
func (t TRowExtend) Clone() Type {
	return clone.Clone(t).(TRowExtend)
}
func (t TVar) Clone() Type {
	return clone.Clone(t).(TVar)
}
