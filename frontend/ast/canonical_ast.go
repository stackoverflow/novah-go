package ast

import (
	"fmt"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type Module struct {
	Name          Spanned[string]
	SourceName    string
	Decls         []Decl
	Imports       []Import
	UnusedImports map[string]data.Span
	Comment       *lexer.Comment
}

type Signature struct {
	Type Type
	Span data.Span
}

////////////////////////////////////
// Declarations
////////////////////////////////////

type Decl interface {
	decl()
	GetSpan() data.Span
	GetComment() *lexer.Comment
	IsPublic() bool
	RawName() string
}

type TypeDecl struct {
	Name       Spanned[string]
	TyVars     []string
	DataCtors  []DataCtor
	Span       data.Span
	Visibility Visibility
	Comment    *lexer.Comment
}

type ValDecl struct {
	Name       Spanned[string]
	Exp        Expr
	Recursive  bool
	Span       data.Span
	Signature  *Signature
	Visibility Visibility
	IsInstance bool
	IsOperator bool
	Comment    *lexer.Comment
}

func (_ TypeDecl) decl() {}
func (d TypeDecl) GetSpan() data.Span {
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
func (d ValDecl) GetSpan() data.Span {
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
	Args       []Type
	Visibility Visibility
	Span       data.Span
}

////////////////////////////////////
// Expressions
////////////////////////////////////

type Expr interface {
	expr()
	GetSpan() data.Span
	GetType() Type
	WithType(Type) Type
}

type Typed struct {
	Type Type
}

type Int struct {
	V    int64
	Span data.Span
	Type *Typed
}

type Float struct {
	V    float64
	Span data.Span
	Type *Typed
}

type Complex struct {
	V    complex128
	Span data.Span
	Type *Typed
}

type Char struct {
	V    rune
	Span data.Span
	Type *Typed
}

type String struct {
	V    string
	Span data.Span
	Type *Typed
}

type Bool struct {
	V    bool
	Span data.Span
	Type *Typed
}

type Var struct {
	Name       string
	ModuleName string
	Span       data.Span
	IsOp       bool
	Type       *Typed
}

type Ctor struct {
	Name       string
	ModuleName string
	Span       data.Span
	Type       *Typed
}

type ImplicitVar struct {
	Name       string
	ModuleName string
	Span       data.Span
	Type       *Typed
}

type Lambda struct {
	Binder Binder
	Body   Expr
	Span   data.Span
	Type   *Typed
}

type App struct {
	Fn   Expr
	Arg  Expr
	Span data.Span
	Type *Typed
}

type If struct {
	Cond Expr
	Then Expr
	Else Expr
	Span data.Span
	Type *Typed
}

type Let struct {
	Def  LetDef
	Body Expr
	Span data.Span
	Type *Typed
}

type Match struct {
	Exps  []Expr
	Cases []Case
	Span  data.Span
	Type  *Typed
}

type Ann struct {
	Exp     Expr
	AnnType Type
	Span    data.Span
	Type    *Typed
}

type Do struct {
	Exps []Expr
	Span data.Span
	Type *Typed
}

type Unit struct {
	Span data.Span
	Type *Typed
}

type RecordEmpty struct {
	Span data.Span
	Type *Typed
}

type RecordSelect struct {
	Exp   Expr
	Label Spanned[string]
	Span  data.Span
	Type  *Typed
}

type RecordExtend struct {
	Labels data.LabelMap[Expr]
	Exp    Expr
	Span   data.Span
	Type   *Typed
}

type RecordRestrict struct {
	Exp   Expr
	Label string
	Span  data.Span
	Type  *Typed
}

type RecordUpdate struct {
	Exp   Expr
	Label Spanned[string]
	Value Expr
	IsSet bool
	Span  data.Span
	Type  *Typed
}

type RecordMerge struct {
	Exp1 Expr
	Exp2 Expr
	Span data.Span
	Type *Typed
}

type ListLiteral struct {
	Exps []Expr
	Span data.Span
	Type *Typed
}

type SetLiteral struct {
	Exps []Expr
	Span data.Span
	Type *Typed
}

type Index struct {
	Exp   Expr
	Index Expr
	Span  data.Span
	Type  *Typed
}

type While struct {
	Cond Expr
	Exps []Expr
	Span data.Span
	Type *Typed
}

type Nil struct {
	Span data.Span
	Type *Typed
}

type TypeCast struct {
	Exp  Expr
	Cast Type
	Span data.Span
	Type *Typed
}

func (_ Int) expr() {}
func (e Int) GetSpan() data.Span {
	return e.Span
}
func (e Int) GetType() Type {
	return e.Type.Type
}

func (_ Float) expr() {}
func (e Float) GetSpan() data.Span {
	return e.Span
}
func (e Float) GetType() Type {
	return e.Type.Type
}

func (_ Complex) expr() {}
func (e Complex) GetSpan() data.Span {
	return e.Span
}
func (e Complex) GetType() Type {
	return e.Type.Type
}

func (_ Char) expr() {}
func (e Char) GetSpan() data.Span {
	return e.Span
}
func (e Char) GetType() Type {
	return e.Type.Type
}

func (_ String) expr() {}
func (e String) GetSpan() data.Span {
	return e.Span
}
func (e String) GetType() Type {
	return e.Type.Type
}

func (_ Bool) expr() {}
func (e Bool) GetSpan() data.Span {
	return e.Span
}
func (e Bool) GetType() Type {
	return e.Type.Type
}

func (_ Var) expr() {}
func (e Var) GetSpan() data.Span {
	return e.Span
}
func (e Var) GetType() Type {
	return e.Type.Type
}
func (e Var) Fullname() string {
	if e.ModuleName != "" {
		return fmt.Sprintf("%s.%s", e.ModuleName, e.Name)
	}
	return e.Name
}

func (_ Ctor) expr() {}
func (e Ctor) GetSpan() data.Span {
	return e.Span
}
func (e Ctor) GetType() Type {
	return e.Type.Type
}
func (e Ctor) Fullname() string {
	if e.ModuleName != "" {
		return fmt.Sprintf("%s.%s", e.ModuleName, e.Name)
	}
	return e.Name
}

func (_ ImplicitVar) expr() {}
func (e ImplicitVar) GetSpan() data.Span {
	return e.Span
}
func (e ImplicitVar) GetType() Type {
	return e.Type.Type
}
func (e ImplicitVar) Fullname() string {
	if e.ModuleName != "" {
		return fmt.Sprintf("%s.%s", e.ModuleName, e.Name)
	}
	return e.Name
}

func (_ Lambda) expr() {}
func (e Lambda) GetSpan() data.Span {
	return e.Span
}
func (e Lambda) GetType() Type {
	return e.Type.Type
}

func (_ App) expr() {}
func (e App) GetSpan() data.Span {
	return e.Span
}
func (e App) GetType() Type {
	return e.Type.Type
}

func (_ If) expr() {}
func (e If) GetSpan() data.Span {
	return e.Span
}
func (e If) GetType() Type {
	return e.Type.Type
}

func (_ Let) expr() {}
func (e Let) GetSpan() data.Span {
	return e.Span
}
func (e Let) GetType() Type {
	return e.Type.Type
}

func (_ Match) expr() {}
func (e Match) GetSpan() data.Span {
	return e.Span
}
func (e Match) GetType() Type {
	return e.Type.Type
}

func (_ Ann) expr() {}
func (e Ann) GetSpan() data.Span {
	return e.Span
}
func (e Ann) GetType() Type {
	return e.Type.Type
}

func (_ Do) expr() {}
func (e Do) GetSpan() data.Span {
	return e.Span
}
func (e Do) GetType() Type {
	return e.Type.Type
}

func (_ Unit) expr() {}
func (e Unit) GetSpan() data.Span {
	return e.Span
}
func (e Unit) GetType() Type {
	return e.Type.Type
}

func (_ RecordEmpty) expr() {}
func (e RecordEmpty) GetSpan() data.Span {
	return e.Span
}
func (e RecordEmpty) GetType() Type {
	return e.Type.Type
}

func (_ RecordSelect) expr() {}
func (e RecordSelect) GetSpan() data.Span {
	return e.Span
}
func (e RecordSelect) GetType() Type {
	return e.Type.Type
}

func (_ RecordExtend) expr() {}
func (e RecordExtend) GetSpan() data.Span {
	return e.Span
}
func (e RecordExtend) GetType() Type {
	return e.Type.Type
}

func (_ RecordRestrict) expr() {}
func (e RecordRestrict) GetSpan() data.Span {
	return e.Span
}
func (e RecordRestrict) GetType() Type {
	return e.Type.Type
}

func (_ RecordUpdate) expr() {}
func (e RecordUpdate) GetSpan() data.Span {
	return e.Span
}
func (e RecordUpdate) GetType() Type {
	return e.Type.Type
}

func (_ RecordMerge) expr() {}
func (e RecordMerge) GetSpan() data.Span {
	return e.Span
}
func (e RecordMerge) GetType() Type {
	return e.Type.Type
}

func (_ ListLiteral) expr() {}
func (e ListLiteral) GetSpan() data.Span {
	return e.Span
}
func (e ListLiteral) GetType() Type {
	return e.Type.Type
}

func (_ SetLiteral) expr() {}
func (e SetLiteral) GetSpan() data.Span {
	return e.Span
}
func (e SetLiteral) GetType() Type {
	return e.Type.Type
}

func (_ Index) expr() {}
func (e Index) GetSpan() data.Span {
	return e.Span
}
func (e Index) GetType() Type {
	return e.Type.Type
}

func (_ While) expr() {}
func (e While) GetSpan() data.Span {
	return e.Span
}
func (e While) GetType() Type {
	return e.Type.Type
}

func (_ Nil) expr() {}
func (e Nil) GetSpan() data.Span {
	return e.Span
}
func (e Nil) GetType() Type {
	return e.Type.Type
}

func (_ TypeCast) expr() {}
func (e TypeCast) GetSpan() data.Span {
	return e.Span
}
func (e TypeCast) GetType() Type {
	return e.Type.Type
}

func (e Int) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Float) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Complex) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Char) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e String) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Bool) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Var) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Ctor) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e ImplicitVar) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Lambda) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e App) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e If) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Let) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Match) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Ann) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Do) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Unit) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e RecordEmpty) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e RecordExtend) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e RecordSelect) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e RecordRestrict) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e RecordUpdate) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e RecordMerge) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e ListLiteral) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e SetLiteral) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Index) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e While) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e Nil) WithType(t Type) Type {
	e.Type.Type = t
	return t
}
func (e TypeCast) WithType(t Type) Type {
	e.Type.Type = t
	return t
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
	GetSpan() data.Span
	GetType() Type
	WithType(Type) Type
}

