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
	Name                Spanned[string]
	SourceName          string
	Imports             []SImport
	Foreigns            []SForeignImport
	Decls               []SDecl
	Meta                *SMetadata
	Span                lexer.Span
	Comment             *lexer.Comment
	ResolvedImports     map[string]string
	ResolvedTypealiases []STypeAliasDecl
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
// Source Declarations
///////////////////////////////////////////////

type SDecl interface {
	GetName() string
	GetVisibility() Visibility
	GetComment() *lexer.Comment
	GetSpan() lexer.Span
	Metadata() SMetadata
}

type STypeDecl struct {
	Name       string
	Visibility Visibility
	Binder     Spanned[string]
	TyVars     []string
	DataCtors  []SDataCtor
	Span       lexer.Span
	Comment    *lexer.Comment
	Meta       SMetadata
}

type SValDecl struct {
	Binder     Spanned[string]
	Pats       []SPattern
	Exp        SExpr
	Signature  SSignature
	Visibility Visibility
	IsInstance bool
	IsOperator bool
	Span       lexer.Span
	Comment    *lexer.Comment
	Meta       SMetadata
}

type STypeAliasDecl struct {
	Name       string
	TyVars     []string
	Type       SType
	Visibility Visibility
	Span       lexer.Span
	Comment    *lexer.Comment
	Meta       SMetadata
	Expanded   *SType
	FreeVars   map[string]bool
}

type SSignature struct {
	Type SType
	Span lexer.Span
}

