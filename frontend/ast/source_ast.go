// This file contains the initial parsed AST before
// desugaring and type checking
package ast

import (
	"fmt"
	"strconv"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type Spanned[T any] struct {
	Val  T
	Span data.Span
}

func (s Spanned[T]) Offside() int {
	return s.Span.Start.Col
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
	Name                Spanned[string]
	SourceName          string
	Imports             []Import
	Foreigns            []SForeignImport
	Decls               []SDecl
	Meta                *SMetadata
	Span                data.Span
	Comment             *lexer.Comment
	ResolvedImports     map[string]string
	ResolvedTypealiases []STypeAliasDecl
}

// All = true means all constructors are imported
type DeclarationRef struct {
	Tag   SRefType
	Name  Spanned[string]
	Span  data.Span
	Ctors []Spanned[string]
	All   bool
}

type Import struct {
	Module  Spanned[string]
	Span    data.Span
	Alias   *string
	Auto    bool
	Comment *lexer.Comment
	Defs    []DeclarationRef
}

type SForeignImport struct {
	Type  string
	Alias *string
	Span  data.Span
}

///////////////////////////////////////////////
// Source Declarations
///////////////////////////////////////////////

type SDecl interface {
	GetName() string
	GetVisibility() Visibility
	GetComment() *lexer.Comment
	GetSpan() data.Span
	Metadata() SMetadata
}

type STypeDecl struct {
	Binder     Spanned[string]
	Visibility Visibility
	TyVars     []string
	DataCtors  []SDataCtor
	Span       data.Span
	Comment    *lexer.Comment
	Meta       SMetadata
}

type SValDecl struct {
	Binder     Spanned[string]
	Pats       []SPattern
	Exp        SExpr
	Signature  *SSignature
	Visibility Visibility
	IsInstance bool
	IsOperator bool
	Span       data.Span
	Comment    *lexer.Comment
	Meta       SMetadata
}

type STypeAliasDecl struct {
	Name       string
	TyVars     []string
	Type       SType
	Visibility Visibility
	Span       data.Span
	Comment    *lexer.Comment
	Meta       SMetadata
	Expanded   SType
	FreeVars   map[string]bool
}

type SSignature struct {
	Type SType
	Span data.Span
}

func (d STypeDecl) GetName() string {
	return d.Binder.Val
}
func (d STypeDecl) GetVisibility() Visibility {
	return d.Visibility
}
func (d STypeDecl) GetSpan() data.Span {
	return d.Span
}
func (d STypeDecl) GetComment() *lexer.Comment {
	return d.Comment
}
func (d STypeDecl) Metadata() SMetadata {
	return d.Meta
}

func (d SValDecl) GetName() string {
	return d.Binder.Val
}
func (d SValDecl) GetVisibility() Visibility {
	return d.Visibility
}
func (d SValDecl) GetSpan() data.Span {
	return d.Span
}
func (d SValDecl) GetComment() *lexer.Comment {
	return d.Comment
}
func (d SValDecl) Metadata() SMetadata {
	return d.Meta
}

func (d STypeAliasDecl) GetName() string {
	return d.Name
}
func (d STypeAliasDecl) GetVisibility() Visibility {
	return d.Visibility
}
func (d STypeAliasDecl) GetSpan() data.Span {
	return d.Span
}
func (d STypeAliasDecl) GetComment() *lexer.Comment {
	return d.Comment
}
func (d STypeAliasDecl) Metadata() SMetadata {
	return d.Meta
}

///////////////////////////////////////////////
// Source Data Constructors
///////////////////////////////////////////////

type SDataCtor struct {
	Name       Spanned[string]
	Args       []SType
	Visibility Visibility
	Span       data.Span
}

///////////////////////////////////////////////
// Source Metadada
///////////////////////////////////////////////

type SMetadata struct {
	Data SRecordExtend
}

///////////////////////////////////////////////
// Source Expressions
///////////////////////////////////////////////

type SExpr interface {
	fmt.Stringer
	sExpr()
	GetSpan() data.Span
	GetComment() *lexer.Comment
}

type SInt struct {
	V       int64
	Text    string
	Span    data.Span
	Comment *lexer.Comment
}

type SFloat struct {
	V       float64
	Text    string
	Span    data.Span
	Comment *lexer.Comment
}

type SComplex struct {
	V       complex128
	Text    string
	Span    data.Span
	Comment *lexer.Comment
}

type SString struct {
	V       string
	Raw     string
	Multi   bool
	Span    data.Span
	Comment *lexer.Comment
}

type SChar struct {
	V       rune
	Raw     string
	Span    data.Span
	Comment *lexer.Comment
}

type SBool struct {
	V       bool
	Span    data.Span
	Comment *lexer.Comment
}

type SVar struct {
	Name    string
	Alias   *string
	Span    data.Span
	Comment *lexer.Comment
}

type SOperator struct {
	Name     string
	Alias    *string
	IsPrefix bool
	Span     data.Span
	Comment  *lexer.Comment
}

type SImplicitVar struct {
	Name    string
	Alias   *string
	Span    data.Span
	Comment *lexer.Comment
}

type SConstructor struct {
	Name    string
	Alias   *string
	Span    data.Span
	Comment *lexer.Comment
}

type SPatternLiteral struct {
	Regex   string
	Raw     string
	Span    data.Span
	Comment *lexer.Comment
}

type SLambda struct {
	Pats    []SPattern
	Body    SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SApp struct {
	Fn      SExpr
	Arg     SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SBinApp struct {
	Op      SExpr
	Left    SExpr
	Right   SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SIf struct {
	Cond    SExpr
	Then    SExpr
	Else    SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SLet struct {
	Def     SLetDef
	Body    SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SMatch struct {
	Exprs   []SExpr
	Cases   []SCase
	Span    data.Span
	Comment *lexer.Comment
}

type SAnn struct {
	Exp     SExpr
	Type    SType
	Span    data.Span
	Comment *lexer.Comment
}

type SDo struct {
	Exps    []SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SDoLet struct {
	Def     SLetDef
	Span    data.Span
	Comment *lexer.Comment
}

type SLetBang struct {
	Def     SLetDef
	Body    SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SFor struct {
	Def     SLetDef
	Body    SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SParens struct {
	Exp     SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SUnit struct {
	Span    data.Span
	Comment *lexer.Comment
}

type SRecordEmpty struct {
	Span    data.Span
	Comment *lexer.Comment
}

type SRecordSelect struct {
	Exp     SExpr
	Labels  []Spanned[string]
	Span    data.Span
	Comment *lexer.Comment
}

type SRecordExtend struct {
	Labels  data.LabelMap[SExpr]
	Exp     SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SRecordRestrict struct {
	Exp     SExpr
	Labels  []string
	Span    data.Span
	Comment *lexer.Comment
}

type SRecordUpdate struct {
	Exp     SExpr
	Labels  []Spanned[string]
	Val     SExpr
	IsSet   bool
	Span    data.Span
	Comment *lexer.Comment
}

type SRecordMerge struct {
	Exp1    SExpr
	Exp2    SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SListLiteral struct {
	Exps    []SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SSetLiteral struct {
	Exps    []SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SIndex struct {
	Exp     SExpr
	Index   SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SUnderscore struct {
	Span    data.Span
	Comment *lexer.Comment
}

type SWhile struct {
	Cond    SExpr
	Exps    []SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SComputation struct {
	Builder SVar
	Exps    []SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SReturn struct {
	Exp     SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SYield struct {
	Exp     SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SDoBang struct {
	Exp     SExpr
	Span    data.Span
	Comment *lexer.Comment
}

type SNil struct {
	Span    data.Span
	Comment *lexer.Comment
}

type STypeCast struct {
	Exp     SExpr
	Cast    SType
	Span    data.Span
	Comment *lexer.Comment
}

func (_ SInt) sExpr() {}
func (e SInt) GetSpan() data.Span {
	return e.Span
}
func (e SInt) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SInt) String() string {
	return e.Text
}

func (_ SFloat) sExpr() {}
func (e SFloat) GetSpan() data.Span {
	return e.Span
}
func (e SFloat) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SFloat) String() string {
	return e.Text
}

func (_ SComplex) sExpr() {}
func (e SComplex) GetSpan() data.Span {
	return e.Span
}
func (e SComplex) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SComplex) String() string {
	return e.Text
}

func (_ SString) sExpr() {}
func (e SString) GetSpan() data.Span {
	return e.Span
}
func (e SString) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SString) String() string {
	return e.Raw
}

func (_ SChar) sExpr() {}
func (e SChar) GetSpan() data.Span {
	return e.Span
}
func (e SChar) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SChar) String() string {
	return e.Raw
}

func (_ SBool) sExpr() {}
func (e SBool) GetSpan() data.Span {
	return e.Span
}
func (e SBool) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SBool) String() string {
	return strconv.FormatBool(e.V)
}

func (_ SVar) sExpr() {}
func (e SVar) GetSpan() data.Span {
	return e.Span
}
func (e SVar) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SVar) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}
func (e SVar) Fullname() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

func (_ SOperator) sExpr() {}
func (e SOperator) GetSpan() data.Span {
	return e.Span
}
func (e SOperator) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SOperator) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}
func (e SOperator) Fullname() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

func (_ SImplicitVar) sExpr() {}
func (e SImplicitVar) GetSpan() data.Span {
	return e.Span
}
func (e SImplicitVar) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SImplicitVar) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}
func (e SImplicitVar) Fullname() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

func (_ SConstructor) sExpr() {}
func (e SConstructor) GetSpan() data.Span {
	return e.Span
}
func (e SConstructor) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SConstructor) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}
func (e SConstructor) Fullname() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

func (_ SPatternLiteral) sExpr() {}
func (e SPatternLiteral) GetSpan() data.Span {
	return e.Span
}
func (e SPatternLiteral) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SPatternLiteral) String() string {
	return fmt.Sprintf("#\"%s\"", e.Raw)
}

func (_ SLambda) sExpr() {}
func (e SLambda) GetSpan() data.Span {
	return e.Span
}
func (e SLambda) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SLambda) String() string {
	return "Lambda"
}

func (_ SApp) sExpr() {}
func (e SApp) GetSpan() data.Span {
	return e.Span
}
func (e SApp) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SApp) String() string {
	return "App"
}

func (_ SBinApp) sExpr() {}
func (e SBinApp) GetSpan() data.Span {
	return e.Span
}
func (e SBinApp) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SBinApp) String() string {
	return "BinApp"
}

func (_ SIf) sExpr() {}
func (e SIf) GetSpan() data.Span {
	return e.Span
}
func (e SIf) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SIf) String() string {
	return "If"
}

