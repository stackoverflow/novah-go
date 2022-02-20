package ast

import (
	"fmt"
	"unicode"

	"github.com/huandu/go-clone"
	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type SType interface {
	sType()
	GetSpan() lexer.Span
	Clone() SType
	WithSpan(lexer.Span) SType
	fmt.Stringer
}

type STConst struct {
	Name  string
	Alias *string
	Span  lexer.Span
}

type STApp struct {
	Type  SType
	Types []SType
	Span  lexer.Span
}

type STFun struct {
	Arg  SType
	Ret  SType
	Span lexer.Span
}

type STParens struct {
	Type SType
	Span lexer.Span
}

type STRecord struct {
	Row  SType
	Span lexer.Span
}

type STRowEmpty struct {
	Span lexer.Span
}

type STRowExtend struct {
	Labels data.LabelMap[SType]
	Row    SType
	Span   lexer.Span
}

type STImplicit struct {
	Type SType
	Span lexer.Span
}

func (t STConst) sType() {}
func (t STConst) GetSpan() lexer.Span {
	return t.Span
}
func (t STConst) String() string {
	return t.Name
}
func (t STConst) Fullname() string {
	if t.Alias != nil {
		return fmt.Sprintf("%s.%s", *t.Alias, t.Name)
	}
	return t.Name
}
func (t STConst) Clone() SType {
	return clone.Clone(t).(STConst)
}
func (t STConst) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func (t STApp) sType() {}
func (t STApp) GetSpan() lexer.Span {
	return t.Span
}
func (t STApp) String() string {
	sname := t.Type.String()
	if len(t.Types) == 0 {
		return sname
	}
	return fmt.Sprintf("%s %s", sname, data.JoinToString(t.Types, " "))
}
func (t STApp) Clone() SType {
	return clone.Clone(t).(STApp)
}
func (t STApp) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func (_ STFun) sType() {}
func (t STFun) GetSpan() lexer.Span {
	return t.Span
}
func (t STFun) String() string {
	return fmt.Sprintf("%s -> %s", t.Arg.String(), t.Ret.String())
}
func (t STFun) Clone() SType {
	return clone.Clone(t).(STFun)
}
func (t STFun) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func (_ STParens) sType() {}
func (t STParens) GetSpan() lexer.Span {
	return t.Span
}
func (t STParens) String() string {
	return fmt.Sprintf("(%s)", t.Type.String())
}
func (t STParens) Clone() SType {
	return clone.Clone(t).(STParens)
}
func (t STParens) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func (_ STImplicit) sType() {}
func (t STImplicit) GetSpan() lexer.Span {
	return t.Span
}
func (t STImplicit) String() string {
	return fmt.Sprintf("{{ %s }}", t.Type.String())
}
func (t STImplicit) Clone() SType {
	return clone.Clone(t).(STImplicit)
}
func (t STImplicit) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func (_ STRowEmpty) sType() {}
func (t STRowEmpty) GetSpan() lexer.Span {
	return t.Span
}
func (t STRowEmpty) String() string {
	return "[]"
}
func (t STRowEmpty) Clone() SType {
	return clone.Clone(t).(STRowEmpty)
}
func (t STRowEmpty) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func (_ STRowExtend) sType() {}
func (t STRowExtend) GetSpan() lexer.Span {
	return t.Span
}
func (t STRowExtend) String() string {
	labels := t.Labels.Show(func(k string, v SType) string {
		return fmt.Sprintf("%s : %s", k, v.String())
	})
	return fmt.Sprintf("[ %s ]", labels)
}
func (t STRowExtend) Clone() SType {
	return clone.Clone(t).(STRowExtend)
}
func (t STRowExtend) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func (_ STRecord) sType() {}
func (t STRecord) GetSpan() lexer.Span {
	return t.Span
}
func (t STRecord) String() string {
	switch t.Row.(type) {
	case STRowEmpty:
		return "{}"
	case STRowExtend:
		{
			rows := t.Row.String()
			return fmt.Sprintf("{%s}", rows[1:len(rows)-1])
		}
	default:
		return fmt.Sprintf("{ | %s }", t.Row.String())
	}
}
func (t STRecord) Clone() SType {
	return clone.Clone(t).(STRecord)
}
func (t STRecord) WithSpan(span lexer.Span) SType {
	t.Span = span
	return t
}

func FindFreeVars(ty SType, bound []string) []string {
	res := make([]string, 0, 2)
	toCheck := []SType{ty}

	for len(toCheck) > 0 {
		t := toCheck[0]
		switch v := t.(type) {
		case STConst:
			{
				ch := rune(v.Name[0])
				if unicode.IsUpper(ch) && data.InSlice(bound, v.Name) {
					res = append(res, v.Name)
				}
			}
		case STFun:
			toCheck = append(toCheck, v.Arg, v.Ret)
		case STApp:
			{
				toCheck = append(toCheck, v.Type)
				toCheck = append(toCheck, v.Types...)
			}
		case STParens:
			toCheck = append(toCheck, v.Type)
		case STRecord:
			toCheck = append(toCheck, v.Row)
		case STRowExtend:
			{
				toCheck = append(toCheck, v.Row)
				toCheck = append(toCheck, v.Labels.Values()...)
			}
		case STImplicit:
			toCheck = append(toCheck, v.Type)
		}
		toCheck = toCheck[1:]
	}
	return res
}

func Everywhere(ty SType, f func(SType) SType) SType {
	var run func(typ SType) SType
	run = func(typ SType) SType {
		switch t := typ.(type) {
		case STConst:
			return f(t)
		case STApp:
			{
				app := t.Clone().(STApp)
				app.Type = run(app.Type)
				app.Types = data.MapSlice(app.Types, run)
				return app
			}
		case STFun:
			{
				fun := t.Clone().(STFun)
				fun.Arg = run(fun.Arg)
				fun.Ret = run(fun.Ret)
				return fun
			}
		case STParens:
			{
				par := t.Clone().(STParens)
				par.Type = run(par.Type)
				return par
			}
		case STRecord:
			{
				par := t.Clone().(STRecord)
				par.Row = run(par.Row)
				return par
			}
		case STRowEmpty:
			return f(t)
		case STRowExtend:
			{
				rec := t.Clone().(STRowExtend)
				rec.Row = run(rec.Row)
				rec.Labels = data.LabelMapValues(rec.Labels, run)
				return rec
			}
		case STImplicit:
			{
				imp := t.Clone().(STImplicit)
				imp.Type = run(imp.Type)
				return imp
			}
		default:
			panic("unknow type in everywhere: " + t.String())
		}
	}
	return run(ty)
}

func SubstVar(t SType, from string, new SType) SType {
	return Everywhere(t, func(s SType) SType {
		switch ty := s.(type) {
		case STConst:
			if ty.Name == from {
				return new
			} else {
				return ty
			}
		default:
			return ty
		}
	})
}