func (d STypeDecl) GetName() string {
	return d.Name
}
func (d STypeDecl) GetVisibility() Visibility {
	return d.Visibility
}
func (d STypeDecl) GetSpan() lexer.Span {
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
func (d SValDecl) GetSpan() lexer.Span {
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
func (d STypeAliasDecl) GetSpan() lexer.Span {
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
	Span       lexer.Span
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
	GetSpan() lexer.Span
	GetComment() *lexer.Comment
}

type SInt struct {
	V       int64
	Text    string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SFloat struct {
	V       float64
	Text    string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SComplex struct {
	V       complex128
	Text    string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SString struct {
	V       string
	Raw     string
	Multi   bool
	Span    lexer.Span
	Comment *lexer.Comment
}

type SChar struct {
	V       rune
	Raw     string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SBool struct {
	V       bool
	Span    lexer.Span
	Comment *lexer.Comment
}

type SVar struct {
	Name    string
	Alias   *string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SOperator struct {
	Name     string
	Alias    *string
	IsPrefix bool
	Span     lexer.Span
	Comment  *lexer.Comment
}

type SImplicitVar struct {
	Name    string
	Alias   *string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SConstructor struct {
	Name    string
	Alias   *string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SPatternLiteral struct {
	Regex   string
	Raw     string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SLambda struct {
	Pats    []SPattern
	Body    SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SApp struct {
	Fn      SExpr
	Arg     SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SBinApp struct {
	Op      SExpr
	Left    SExpr
	Right   SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SIf struct {
	Cond    SExpr
	Then    SExpr
	Else    data.Option[SExpr]
	Span    lexer.Span
	Comment *lexer.Comment
}

type SLet struct {
	Def     SLetDef
	Body    SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SMatch struct {
	Exprs   []SExpr
	Cases   []SCase
	Span    lexer.Span
	Comment *lexer.Comment
}

type SAnn struct {
	Exp     SExpr
	Type    SType
	Span    lexer.Span
	Comment *lexer.Comment
}

type SDo struct {
	Exps    []SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SDoLet struct {
	Def     SLetDef
	Span    lexer.Span
	Comment *lexer.Comment
}

type SLetBang struct {
	Def     SLetDef
	Body    *SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SFor struct {
	Def     SLetDef
	Body    SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SParens struct {
	Exp     SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SUnit struct {
	Span    lexer.Span
	Comment *lexer.Comment
}

type SRecordEmpty struct {
	Span    lexer.Span
	Comment *lexer.Comment
}

type SRecordSelect struct {
	Exp     SExpr
	Labels  []Spanned[string]
	Span    lexer.Span
	Comment *lexer.Comment
}

type SRecordExtend struct {
	Labels  LabelMap[SExpr]
	Exp     SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SRecordRestrict struct {
	Exp     SExpr
	Labels  []string
	Span    lexer.Span
	Comment *lexer.Comment
}

type SRecordUpdate struct {
	Exp     SExpr
	Labels  []Spanned[string]
	Val     SExpr
	IsSet   bool
	Span    lexer.Span
	Comment *lexer.Comment
}

type SRecordMerge struct {
	Exp1    SExpr
	Exp2    SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SListLiteral struct {
	Exps    []SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SSetLiteral struct {
	Exps    []SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SIndex struct {
	Exp     SExpr
	Index   SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SUnderscore struct {
	Span    lexer.Span
	Comment *lexer.Comment
}

type SWhile struct {
	Cond    SExpr
	Exps    []SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SComputation struct {
	Builder SVar
	Exps    []SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SReturn struct {
	Exp     SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SYield struct {
	Exp     SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SDoBang struct {
	Exp     SExpr
	Span    lexer.Span
	Comment *lexer.Comment
}

type SNil struct {
	Span    lexer.Span
	Comment *lexer.Comment
}

type STypeCast struct {
	Exp     SExpr
	Cast    SType
	Span    lexer.Span
	Comment *lexer.Comment
}

func (_ SInt) sExpr() {}
func (e SInt) GetSpan() lexer.Span {
	return e.Span
}
func (e SInt) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SInt) String() string {
	return e.Text
}

func (_ SFloat) sExpr() {}
func (e SFloat) GetSpan() lexer.Span {
	return e.Span
}
func (e SFloat) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SFloat) String() string {
	return e.Text
}

func (_ SComplex) sExpr() {}
func (e SComplex) GetSpan() lexer.Span {
	return e.Span
}
func (e SComplex) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SComplex) String() string {
	return e.Text
}

func (_ SString) sExpr() {}
func (e SString) GetSpan() lexer.Span {
	return e.Span
}
func (e SString) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SString) String() string {
	return e.Raw
}

func (_ SChar) sExpr() {}
func (e SChar) GetSpan() lexer.Span {
	return e.Span
}
func (e SChar) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SChar) String() string {
	return e.Raw
}

func (_ SBool) sExpr() {}
func (e SBool) GetSpan() lexer.Span {
	return e.Span
}
func (e SBool) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SBool) String() string {
	return strconv.FormatBool(e.V)
}

func (_ SVar) sExpr() {}
func (e SVar) GetSpan() lexer.Span {
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

func (_ SOperator) sExpr() {}
func (e SOperator) GetSpan() lexer.Span {
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

func (_ SImplicitVar) sExpr() {}
func (e SImplicitVar) GetSpan() lexer.Span {
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

func (_ SConstructor) sExpr() {}
func (e SConstructor) GetSpan() lexer.Span {
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

func (_ SPatternLiteral) sExpr() {}
func (e SPatternLiteral) GetSpan() lexer.Span {
	return e.Span
}
func (e SPatternLiteral) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SPatternLiteral) String() string {
	return fmt.Sprintf("#\"%s\"", e.Raw)
}

func (_ SLambda) sExpr() {}
func (e SLambda) GetSpan() lexer.Span {
	return e.Span
}
func (e SLambda) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SLambda) String() string {
	return "Lambda"
}

func (_ SApp) sExpr() {}
func (e SApp) GetSpan() lexer.Span {
	return e.Span
}
func (e SApp) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SApp) String() string {
	return "App"
}

func (_ SBinApp) sExpr() {}
func (e SBinApp) GetSpan() lexer.Span {
	return e.Span
}
func (e SBinApp) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SBinApp) String() string {
	return "BinApp"
}

func (_ SIf) sExpr() {}
func (e SIf) GetSpan() lexer.Span {
	return e.Span
}
func (e SIf) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SIf) String() string {
	return "If"
}

func (_ SLet) sExpr() {}
func (e SLet) GetSpan() lexer.Span {
	return e.Span
}
func (e SLet) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SLet) String() string {
	return "Let"
}

func (_ SMatch) sExpr() {}
func (e SMatch) GetSpan() lexer.Span {
	return e.Span
}
func (e SMatch) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SMatch) String() string {
	return "Match"
}

func (_ SAnn) sExpr() {}
func (e SAnn) GetSpan() lexer.Span {
	return e.Span
}
func (e SAnn) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SAnn) String() string {
	return "Ann"
}

func (_ SDo) sExpr() {}
func (e SDo) GetSpan() lexer.Span {
	return e.Span
}
func (e SDo) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SDo) String() string {
	return "Do"
}

func (_ SDoLet) sExpr() {}
func (e SDoLet) GetSpan() lexer.Span {
	return e.Span
}
func (e SDoLet) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SDoLet) String() string {
	return "DoLet"
}

func (_ SLetBang) sExpr() {}
func (e SLetBang) GetSpan() lexer.Span {
	return e.Span
}
func (e SLetBang) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SLetBang) String() string {
	return "LetBang"
}

func (_ SFor) sExpr() {}
func (e SFor) GetSpan() lexer.Span {
	return e.Span
}
func (e SFor) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SFor) String() string {
	return "For"
}

func (_ SParens) sExpr() {}
func (e SParens) GetSpan() lexer.Span {
	return e.Span
}
func (e SParens) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SParens) String() string {
	return fmt.Sprintf("(%s)", e.Exp.String())
}

func (_ SUnit) sExpr() {}
func (e SUnit) GetSpan() lexer.Span {
	return e.Span
}
func (e SUnit) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SUnit) String() string {
	return "()"
}

func (_ SRecordEmpty) sExpr() {}
func (e SRecordEmpty) GetSpan() lexer.Span {
	return e.Span
}
func (e SRecordEmpty) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordEmpty) String() string {
	return "{}"
}

func (_ SRecordSelect) sExpr() {}
func (e SRecordSelect) GetSpan() lexer.Span {
	return e.Span
}
func (e SRecordSelect) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordSelect) String() string {
	return "RecordSelect"
}

func (_ SRecordExtend) sExpr() {}
func (e SRecordExtend) GetSpan() lexer.Span {
	return e.Span
}
func (e SRecordExtend) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordExtend) String() string {
	return "RecordExtend"
}

func (_ SRecordRestrict) sExpr() {}
func (e SRecordRestrict) GetSpan() lexer.Span {
	return e.Span
}
func (e SRecordRestrict) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordRestrict) String() string {
	return "RecordRestrict"
}

func (_ SRecordUpdate) sExpr() {}
func (e SRecordUpdate) GetSpan() lexer.Span {
	return e.Span
}
func (e SRecordUpdate) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordUpdate) String() string {
	return "RecordUpdate"
}

func (_ SRecordMerge) sExpr() {}
func (e SRecordMerge) GetSpan() lexer.Span {
	return e.Span
}
func (e SRecordMerge) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SRecordMerge) String() string {
	return "RecordMerge"
}

func (_ SListLiteral) sExpr() {}
func (e SListLiteral) GetSpan() lexer.Span {
	return e.Span
}
func (e SListLiteral) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SListLiteral) String() string {
	return "[]"
}

func (_ SSetLiteral) sExpr() {}
func (e SSetLiteral) GetSpan() lexer.Span {
	return e.Span
}
func (e SSetLiteral) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SSetLiteral) String() string {
	return "#{}"
}

func (_ SIndex) sExpr() {}
func (e SIndex) GetSpan() lexer.Span {
	return e.Span
}
func (e SIndex) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SIndex) String() string {
	return "%.[]"
}

func (_ SUnderscore) sExpr() {}
func (e SUnderscore) GetSpan() lexer.Span {
	return e.Span
}
func (e SUnderscore) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SUnderscore) String() string {
	return "_"
}

func (_ SWhile) sExpr() {}
func (e SWhile) GetSpan() lexer.Span {
	return e.Span
}
func (e SWhile) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SWhile) String() string {
	return "While"
}