func (_ SLet) sExpr() {}
func (e SLet) GetSpan() data.Span {
	return e.Span
}
func (e SLet) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SLet) String() string {
	return "Let"
}

func (_ SMatch) sExpr() {}
func (e SMatch) GetSpan() data.Span {
	return e.Span
}
func (e SMatch) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SMatch) String() string {
	return "Match"
}

func (_ SAnn) sExpr() {}
func (e SAnn) GetSpan() data.Span {
	return e.Span
}
func (e SAnn) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SAnn) String() string {
	return "Ann"
}

func (_ SDo) sExpr() {}
func (e SDo) GetSpan() data.Span {
	return e.Span
}
func (e SDo) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SDo) String() string {
	return "Do"
}

func (_ SDoLet) sExpr() {}
func (e SDoLet) GetSpan() data.Span {
	return e.Span
}
func (e SDoLet) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SDoLet) String() string {
	return "DoLet"
}

func (_ SLetBang) sExpr() {}
func (e SLetBang) GetSpan() data.Span {
	return e.Span
}
func (e SLetBang) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SLetBang) String() string {
	return "LetBang"
}

func (_ SFor) sExpr() {}
func (e SFor) GetSpan() data.Span {
	return e.Span
}
func (e SFor) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SFor) String() string {
	return "For"
}

