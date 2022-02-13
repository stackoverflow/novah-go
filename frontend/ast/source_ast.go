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
	Name() string
	Visibility() Visibility
	Comment() *lexer.Comment
	Span() lexer.Span
	Meta() SMetadata
}

type STypeDecl struct {
	name       string
	visibility Visibility
	Binder     Spanned[string]
	TyVars     []string
	DataCtors  []SDataCtor
	span       lexer.Span
	comment    *lexer.Comment
	meta       SMetadata
}

type SValDecl struct {
	Binder     Spanned[string]
	Pats       []SPattern
	Exp        SExpr
	Signature  SSignature
	visibility Visibility
	IsInstance bool
	IsOperator bool
	span       lexer.Span
	comment    *lexer.Comment
	meta       SMetadata
}

type STypeAliasDecl struct {
	name       string
	TyVars     []string
	Type       SType
	visibility Visibility
	span       lexer.Span
	comment    *lexer.Comment
	meta       SMetadata
	Expanded   *SType
	FreeVars   map[string]bool
}

type SSignature struct {
	Type SType
	Span lexer.Span
}

func (d STypeDecl) Name() string {
	return d.name
}
func (d STypeDecl) Visibility() Visibility {
	return d.visibility
}
func (d STypeDecl) Span() lexer.Span {
	return d.span
}
func (d STypeDecl) Comment() *lexer.Comment {
	return d.comment
}
func (d STypeDecl) Meta() SMetadata {
	return d.meta
}

func (d SValDecl) Name() string {
	return d.Binder.Val
}
func (d SValDecl) Visibility() Visibility {
	return d.visibility
}
func (d SValDecl) Span() lexer.Span {
	return d.span
}
func (d SValDecl) Comment() *lexer.Comment {
	return d.comment
}
func (d SValDecl) Meta() SMetadata {
	return d.meta
}

func (d STypeAliasDecl) Name() string {
	return d.name
}
func (d STypeAliasDecl) Visibility() Visibility {
	return d.visibility
}
func (d STypeAliasDecl) Span() lexer.Span {
	return d.span
}
func (d STypeAliasDecl) Comment() *lexer.Comment {
	return d.comment
}
func (d STypeAliasDecl) Meta() SMetadata {
	return d.meta
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

func (_ SInt) sExpr() {}
func (e SInt) Span() lexer.Span {
	return e.span
}
func (e SInt) Comment() *lexer.Comment {
	return e.comment
}
func (e SInt) String() string {
	return e.Text
}

func (_ SFloat) sExpr() {}
func (e SFloat) Span() lexer.Span {
	return e.span
}
func (e SFloat) Comment() *lexer.Comment {
	return e.comment
}
func (e SFloat) String() string {
	return e.Text
}

func (_ SComplex) sExpr() {}
func (e SComplex) Span() lexer.Span {
	return e.span
}
func (e SComplex) Comment() *lexer.Comment {
	return e.comment
}
func (e SComplex) String() string {
	return e.Text
}

func (_ SString) sExpr() {}
func (e SString) Span() lexer.Span {
	return e.span
}
func (e SString) Comment() *lexer.Comment {
	return e.comment
}
func (e SString) String() string {
	return e.Raw
}

func (_ SChar) sExpr() {}
func (e SChar) Span() lexer.Span {
	return e.span
}
func (e SChar) Comment() *lexer.Comment {
	return e.comment
}
func (e SChar) String() string {
	return e.Raw
}

func (_ SBool) sExpr() {}
func (e SBool) Span() lexer.Span {
	return e.span
}
func (e SBool) Comment() *lexer.Comment {
	return e.comment
}
func (e SBool) String() string {
	return strconv.FormatBool(e.V)
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

func (_ SOperator) sExpr() {}
func (e SOperator) Span() lexer.Span {
	return e.span
}
func (e SOperator) Comment() *lexer.Comment {
	return e.comment
}
func (e SOperator) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

func (_ SImplicitVar) sExpr() {}
func (e SImplicitVar) Span() lexer.Span {
	return e.span
}
func (e SImplicitVar) Comment() *lexer.Comment {
	return e.comment
}
func (e SImplicitVar) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

func (_ SConstructor) sExpr() {}
func (e SConstructor) Span() lexer.Span {
	return e.span
}
func (e SConstructor) Comment() *lexer.Comment {
	return e.comment
}
func (e SConstructor) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s.%s", *e.Alias, e.Name)
	}
	return e.Name
}

