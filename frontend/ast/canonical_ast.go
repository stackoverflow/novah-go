package ast

import (
	"fmt"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/lexer"
	"github.com/stackoverflow/novah-go/frontend/typechecker"
)

type Module struct {
	Name          Spanned[string]
	SourceName    string
	Decls         []Decl
	Imports       []Import
	UnusedImports map[string]lexer.Span
	Comment       *lexer.Comment
}

type Signature struct {
	Type typechecker.Type
	Span lexer.Span
}

////////////////////////////////////
// Declarations
////////////////////////////////////

type Decl interface {
	decl()
	GetSpan() lexer.Span
	GetComment() *lexer.Comment
	IsPublic() bool
	RawName() string
}

type TypeDecl struct {
	Name       Spanned[string]
	TyVars     []string
	DataCtors  []DataCtor
	Span       lexer.Span
	Visibility Visibility
	Comment    *lexer.Comment
}

type ValDecl struct {
	Name       Spanned[string]
	Exp        Expr
	Recursive  bool
	Span       lexer.Span
	Signature  *Signature
	Visibility Visibility
	IsInstance bool
	IsOperator bool
	Comment    *lexer.Comment
}

func (_ TypeDecl) decl() {}
func (d TypeDecl) GetSpan() lexer.Span {
	return d.Span
}
func (d TypeDecl) GetComment() *lexer.Comment {
	return d.Comment
}
func (d TypeDecl) IsPublic() bool {
	return d.Visibility == PUBLIC
}
func (d TypeDecl) RawName() string {
	return d.Name.Val
}

func (_ ValDecl) decl() {}
func (d ValDecl) GetSpan() lexer.Span {
	return d.Span
}
func (d ValDecl) GetComment() *lexer.Comment {
	return d.Comment
}
func (d ValDecl) IsPublic() bool {
	return d.Visibility == PUBLIC
}
func (d ValDecl) RawName() string {
	return d.Name.Val
}

type DataCtor struct {
	Name       Spanned[string]
	Args       []typechecker.Type
	Visibility Visibility
	Span       lexer.Span
}

////////////////////////////////////
// Expressions
////////////////////////////////////

type Expr interface {
	expr()
	GetSpan() lexer.Span
	GetType() typechecker.Type
}

type Int struct {
	V    int64
	Span lexer.Span
	Type typechecker.Type
}

type Float struct {
	V    float64
	Span lexer.Span
	Type typechecker.Type
}

type Complex struct {
	V    complex128
	Span lexer.Span
	Type typechecker.Type
}

type Char struct {
	V    rune
	Span lexer.Span
	Type typechecker.Type
}

type String struct {
	V    string
	Span lexer.Span
	Type typechecker.Type
}

type Bool struct {
	V    bool
	Span lexer.Span
	Type typechecker.Type
}

type Var struct {
	Name       string
	ModuleName *string
	Span       lexer.Span
	IsOp       bool
	Type       typechecker.Type
}

type Ctor struct {
	Name       string
	ModuleName *string
	Span       lexer.Span
	Type       typechecker.Type
}

type ImplicitVar struct {
	Name       string
	ModuleName *string
	Span       lexer.Span
	Type       typechecker.Type
}

type Lambda struct {
	Binder Binder
	Body   Expr
	Span   lexer.Span
	Type   typechecker.Type
}

type App struct {
	Fn   Expr
	Arg  Expr
	Span lexer.Span
	Type typechecker.Type
}

type If struct {
	Cond Expr
	Then Expr
	Else Expr
	Span lexer.Span
	Type typechecker.Type
}

type Let struct {
	Def  LetDef
	Body Expr
	Span lexer.Span
	Type typechecker.Type
}

type Match struct {
	Exps  []Expr
	Cases []Case
	Span  lexer.Span
	Type  typechecker.Type
}

type Ann struct {
	Exp     Expr
	AnnType typechecker.Type
	Span    lexer.Span
	Type    typechecker.Type
}

type Do struct {
	Exps []Expr
	Span lexer.Span
	Type typechecker.Type
}

type Unit struct {
	Span lexer.Span
	Type typechecker.Type
}

type RecordEmpty struct {
	Span lexer.Span
	Type typechecker.Type
}

type RecordSelect struct {
	Exp   Expr
	Label Spanned[string]
	Span  lexer.Span
	Type  typechecker.Type
}

type RecordExtend struct {
	Labels data.LabelMap[Expr]
	Exp    Expr
	Span   lexer.Span
	Type   typechecker.Type
}