func (_ SParens) sExpr() {}
func (e SParens) GetSpan() data.Span {
	return e.Span
}
func (e SParens) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SParens) String() string {
	return fmt.Sprintf("(%s)", e.Exp.String())
}

func (_ SUnit) sExpr() {}
func (e SUnit) GetSpan() data.Span {
	return e.Span
}
func (e SUnit) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SUnit) String() string {
	return "()"
}

func (_ SRecordEmpty) sExpr() {}
func (e SRecordEmpty) GetSpan() data.Span {
	return e.Span
}
func (e SRecordEmpty) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordEmpty) String() string {
	return "{}"
}

func (_ SRecordSelect) sExpr() {}
func (e SRecordSelect) GetSpan() data.Span {
	return e.Span
}
func (e SRecordSelect) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordSelect) String() string {
	return "RecordSelect"
}

func (_ SRecordExtend) sExpr() {}
func (e SRecordExtend) GetSpan() data.Span {
	return e.Span
}
func (e SRecordExtend) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordExtend) String() string {
	return "RecordExtend"
}

func (_ SRecordRestrict) sExpr() {}
func (e SRecordRestrict) GetSpan() data.Span {
	return e.Span
}
func (e SRecordRestrict) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordRestrict) String() string {
	return "RecordRestrict"
}