func (_ SPatternLiteral) sExpr() {}
func (e SPatternLiteral) Span() lexer.Span {
	return e.span
}
func (e SPatternLiteral) Comment() *lexer.Comment {
	return e.comment
}
func (e SPatternLiteral) String() string {
	return fmt.Sprintf("#\"%s\"", e.Raw)
}

func (_ SLambda) sExpr() {}
func (e SLambda) Span() lexer.Span {
	return e.span
}
func (e SLambda) Comment() *lexer.Comment {
	return e.comment
}
func (e SLambda) String() string {
	return "Lambda"
}

func (_ SApp) sExpr() {}
func (e SApp) Span() lexer.Span {
	return e.span
}
func (e SApp) Comment() *lexer.Comment {
	return e.comment
}
func (e SApp) String() string {
	return "App"
}

func (_ SBinApp) sExpr() {}
func (e SBinApp) Span() lexer.Span {
	return e.span
}
func (e SBinApp) Comment() *lexer.Comment {
	return e.comment
}
func (e SBinApp) String() string {
	return "BinApp"
}

func (_ SIf) sExpr() {}
func (e SIf) Span() lexer.Span {
	return e.span
}
func (e SIf) Comment() *lexer.Comment {
	return e.comment
}
func (e SIf) String() string {
	return "If"
}

func (_ SLet) sExpr() {}
func (e SLet) Span() lexer.Span {
	return e.span
}
func (e SLet) Comment() *lexer.Comment {
	return e.comment
}
func (e SLet) String() string {
	return "Let"
}

func (_ SMatch) sExpr() {}
func (e SMatch) Span() lexer.Span {
	return e.span
}
func (e SMatch) Comment() *lexer.Comment {
	return e.comment
}
func (e SMatch) String() string {
	return "Match"
}

func (_ SAnn) sExpr() {}
func (e SAnn) Span() lexer.Span {
	return e.span
}
func (e SAnn) Comment() *lexer.Comment {
	return e.comment
}
func (e SAnn) String() string {
	return "Ann"
}

func (_ SDo) sExpr() {}
func (e SDo) Span() lexer.Span {
	return e.span
}
func (e SDo) Comment() *lexer.Comment {
	return e.comment
}
func (e SDo) String() string {
	return "Do"
}

func (_ SDoLet) sExpr() {}
func (e SDoLet) Span() lexer.Span {
	return e.span
}
func (e SDoLet) Comment() *lexer.Comment {
	return e.comment
}
func (e SDoLet) String() string {
	return "DoLet"
}

func (_ SLetBang) sExpr() {}
func (e SLetBang) Span() lexer.Span {
	return e.span
}
func (e SLetBang) Comment() *lexer.Comment {
	return e.comment
}
func (e SLetBang) String() string {
	return "LetBang"
}

func (_ SFor) sExpr() {}
func (e SFor) Span() lexer.Span {
	return e.span
}
func (e SFor) Comment() *lexer.Comment {
	return e.comment
}
func (e SFor) String() string {
	return "For"
}

func (_ SParens) sExpr() {}
func (e SParens) Span() lexer.Span {
	return e.span
}
func (e SParens) Comment() *lexer.Comment {
	return e.comment
}
func (e SParens) String() string {
	return fmt.Sprintf("(%s)", e.Exp.String())
}

func (_ SUnit) sExpr() {}
func (e SUnit) Span() lexer.Span {
	return e.span
}
func (e SUnit) Comment() *lexer.Comment {
	return e.comment
}
func (e SUnit) String() string {
	return "()"
}

func (_ SRecordEmpty) sExpr() {}
func (e SRecordEmpty) Span() lexer.Span {
	return e.span
}
func (e SRecordEmpty) Comment() *lexer.Comment {
	return e.comment
}
func (e SRecordEmpty) String() string {
	return "{}"
}

func (_ SRecordSelect) sExpr() {}
func (e SRecordSelect) Span() lexer.Span {
	return e.span
}
func (e SRecordSelect) Comment() *lexer.Comment {
	return e.comment
}
func (e SRecordSelect) String() string {
	return "RecordSelect"
}