func (_ SComputation) sExpr() {}
func (e SComputation) GetSpan() lexer.Span {
	return e.Span
}
func (e SComputation) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SComputation) String() string {
	return "Computation"
}

func (_ SReturn) sExpr() {}
func (e SReturn) GetSpan() lexer.Span {
	return e.Span
}
func (e SReturn) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SReturn) String() string {
	return fmt.Sprintf("return %s", e.Exp.String())
}

func (_ SYield) sExpr() {}
func (e SYield) GetSpan() lexer.Span {
	return e.Span
}
func (e SYield) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SYield) String() string {
	return fmt.Sprintf("yield %s", e.Exp.String())
}

func (_ SDoBang) sExpr() {}
func (e SDoBang) GetSpan() lexer.Span {
	return e.Span
}
func (e SDoBang) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SDoBang) String() string {
	return "Do!"
}

func (_ SNil) sExpr() {}
func (e SNil) GetSpan() lexer.Span {
	return e.Span
}
func (e SNil) GetComment() *lexer.Comment {
	return e.Comment
}
func (e SNil) String() string {
	return "nil"
}

func (_ STypeCast) sExpr() {}
func (e STypeCast) GetSpan() lexer.Span {
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
	GetSpan() lexer.Span
}

type SWildcard struct {
	Span lexer.Span
}