type Wildcard struct {
	Span data.Span
	Type *Typed
}

type LiteralP struct {
	Lit  Expr
	Span data.Span
	Type *Typed
}

type VarP struct {
	V    Var
	Type *Typed
}

type CtorP struct {
	Ctor   Ctor
	Fields []Pattern
	Span   data.Span
	Type   *Typed
}

type RecordP struct {
	Labels data.LabelMap[Pattern]
	Span   data.Span
	Type   *Typed
}

type ListP struct {
	Elems []Pattern
	Tail  Pattern
	Span  data.Span
	Type  *Typed
}

type NamedP struct {
	Pat  Pattern
	Name Spanned[string]
	Span data.Span
	Type *Typed
}

type UnitP struct {
	Span data.Span
	Type *Typed
}

type TypeTest struct {
	Test  Type
	Alias *string
	Span  data.Span
	Type  *Typed
}

type RegexP struct {
	Regex string
	Span  data.Span
	Type  *Typed
}

func (_ Wildcard) pattern() {}
func (p Wildcard) GetSpan() data.Span {
	return p.Span
}
func (p Wildcard) GetType() Type {
	return p.Type.Type
}

func (_ LiteralP) pattern() {}
func (p LiteralP) GetSpan() data.Span {
	return p.Span
}
func (p LiteralP) GetType() Type {
	return p.Type.Type
}

