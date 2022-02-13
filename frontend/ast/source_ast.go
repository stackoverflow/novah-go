// This file contains the initial parsed AST before
// desugaring and type checking
package ast

import (
	"fmt"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type Spanned[T any] struct {
	Val  T
	Span lexer.Span
}

type Visibility = int

const (
	PUBLIC Visibility = iota
	PRIVATE
)

type SRefType = int

const (
	VAR SRefType = iota
	TYPE
)

// Represents a source module AST
type SModule struct {
	Name       Spanned[string]
	SourceName string
	Imports    []SImport
	Foreigns   []SForeignImport
}

// `null` ctors mean only the type is imported, but no constructors.
// Empty ctors mean all constructors are imported.
type SDeclarationRef struct {
	Tag   SRefType
	Name  Spanned[string]
	Span  lexer.Span
	Ctors *[]Spanned[string]
}

type SImport struct {
	Module  Spanned[string]
	Span    lexer.Span
	Alias   *string
	Auto    bool
	Comment *lexer.Comment
	Defs    *[]SDeclarationRef
}

type SForeignImport struct {
	Type  string
	Alias *string
	Span  lexer.Span
}

///////////////////////////////////////////////
// Source Expressions
///////////////////////////////////////////////

type SExpr interface {
	fmt.Stringer
	sExpr()
	Span() lexer.Span
	Comment() *lexer.Comment
}

type SInt struct {
	V       int64
	Text    string
	span    lexer.Span
	comment *lexer.Comment
}

type SFloat struct {
	V       float64
	Text    string
	span    lexer.Span
	comment *lexer.Comment
}

type SComplex struct {
	V       complex128
	Text    string
	span    lexer.Span
	comment *lexer.Comment
}

type SString struct {
	V       string
	Raw     string
	span    lexer.Span
	comment *lexer.Comment
}

type SChar struct {
	V       rune
	Raw     string
	span    lexer.Span
	comment *lexer.Comment
}

type SBool struct {
	V       bool
	span    lexer.Span
	comment *lexer.Comment
}

type SVar struct {
	Name    string
	Alias   *string
	span    lexer.Span
	comment *lexer.Comment
}

type SOperator struct {
	Name     string
	Alias    *string
	IsPrefix bool
	span     lexer.Span
	comment  *lexer.Comment
}

type SImplicitVar struct {
	Name    string
	Alias   *string
	span    lexer.Span
	comment *lexer.Comment
}

type SConstructor struct {
	Name    string
	Alias   *string
	span    lexer.Span
	comment *lexer.Comment
}

type SPatternLiteral struct {
	Regex   string
	Raw     string
	span    lexer.Span
	comment *lexer.Comment
}

type SLambda struct {
	Pats    []SPattern
	Body    SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SApp struct {
	Fn      SExpr
	Arg     SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SBinApp struct {
	Op      SExpr
	Left    SExpr
	Right   SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SIf struct {
	Cond    SExpr
	Then    SExpr
	Else    *SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SLet struct {
	Def     SLetDef
	Body    SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SMatch struct {
	Exprs   []SExpr
	Cases   []SCase
	span    lexer.Span
	comment *lexer.Comment
}

type SAnn struct {
	Exp     SExpr
	Type    SType
	span    lexer.Span
	comment *lexer.Comment
}

type SDo struct {
	Exps    []SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SDoLet struct {
	Def     SLetDef
	span    lexer.Span
	comment *lexer.Comment
}

type SLetBang struct {
	Def     SLetDef
	Body    *SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SFor struct {
	Def     SLetDef
	Body    SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SParens struct {
	Exp     SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SUnit struct {
	span    lexer.Span
	comment *lexer.Comment
}

type SRecordEmpty struct {
	span    lexer.Span
	comment *lexer.Comment
}

type SRecordSelect struct {
	Exp     SExpr
	Labels  []Spanned[string]
	span    lexer.Span
	comment *lexer.Comment
}

type SRecordExtend struct {
	Labels  LabelMap[SExpr]
	Exp     SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SRecordRestrict struct {
	Exp     SExpr
	Labels  []string
	span    lexer.Span
	comment *lexer.Comment
}

type SRecordUpdate struct {
	Exp     SExpr
	Labels  []Spanned[string]
	Val     SExpr
	IsSet   bool
	span    lexer.Span
	comment *lexer.Comment
}

type SRecordMerge struct {
	Exp1    SExpr
	Exp2    SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SListLiteral struct {
	Exps    []SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SSetLiteral struct {
	Exps    []SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SIndex struct {
	Exp     SExpr
	Index   SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SUnderscore struct {
	span    lexer.Span
	comment *lexer.Comment
}

type SWhile struct {
	Cond    SExpr
	Exps    []SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SComputation struct {
	Builder SVar
	Exps    []SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SReturn struct {
	Exp     SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SYield struct {
	Exp     SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SDoBang struct {
	Exp     SExpr
	span    lexer.Span
	comment *lexer.Comment
}

type SNil struct {
	span    lexer.Span
	comment *lexer.Comment
}

type STypeCast struct {
	Exp     SExpr
	Cast    SType
	span    lexer.Span
	comment *lexer.Comment
}

func (_ SVar) sExpr() {}
func (e SVar) Span() lexer.Span {
	return e.span
}
func (e SVar) Comment() *lexer.Comment {
	return e.comment
}
func (e SVar) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

///////////////////////////////////////////////
// cases in pattern matching
///////////////////////////////////////////////

type SCase struct {
	Pats  []SPattern
	Exp   SExpr
	Guard *SExpr
}

///////////////////////////////////////////////
// let definitions
///////////////////////////////////////////////

type SLetDef interface {
	Expr() SExpr
}

type SLetBind struct {
	expr       SExpr
	Name       Spanned[string]
	Pats       []SPattern
	IsInstance bool
	Type       *SType
}

type SLetPat struct {
	expr SExpr
	Pat  SPattern
}

func (def SLetBind) Expr() SExpr {
	return def.expr
}

func (def SLetPat) Expr() SExpr {
	return def.expr
}

///////////////////////////////////////////////
// patterns for pattern matching
///////////////////////////////////////////////

type SPattern interface {
	fmt.Stringer
	sPattern()
	Span() lexer.Span
}

type SWildcard struct {
	span lexer.Span
}

type SLiteralP struct {
	Lit  SExpr
	span lexer.Span
}

type SVarP struct {
	V SVar
}

type SCtorP struct {
	Ctor   SConstructor
	Fields []SPattern
	span   lexer.Span
}

type SRecordP struct {
	Labels LabelMap[SPattern]
	span   lexer.Span
}

type SListP struct {
	Elems []SPattern
	Tail  *SPattern
	span  lexer.Span
}

type SNamed struct {
	Pat  SPattern
	Name Spanned[string]
	span lexer.Span
}

type SUnitP struct {
	span lexer.Span
}

type STypeTest struct {
	Type  SType
	Alias *string
	span  lexer.Span
}

type SImplicitP struct {
	Pat  SPattern
	span lexer.Span
}

type STupleP struct {
	P1   SPattern
	P2   SPattern
	span lexer.Span
}

type SRegexP struct {
	Regex SPatternLiteral
}

// will be desugared in the desugar phase
type SParensP struct {
	Pat  SPattern
	span lexer.Span
}

// will be desugared in the desugar phase
type STypeAnnotationP struct {
	Par  SVar
	Type SType
	span lexer.Span
}

func (_ SWildcard) sPattern() {}
func (p SWildcard) Span() lexer.Span {
	return p.span
}
func (p SWildcard) String() string {
	return "_"
}

func (_ SLiteralP) sPattern() {}
func (p SLiteralP) Span() lexer.Span {
	return p.span
}
func (p SLiteralP) String() string {
	return p.Lit.String()
}

func (_ SVarP) sPattern() {}
func (p SVarP) Span() lexer.Span {
	return p.V.span
}
func (p SVarP) String() string {
	return p.V.Name
}

func (_ SCtorP) sPattern() {}
func (p SCtorP) Span() lexer.Span {
	return p.span
}
func (p SCtorP) String() string {
	return fmt.Sprintf("%s %s", p.Ctor.Name, data.JoinToString(p.Fields, " ", func(p SPattern) string {
		return p.String()
	}))
}

func (_ SRecordP) sPattern() {}
func (p SRecordP) Span() lexer.Span {
	return p.span
}
func (p SRecordP) String() string {
	return ShowLabelMap(p.Labels)
}

func (_ SListP) sPattern() {}
func (p SListP) Span() lexer.Span {
	return p.span
}
func (p SListP) String() string {
	if len(p.Elems) == 0 && p.Tail == nil {
		return "[]"
	}
	elems := data.JoinToString(p.Elems, ", ", func(sp SPattern) string {
		return sp.String()
	})
	if p.Tail == nil {
		return fmt.Sprintf("[%s]", elems)
	}
	return fmt.Sprintf("[%s :: %s]", elems, (*p.Tail).String())
}

func (_ SNamed) sPattern() {}
func (p SNamed) Span() lexer.Span {
	return p.span
}
func (p SNamed) String() string {
	return fmt.Sprintf("%s as %s", p.Pat.String(), p.Name.Val)
}

func (_ SUnitP) sPattern() {}
func (p SUnitP) Span() lexer.Span {
	return p.span
}
func (p SUnitP) String() string {
	return "()"
}

func (_ STypeTest) sPattern() {}
func (p STypeTest) Span() lexer.Span {
	return p.span
}
func (p STypeTest) String() string {
	if p.Alias != nil {
		return fmt.Sprintf(":? %s as %s", p.Type.String(), *p.Alias)
	}
	return fmt.Sprintf(":? %s", p.Type.String())
}

func (_ SImplicitP) sPattern() {}
func (p SImplicitP) Span() lexer.Span {
	return p.span
}
func (p SImplicitP) String() string {
	return fmt.Sprintf("{{%s}}", p.Pat.String())
}

func (_ STupleP) sPattern() {}
func (p STupleP) Span() lexer.Span {
	return p.span
}
func (p STupleP) String() string {
	return fmt.Sprintf("%s ; %s", p.P1.String(), p.P2.String())
}

func (_ SRegexP) sPattern() {}
func (p SRegexP) Span() lexer.Span {
	return p.Regex.span
}
func (p SRegexP) String() string {
	return fmt.Sprintf("#\"%s\"", p.Regex.Raw)
}

func (_ SParensP) sPattern() {}
func (p SParensP) Span() lexer.Span {
	return p.span
}
func (p SParensP) String() string {
	return fmt.Sprintf("(%s)", p.Pat.String())
}

func (_ STypeAnnotationP) sPattern() {}
func (p STypeAnnotationP) Span() lexer.Span {
	return p.span
}
func (p STypeAnnotationP) String() string {
	return fmt.Sprintf("%s : %s", p.Par.String(), p.Type.String())
}

////////////////////////////
// source types
////////////////////////////

type SType interface {
	sType()
	Span() lexer.Span
	fmt.Stringer
}

type STConst struct {
	Name  string
	Alias *string
	span  lexer.Span
}

type STApp struct {
	Type  SType
	Types []SType
	span  lexer.Span
}

type STFun struct {
	Arg  SType
	Ret  SType
	span lexer.Span
}

type STParens struct {
	Type SType
	span lexer.Span
}

type STRecord struct {
	Row  SType
	span lexer.Span
}

type STRowEmpty struct {
	span lexer.Span
}

type STRowExtend struct {
	Labels LabelMap[SType]
	Row    SType
	span   lexer.Span
}

type STImplicit struct {
	Type SType
	span lexer.Span
}

func (t STConst) sType() {}
func (t STConst) Span() lexer.Span {
	return t.span
}
func (t STConst) String() string {
	return t.Name
}

func (t STApp) sType() {}
func (t STApp) Span() lexer.Span {
	return t.span
}
func (t STApp) String() string {
	sname := t.Type.String()
	if len(t.Types) == 0 {
		return sname
	}
	return fmt.Sprintf("%s %s", sname, data.JoinToString(t.Types, " ", func(st SType) string {
		return st.String()
	}))
}

func (_ STFun) sType() {}
func (t STFun) Span() lexer.Span {
	return t.span
}
func (t STFun) String() string {
	return fmt.Sprintf("%s -> %s", t.Arg.String(), t.Ret.String())
}

func (_ STParens) sType() {}
func (t STParens) Span() lexer.Span {
	return t.span
}
func (t STParens) String() string {
	return fmt.Sprintf("(%s)", t.Type.String())
}

func (_ STImplicit) sType() {}
func (t STImplicit) Span() lexer.Span {
	return t.span
}
func (t STImplicit) String() string {
	return fmt.Sprintf("{{ %s }}", t.Type.String())
}

func (_ STRowEmpty) sType() {}
func (t STRowEmpty) Span() lexer.Span {
	return t.span
}
func (t STRowEmpty) String() string {
	return "[]"
}

func (_ STRowExtend) sType() {}
func (t STRowExtend) Span() lexer.Span {
	return t.span
}
func (t STRowExtend) String() string {
	labels := ShowLabels(t.Labels, func(k string, v SType) string {
		return fmt.Sprintf("%s : %s", k, v.String())
	})
	return fmt.Sprintf("[ %s ]", labels)
}

func (_ STRecord) sType() {}
func (t STRecord) Span() lexer.Span {
	return t.span
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

// helpers