type SLiteralP struct {
	Lit  SExpr
	Span lexer.Span
}

type SVarP struct {
	V SVar
}

type SCtorP struct {
	Ctor   SConstructor
	Fields []SPattern
	Span   lexer.Span
}

type SRecordP struct {
	Labels LabelMap[SPattern]
	Span   lexer.Span
}

type SListP struct {
	Elems []SPattern
	Tail  *SPattern
	Span  lexer.Span
}

type SNamed struct {
	Pat  SPattern
	Name Spanned[string]
	Span lexer.Span
}

type SUnitP struct {
	Span lexer.Span
}

type STypeTest struct {
	Type  SType
	Alias *string
	Span  lexer.Span
}

type SImplicitP struct {
	Pat  SPattern
	Span lexer.Span
}

type STupleP struct {
	P1   SPattern
	P2   SPattern
	Span lexer.Span
}

type SRegexP struct {
	Regex SPatternLiteral
}

// will be desugared in the desugar phase
type SParensP struct {
	Pat  SPattern
	Span lexer.Span
}

// will be desugared in the desugar phase
type STypeAnnotationP struct {
	Par  SVar
	Type SType
	Span lexer.Span
}

func (_ SWildcard) sPattern() {}
func (p SWildcard) GetSpan() lexer.Span {
	return p.Span
}
func (p SWildcard) String() string {
	return "_"
}

func (_ SLiteralP) sPattern() {}
func (p SLiteralP) GetSpan() lexer.Span {
	return p.Span
}
func (p SLiteralP) String() string {
	return p.Lit.String()
}

func (_ SVarP) sPattern() {}
func (p SVarP) GetSpan() lexer.Span {
	return p.V.Span
}
func (p SVarP) String() string {
	return p.V.Name
}

func (_ SCtorP) sPattern() {}
func (p SCtorP) GetSpan() lexer.Span {
	return p.Span
}
func (p SCtorP) String() string {
	return fmt.Sprintf("%s %s", p.Ctor.Name, data.JoinToString(p.Fields, " ", func(p SPattern) string {
		return p.String()
	}))
}