type RecordRestrict struct {
	Exp   Expr
	Label string
	Span  lexer.Span
	Type  typechecker.Type
}

type RecordUpdate struct {
	Exp   Expr
	Label Spanned[string]
	Value Expr
	IsSet bool
	Span  lexer.Span
	Type  typechecker.Type
}

type RecordMerge struct {
	Exp1 Expr
	Exp2 Expr
	Span lexer.Span
	Type typechecker.Type
}

type ListLiteral struct {
	Exps []Expr
	Span lexer.Span
	Type typechecker.Type
}

type SetLiteral struct {
	Exps []Expr
	Span lexer.Span
	Type typechecker.Type
}

type Index struct {
	Exp   Expr
	Index Expr
	Span  lexer.Span
	Type  typechecker.Type
}

type While struct {
	Cond Expr
	Exps []Expr
	Span lexer.Span
	Type typechecker.Type
}

type Nil struct {
	Span lexer.Span
	Type typechecker.Type
}

type TypeCast struct {
	Exp  Expr
	Cast typechecker.Type
	Span lexer.Span
	Type typechecker.Type
}

func (_ Int) expr() {}
func (e Int) GetSpan() lexer.Span {
	return e.Span
}
func (e Int) GetType() typechecker.Type {
	return e.Type
}

func (_ Float) expr() {}
func (e Float) GetSpan() lexer.Span {
	return e.Span
}
func (e Float) GetType() typechecker.Type {
	return e.Type
}

func (_ Complex) expr() {}
func (e Complex) GetSpan() lexer.Span {
	return e.Span
}
func (e Complex) GetType() typechecker.Type {
	return e.Type
}

func (_ Char) expr() {}
func (e Char) GetSpan() lexer.Span {
	return e.Span
}
func (e Char) GetType() typechecker.Type {
	return e.Type
}

func (_ String) expr() {}
func (e String) GetSpan() lexer.Span {
	return e.Span
}
func (e String) GetType() typechecker.Type {
	return e.Type
}

func (_ Bool) expr() {}
func (e Bool) GetSpan() lexer.Span {
	return e.Span
}
func (e Bool) GetType() typechecker.Type {
	return e.Type
}

func (_ Var) expr() {}
func (e Var) GetSpan() lexer.Span {
	return e.Span
}
func (e Var) GetType() typechecker.Type {
	return e.Type
}
func (e Var) Fullname() string {
	if e.ModuleName != nil {
		return fmt.Sprintf("%s.%s", *e.ModuleName, e.Name)
	}
	return e.Name
}

func (_ Ctor) expr() {}
func (e Ctor) GetSpan() lexer.Span {
	return e.Span
}
func (e Ctor) GetType() typechecker.Type {
	return e.Type
}
func (e Ctor) Fullname() string {
	if e.ModuleName != nil {
		return fmt.Sprintf("%s.%s", *e.ModuleName, e.Name)
	}
	return e.Name
}

func (_ ImplicitVar) expr() {}
func (e ImplicitVar) GetSpan() lexer.Span {
	return e.Span
}
func (e ImplicitVar) GetType() typechecker.Type {
	return e.Type
}
func (e ImplicitVar) Fullname() string {
	if e.ModuleName != nil {
		return fmt.Sprintf("%s.%s", *e.ModuleName, e.Name)
	}
	return e.Name
}

func (_ Lambda) expr() {}
func (e Lambda) GetSpan() lexer.Span {
	return e.Span
}
func (e Lambda) GetType() typechecker.Type {
	return e.Type
}

func (_ App) expr() {}
func (e App) GetSpan() lexer.Span {
	return e.Span
}
func (e App) GetType() typechecker.Type {
	return e.Type
}

func (_ If) expr() {}
func (e If) GetSpan() lexer.Span {
	return e.Span
}
func (e If) GetType() typechecker.Type {
	return e.Type
}

func (_ Let) expr() {}
func (e Let) GetSpan() lexer.Span {
	return e.Span
}
func (e Let) GetType() typechecker.Type {
	return e.Type
}

func (_ Match) expr() {}
func (e Match) GetSpan() lexer.Span {
	return e.Span
}
func (e Match) GetType() typechecker.Type {
	return e.Type
}

func (_ Ann) expr() {}
func (e Ann) GetSpan() lexer.Span {
	return e.Span
}
func (e Ann) GetType() typechecker.Type {
	return e.Type
}