func (_ VarP) pattern() {}
func (p VarP) GetSpan() data.Span {
	return p.V.Span
}
func (p VarP) GetType() Type {
	return p.Type.Type
}

func (_ CtorP) pattern() {}
func (p CtorP) GetSpan() data.Span {
	return p.Ctor.Span
}
func (p CtorP) GetType() Type {
	return p.Type.Type
}

func (_ RecordP) pattern() {}
func (p RecordP) GetSpan() data.Span {
	return p.Span
}
func (p RecordP) GetType() Type {
	return p.Type.Type
}

func (_ ListP) pattern() {}
func (p ListP) GetSpan() data.Span {
	return p.Span
}
func (p ListP) GetType() Type {
	return p.Type.Type
}

func (_ NamedP) pattern() {}
func (p NamedP) GetSpan() data.Span {
	return p.Span
}
func (p NamedP) GetType() Type {
	return p.Type.Type
}

func (_ UnitP) pattern() {}
func (p UnitP) GetSpan() data.Span {
	return p.Span
}
func (p UnitP) GetType() Type {
	return p.Type.Type
}

func (_ TypeTest) pattern() {}
func (p TypeTest) GetSpan() data.Span {
	return p.Span
}
func (p TypeTest) GetType() Type {
	return p.Type.Type
}

func (_ RegexP) pattern() {}
func (p RegexP) GetSpan() data.Span {
	return p.Span
}
func (p RegexP) GetType() Type {
	return p.Type.Type
}

func (p Wildcard) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p LiteralP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p VarP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p CtorP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p RecordP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p ListP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p NamedP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p UnitP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p TypeTest) WithType(t Type) Type {
	p.Type.Type = t
	return t
}
func (p RegexP) WithType(t Type) Type {
	p.Type.Type = t
	return t
}

////////////////////////////////////
// Others
////////////////////////////////////

type ImplicitContext struct {
	Types []Type
	//Env typechecker.Env
	Resolveds []Expr
}

type Binder struct {
	Name       string
	Span       data.Span
	IsImplicit bool
	Type       *Type
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