func (_ SRecordP) sPattern() {}
func (p SRecordP) GetSpan() lexer.Span {
	return p.Span
}
func (p SRecordP) String() string {
	return ShowLabelMap(p.Labels)
}

func (_ SListP) sPattern() {}
func (p SListP) GetSpan() lexer.Span {
	return p.Span
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
func (p SNamed) GetSpan() lexer.Span {
	return p.Span
}
func (p SNamed) String() string {
	return fmt.Sprintf("%s as %s", p.Pat.String(), p.Name.Val)
}

func (_ SUnitP) sPattern() {}
func (p SUnitP) GetSpan() lexer.Span {
	return p.Span
}
func (p SUnitP) String() string {
	return "()"
}

func (_ STypeTest) sPattern() {}
func (p STypeTest) GetSpan() lexer.Span {
	return p.Span
}
func (p STypeTest) String() string {
	if p.Alias != nil {
		return fmt.Sprintf(":? %s as %s", p.Type.String(), *p.Alias)
	}
	return fmt.Sprintf(":? %s", p.Type.String())
}

func (_ SImplicitP) sPattern() {}
func (p SImplicitP) GetSpan() lexer.Span {
	return p.Span
}
func (p SImplicitP) String() string {
	return fmt.Sprintf("{{%s}}", p.Pat.String())
}

func (_ STupleP) sPattern() {}
func (p STupleP) GetSpan() lexer.Span {
	return p.Span
}
func (p STupleP) String() string {
	return fmt.Sprintf("%s ; %s", p.P1.String(), p.P2.String())
}

func (_ SRegexP) sPattern() {}
func (p SRegexP) GetSpan() lexer.Span {
	return p.Regex.Span
}
func (p SRegexP) String() string {
	return fmt.Sprintf("#\"%s\"", p.Regex.Raw)
}

func (_ SParensP) sPattern() {}
func (p SParensP) GetSpan() lexer.Span {
	return p.Span
}
func (p SParensP) String() string {
	return fmt.Sprintf("(%s)", p.Pat.String())
}

func (_ STypeAnnotationP) sPattern() {}
func (p STypeAnnotationP) GetSpan() lexer.Span {
	return p.Span
}
func (p STypeAnnotationP) String() string {
	return fmt.Sprintf("%s : %s", p.Par.String(), p.Type.String())
}

////////////////////////////
// source types
////////////////////////////

type SType interface {
	sType()
	GetSpan() lexer.Span
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
	Labels LabelMap[SType]
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

func (t STApp) sType() {}
func (t STApp) GetSpan() lexer.Span {
	return t.Span
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
func (t STFun) GetSpan() lexer.Span {
	return t.Span
}
func (t STFun) String() string {
	return fmt.Sprintf("%s -> %s", t.Arg.String(), t.Ret.String())
}

func (_ STParens) sType() {}
func (t STParens) GetSpan() lexer.Span {
	return t.Span
}
func (t STParens) String() string {
	return fmt.Sprintf("(%s)", t.Type.String())
}

func (_ STImplicit) sType() {}
func (t STImplicit) GetSpan() lexer.Span {
	return t.Span
}
func (t STImplicit) String() string {
	return fmt.Sprintf("{{ %s }}", t.Type.String())
}

func (_ STRowEmpty) sType() {}
func (t STRowEmpty) GetSpan() lexer.Span {
	return t.Span
}
func (t STRowEmpty) String() string {
	return "[]"
}

func (_ STRowExtend) sType() {}
func (t STRowExtend) GetSpan() lexer.Span {
	return t.Span
}
func (t STRowExtend) String() string {
	labels := ShowLabels(t.Labels, func(k string, v SType) string {
		return fmt.Sprintf("%s : %s", k, v.String())
	})
	return fmt.Sprintf("[ %s ]", labels)
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

// helpers