func (_ SRecordExtend) sExpr() {}
func (e SRecordExtend) Span() lexer.Span {
	return e.span
}
func (e SRecordExtend) Comment() *lexer.Comment {
	return e.comment
}
func (e SRecordExtend) String() string {
	return "RecordExtend"
}

func (_ SRecordRestrict) sExpr() {}
func (e SRecordRestrict) Span() lexer.Span {
	return e.span
}
func (e SRecordRestrict) Comment() *lexer.Comment {
	return e.comment
}
func (e SRecordRestrict) String() string {
	return "RecordRestrict"
}

func (_ SRecordUpdate) sExpr() {}
func (e SRecordUpdate) Span() lexer.Span {
	return e.span
}
func (e SRecordUpdate) Comment() *lexer.Comment {
	return e.comment
}
func (e SRecordUpdate) String() string {
	return "RecordUpdate"
}

func (_ SRecordMerge) sExpr() {}
func (e SRecordMerge) Span() lexer.Span {
	return e.span
}
func (e SRecordMerge) Comment() *lexer.Comment {
	return e.comment
}
func (e SRecordMerge) String() string {
	return "RecordMerge"
}

func (_ SListLiteral) sExpr() {}
func (e SListLiteral) Span() lexer.Span {
	return e.span
}
func (e SListLiteral) Comment() *lexer.Comment {
	return e.comment
}
func (e SListLiteral) String() string {
	return "[]"
}

func (_ SSetLiteral) sExpr() {}
func (e SSetLiteral) Span() lexer.Span {
	return e.span
}
func (e SSetLiteral) Comment() *lexer.Comment {
	return e.comment
}
func (e SSetLiteral) String() string {
	return "#{}"
}

func (_ SIndex) sExpr() {}
func (e SIndex) Span() lexer.Span {
	return e.span
}
func (e SIndex) Comment() *lexer.Comment {
	return e.comment
}
func (e SIndex) String() string {
	return "%.[]"
}

func (_ SUnderscore) sExpr() {}
func (e SUnderscore) Span() lexer.Span {
	return e.span
}
func (e SUnderscore) Comment() *lexer.Comment {
	return e.comment
}
func (e SUnderscore) String() string {
	return "_"
}

func (_ SWhile) sExpr() {}
func (e SWhile) Span() lexer.Span {
	return e.span
}
func (e SWhile) Comment() *lexer.Comment {
	return e.comment
}
func (e SWhile) String() string {
	return "While"
}

func (_ SComputation) sExpr() {}
func (e SComputation) Span() lexer.Span {
	return e.span
}
func (e SComputation) Comment() *lexer.Comment {
	return e.comment
}
func (e SComputation) String() string {
	return "Computation"
}

func (_ SReturn) sExpr() {}
func (e SReturn) Span() lexer.Span {
	return e.span
}
func (e SReturn) Comment() *lexer.Comment {
	return e.comment
}
func (e SReturn) String() string {
	return fmt.Sprintf("return %s", e.Exp.String())
}

func (_ SYield) sExpr() {}
func (e SYield) Span() lexer.Span {
	return e.span
}
func (e SYield) Comment() *lexer.Comment {
	return e.comment
}
func (e SYield) String() string {
	return fmt.Sprintf("yield %s", e.Exp.String())
}

func (_ SDoBang) sExpr() {}
func (e SDoBang) Span() lexer.Span {
	return e.span
}
func (e SDoBang) Comment() *lexer.Comment {
	return e.comment
}
func (e SDoBang) String() string {
	return "Do!"
}

func (_ SNil) sExpr() {}
func (e SNil) Span() lexer.Span {
	return e.span
}
func (e SNil) Comment() *lexer.Comment {
	return e.comment
}
func (e SNil) String() string {
	return "nil"
}

func (_ STypeCast) sExpr() {}
func (e STypeCast) Span() lexer.Span {
	return e.span
}
func (e STypeCast) Comment() *lexer.Comment {
	return e.comment
}
func (e STypeCast) String() string {
	return fmt.Sprintf("%s as %s", e.Exp.String(), e.Cast.String())
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