func (_ SRecordUpdate) sExpr() {}
func (e SRecordUpdate) GetSpan() data.Span {
	return e.Span
}
func (e SRecordUpdate) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordUpdate) String() string {
	return "RecordUpdate"
}

func (_ SRecordMerge) sExpr() {}
func (e SRecordMerge) GetSpan() data.Span {
	return e.Span
}
func (e SRecordMerge) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordMerge) String() string {
	return "RecordMerge"
}

func (_ SListLiteral) sExpr() {}
func (e SListLiteral) GetSpan() data.Span {
	return e.Span
}
func (e SListLiteral) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SListLiteral) String() string {
	return "[]"
}

func (_ SSetLiteral) sExpr() {}
func (e SSetLiteral) GetSpan() data.Span {
	return e.Span
}
func (e SSetLiteral) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SSetLiteral) String() string {
	return "#{}"
}

func (_ SIndex) sExpr() {}
func (e SIndex) GetSpan() data.Span {
	return e.Span
}
func (e SIndex) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SIndex) String() string {
	return "%.[]"
}

func (_ SUnderscore) sExpr() {}
func (e SUnderscore) GetSpan() data.Span {
	return e.Span
}
func (e SUnderscore) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SUnderscore) String() string {
	return "_"
}

func (_ SWhile) sExpr() {}
func (e SWhile) GetSpan() data.Span {
	return e.Span
}
func (e SWhile) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SWhile) String() string {
	return "While"
}

func (_ SComputation) sExpr() {}
func (e SComputation) GetSpan() data.Span {
	return e.Span
}
func (e SComputation) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SComputation) String() string {
	return "Computation"
}

func (_ SReturn) sExpr() {}
func (e SReturn) GetSpan() data.Span {
	return e.Span
}
func (e SReturn) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SReturn) String() string {
	return fmt.Sprintf("return %s", e.Exp.String())
}

func (_ SYield) sExpr() {}
func (e SYield) GetSpan() data.Span {
	return e.Span
}
func (e SYield) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SYield) String() string {
	return fmt.Sprintf("yield %s", e.Exp.String())
}

func (_ SDoBang) sExpr() {}
func (e SDoBang) GetSpan() data.Span {
	return e.Span
}
func (e SDoBang) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SDoBang) String() string {
	return "Do!"
}

func (_ SNil) sExpr() {}
func (e SNil) GetSpan() data.Span {
	return e.Span
}
func (e SNil) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SNil) String() string {
	return "nil"
}

func (_ STypeCast) sExpr() {}
func (e STypeCast) GetSpan() data.Span {
	return e.Span
}
func (e STypeCast) GetComment() *lexer.Comment {
	return e.Comment
}
func (e STypeCast) String() string {
	return fmt.Sprintf("%s as %s", e.Exp.String(), e.Cast.String())
}

func IsSimple(e *SExpr) bool {
	switch t := (*e).(type) {
	case SIf, SLet, SMatch, SDo, SDoLet, SWhile, SComputation:
		return false
	case SAnn:
		return IsSimple(&t.Exp)
	default:
		return true
	}
}

///////////////////////////////////////////////
// cases in pattern matching
///////////////////////////////////////////////

type SCase struct {
	Pats  []SPattern
	Exp   SExpr
	Guard SExpr
}

func (c *SCase) PatternSpan() data.Span {
	return data.NewSpan(c.Pats[0].GetSpan(), c.Pats[len(c.Pats)-1].GetSpan())
}

///////////////////////////////////////////////
// let definitions
///////////////////////////////////////////////

type SLetDef interface {
	GetExpr() SExpr
}

type SLetBind struct {
	Expr       SExpr
	Name       SBinder
	Pats       []SPattern
	IsInstance bool
	Type       SType
}

type SLetPat struct {
	Expr SExpr
	Pat  SPattern
}

func (def SLetBind) GetExpr() SExpr {
	return def.Expr
}

func (def SLetPat) GetExpr() SExpr {
	return def.Expr
}

type SBinder struct {
	Name       string
	Span       data.Span
	IsImplicit bool
}

///////////////////////////////////////////////
// patterns for pattern matching
///////////////////////////////////////////////

type SPattern interface {
	fmt.Stringer
	sPattern()
	GetSpan() data.Span
}

type SWildcard struct {
	Span data.Span
}

type SLiteralP struct {
	Lit  SExpr
	Span data.Span
}

type SVarP struct {
	V SVar
}