func (_ Do) expr() {}
func (e Do) GetSpan() lexer.Span {
	return e.Span
}
func (e Do) GetType() typechecker.Type {
	return e.Type
}

func (_ Unit) expr() {}
func (e Unit) GetSpan() lexer.Span {
	return e.Span
}
func (e Unit) GetType() typechecker.Type {
	return e.Type
}

func (_ RecordEmpty) expr() {}
func (e RecordEmpty) GetSpan() lexer.Span {
	return e.Span
}
func (e RecordEmpty) GetType() typechecker.Type {
	return e.Type
}

func (_ RecordSelect) expr() {}
func (e RecordSelect) GetSpan() lexer.Span {
	return e.Span
}
func (e RecordSelect) GetType() typechecker.Type {
	return e.Type
}

func (_ RecordExtend) expr() {}
func (e RecordExtend) GetSpan() lexer.Span {
	return e.Span
}
func (e RecordExtend) GetType() typechecker.Type {
	return e.Type
}

func (_ RecordRestrict) expr() {}
func (e RecordRestrict) GetSpan() lexer.Span {
	return e.Span
}
func (e RecordRestrict) GetType() typechecker.Type {
	return e.Type
}

func (_ RecordUpdate) expr() {}
func (e RecordUpdate) GetSpan() lexer.Span {
	return e.Span
}
func (e RecordUpdate) GetType() typechecker.Type {
	return e.Type
}

func (_ RecordMerge) expr() {}
func (e RecordMerge) GetSpan() lexer.Span {
	return e.Span
}
func (e RecordMerge) GetType() typechecker.Type {
	return e.Type
}

func (_ ListLiteral) expr() {}
func (e ListLiteral) GetSpan() lexer.Span {
	return e.Span
}
func (e ListLiteral) GetType() typechecker.Type {
	return e.Type
}

func (_ SetLiteral) expr() {}
func (e SetLiteral) GetSpan() lexer.Span {
	return e.Span
}
func (e SetLiteral) GetType() typechecker.Type {
	return e.Type
}

func (_ Index) expr() {}
func (e Index) GetSpan() lexer.Span {
	return e.Span
}
func (e Index) GetType() typechecker.Type {
	return e.Type
}

func (_ While) expr() {}
func (e While) GetSpan() lexer.Span {
	return e.Span
}
func (e While) GetType() typechecker.Type {
	return e.Type
}

func (_ Nil) expr() {}
func (e Nil) GetSpan() lexer.Span {
	return e.Span
}
func (e Nil) GetType() typechecker.Type {
	return e.Type
}

func (_ TypeCast) expr() {}
func (e TypeCast) GetSpan() lexer.Span {
	return e.Span
}
func (e TypeCast) GetType() typechecker.Type {
	return e.Type
}

////////////////////////////////////
// Patterns
////////////////////////////////////

type Case struct {
	Patterns []Pattern
	Exp      Expr
	Guard    Expr
}

type Pattern interface {
	pattern()
	GetSpan() lexer.Span
	GetType() typechecker.Type
}

type Wildcard struct {
	Span lexer.Span
	Type typechecker.Type
}

type LiteralP struct {
	Lit  Expr
	Span lexer.Span
	Type typechecker.Type
}

type VarP struct {
	V    Var
	Type typechecker.Type
}

type CtorP struct {
	Ctor   Ctor
	Fields []Pattern
	Span   lexer.Span
	Type   typechecker.Type
}

type RecordP struct {
	Labels data.LabelMap[Pattern]
	Span   lexer.Span
	Type   typechecker.Type
}

type ListP struct {
	Elems []Pattern
	Tail  Pattern
	Span  lexer.Span
	Type  typechecker.Type
}

type NamedP struct {
	Pat  Pattern
	Name Spanned[string]
	Span lexer.Span
	Type typechecker.Type
}

type UnitP struct {
	Span lexer.Span
	Type typechecker.Type
}

type TypeTest struct {
	Test  typechecker.Type
	Alias *string
	Span  lexer.Span
	Type  typechecker.Type
}

type RegexP struct {
	Regex string
	Span  lexer.Span
	Type  typechecker.Type
}

func (_ Wildcard) pattern() {}
func (p Wildcard) GetSpan() lexer.Span {
	return p.Span
}
func (p Wildcard) GetType() typechecker.Type {
	return p.Type
}