type SCtorP struct {
	Ctor   SConstructor
	Fields []SPattern
	Span   data.Span
}

type SRecordP struct {
	Labels data.LabelMap[SPattern]
	Span   data.Span
}

type SListP struct {
	Elems []SPattern
	Tail  SPattern
	Span  data.Span
}

type SNamed struct {
	Pat  SPattern
	Name Spanned[string]
	Span data.Span
}

type SUnitP struct {
	Span data.Span
}

type STypeTest struct {
	Type  SType
	Alias *string
	Span  data.Span
}

type SImplicitP struct {
	Pat  SPattern
	Span data.Span
}

type STupleP struct {
	P1   SPattern
	P2   SPattern
	Span data.Span
}

type SRegexP struct {
	Regex SPatternLiteral
}

// will be desugared in the desugar phase
type SParensP struct {
	Pat  SPattern
	Span data.Span
}

// will be desugared in the desugar phase
type STypeAnnotationP struct {
	Par  SVar
	Type SType
	Span data.Span
}

func (_ SWildcard) sPattern() {}
func (p SWildcard) GetSpan() data.Span {
	return p.Span
}
func (p SWildcard) String() string {
	return "_"
}

func (_ SLiteralP) sPattern() {}
func (p SLiteralP) GetSpan() data.Span {
	return p.Span
}
func (p SLiteralP) String() string {
	return p.Lit.String()
}

func (_ SVarP) sPattern() {}
func (p SVarP) GetSpan() data.Span {
	return p.V.Span
}
func (p SVarP) String() string {
	return p.V.Name
}

func (_ SCtorP) sPattern() {}
func (p SCtorP) GetSpan() data.Span {
	return p.Span
}
func (p SCtorP) String() string {
	return fmt.Sprintf("%s %s", p.Ctor.Name, data.JoinToString(p.Fields, " "))
}

func (_ SRecordP) sPattern() {}
func (p SRecordP) GetSpan() data.Span {
	return p.Span
}
func (p SRecordP) String() string {
	return data.ShowRaw(p.Labels)
}

func (_ SListP) sPattern() {}
func (p SListP) GetSpan() data.Span {
	return p.Span
}
func (p SListP) String() string {
	if len(p.Elems) == 0 && p.Tail == nil {
		return "[]"
	}
	elems := data.JoinToString(p.Elems, ", ")
	if p.Tail == nil {
		return fmt.Sprintf("[%s]", elems)
	}
	return fmt.Sprintf("[%s :: %s]", elems, p.Tail.String())
}

func (_ SNamed) sPattern() {}
func (p SNamed) GetSpan() data.Span {
	return p.Span
}
func (p SNamed) String() string {
	return fmt.Sprintf("%s as %s", p.Pat.String(), p.Name.Val)
}

func (_ SUnitP) sPattern() {}
func (p SUnitP) GetSpan() data.Span {
	return p.Span
}
func (p SUnitP) String() string {
	return "()"
}

func (_ STypeTest) sPattern() {}
func (p STypeTest) GetSpan() data.Span {
	return p.Span
}
func (p STypeTest) String() string {
	if p.Alias != nil {
		return fmt.Sprintf(":? %s as %s", p.Type.String(), *p.Alias)
	}
	return fmt.Sprintf(":? %s", p.Type.String())
}

func (_ SImplicitP) sPattern() {}
func (p SImplicitP) GetSpan() data.Span {
	return p.Span
}
func (p SImplicitP) String() string {
	return fmt.Sprintf("{{%s}}", p.Pat.String())
}

func (_ STupleP) sPattern() {}
func (p STupleP) GetSpan() data.Span {
	return p.Span
}
func (p STupleP) String() string {
	return fmt.Sprintf("%s ; %s", p.P1.String(), p.P2.String())
}

func (_ SRegexP) sPattern() {}
func (p SRegexP) GetSpan() data.Span {
	return p.Regex.Span
}
func (p SRegexP) String() string {
	return fmt.Sprintf("#\"%s\"", p.Regex.Raw)
}

func (_ SParensP) sPattern() {}
func (p SParensP) GetSpan() data.Span {
	return p.Span
}
func (p SParensP) String() string {
	return fmt.Sprintf("(%s)", p.Pat.String())
}

func (_ STypeAnnotationP) sPattern() {}
func (p STypeAnnotationP) GetSpan() data.Span {
	return p.Span
}
func (p STypeAnnotationP) String() string {
	return fmt.Sprintf("%s : %s", p.Par.String(), p.Type.String())
}

// helpers