func (_ LiteralP) pattern() {}
func (p LiteralP) GetSpan() lexer.Span {
	return p.Span
}
func (p LiteralP) GetType() typechecker.Type {
	return p.Type
}

func (_ VarP) pattern() {}
func (p VarP) GetSpan() lexer.Span {
	return p.V.Span
}
func (p VarP) GetType() typechecker.Type {
	return p.Type
}

func (_ CtorP) pattern() {}
func (p CtorP) GetSpan() lexer.Span {
	return p.Ctor.Span
}
func (p CtorP) GetType() typechecker.Type {
	return p.Type
}

func (_ RecordP) pattern() {}
func (p RecordP) GetSpan() lexer.Span {
	return p.Span
}
func (p RecordP) GetType() typechecker.Type {
	return p.Type
}

func (_ ListP) pattern() {}
func (p ListP) GetSpan() lexer.Span {
	return p.Span
}
func (p ListP) GetType() typechecker.Type {
	return p.Type
}

func (_ NamedP) pattern() {}
func (p NamedP) GetSpan() lexer.Span {
	return p.Span
}
func (p NamedP) GetType() typechecker.Type {
	return p.Type
}

func (_ UnitP) pattern() {}
func (p UnitP) GetSpan() lexer.Span {
	return p.Span
}
func (p UnitP) GetType() typechecker.Type {
	return p.Type
}

func (_ TypeTest) pattern() {}
func (p TypeTest) GetSpan() lexer.Span {
	return p.Span
}
func (p TypeTest) GetType() typechecker.Type {
	return p.Type
}

func (_ RegexP) pattern() {}
func (p RegexP) GetSpan() lexer.Span {
	return p.Span
}
func (p RegexP) GetType() typechecker.Type {
	return p.Type
}

////////////////////////////////////
// Others
////////////////////////////////////

type ImplicitContext struct {
	Types []typechecker.Type
	//Env typechecker.Env
	Resolveds []Expr
}

type Binder struct {
	Name       string
	Span       lexer.Span
	IsImplicit bool
	Type       typechecker.Type
}

type LetDef struct {
	Binder     Binder
	Expr       Expr
	Recursive  bool
	IsInstance bool
}

// helpers

func EverywhereExprAcc[T any](this Expr, f func(Expr) []T) []T {
	res := make([]T, 0, 5)
	EverywhereExprUnit(this, func(e Expr) {
		res = append(res, f(e)...)
	})
	return res
}

func EverywhereExprUnit(this Expr, f func(Expr)) {
	var run func(Expr)
	run = func(exp Expr) {
		switch e := exp.(type) {
		case Lambda:
			{
				f(e)
				run(e.Body)
			}
		case App:
			{
				f(e)
				run(e.Fn)
				run(e.Arg)
			}
		case If:
			{
				f(e)
				run(e.Cond)
				run(e.Then)
				run(e.Else)
			}
		case Let:
			{
				f(e)
				run(e.Def.Expr)
				run(e.Body)
			}
		case Match:
			{
				f(e)
				for _, ex := range e.Exps {
					run(ex)
				}
				for _, cas := range e.Cases {
					run(cas.Exp)
					if cas.Guard != nil {
						run(cas.Guard)
					}
				}
			}
		case Ann:
			{
				run(e.Exp)
				f(e)
			}
		case Do:
			{
				f(e)
				for _, ex := range e.Exps {
					run(ex)
				}
			}
		case RecordSelect:
			{
				f(e)
				run(e.Exp)
			}
		case RecordRestrict:
			{
				f(e)
				run(e.Exp)
			}
		case RecordUpdate:
			{
				f(e)
				run(e.Value)
				run(e.Exp)
			}
		case RecordExtend:
			{
				f(e)
				for _, ex := range e.Labels.Values() {
					run(ex)
				}
				run(e.Exp)
			}
		case RecordMerge:
			{
				f(e)
				run(e.Exp1)
				run(e.Exp2)
			}
		case ListLiteral:
			{
				f(e)
				for _, ex := range e.Exps {
					run(ex)
				}
			}
		case SetLiteral:
			{
				f(e)
				for _, ex := range e.Exps {
					run(ex)
				}
			}
		case Index:
			{
				f(e)
				run(e.Exp)
				run(e.Index)
			}
		case While:
			{
				f(e)
				run(e.Cond)
				for _, ex := range e.Exps {
					run(ex)
				}
			}
		case TypeCast:
			{
				run(e.Exp)
				f(e)
			}
		default:
			f(e)
		}
	}
	run(this)
}
