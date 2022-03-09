package ast

import (
	"github.com/stackoverflow/novah-go/compiler/lexer"
	"github.com/stackoverflow/novah-go/data"
)

type GoPackage struct {
	Name       string
	SourceName string
	Decls      []GoDecl
	Pos        data.Pos
	Comment    *lexer.Comment
}

////////////////////////////////////
// Declarations
////////////////////////////////////

type GoDecl interface {
	GetPos() data.Pos
	GetComment() *lexer.Comment
}

type GoInterface struct {
	Name    string
	Methods []InterMethod
	Pos     data.Pos
	Comment *lexer.Comment
}

type GoStruct struct {
	Name    string
	Fields  map[string]GoType
	Pos     data.Pos
	Comment *lexer.Comment
}

type GoConstDecl struct {
	Name    string
	Val     GoConst
	Pos     data.Pos
	Comment *lexer.Comment
}

type GoVarDecl struct {
	Name    string
	Type    GoType
	Pos     data.Pos
	Comment *lexer.Comment
}

type GoFuncDecl struct {
	Name    string
	Params  map[string]GoType
	Returns []GoType
	Body    *GoExpr
	Pos     data.Pos
	Comment *lexer.Comment
}

func (d GoStruct) GetPos() data.Pos {
	return d.Pos
}
func (d GoInterface) GetPos() data.Pos {
	return d.Pos
}
func (d GoConstDecl) GetPos() data.Pos {
	return d.Pos
}
func (d GoVarDecl) GetPos() data.Pos {
	return d.Pos
}
func (d GoFuncDecl) GetPos() data.Pos {
	return d.Pos
}
func (d GoStruct) GetComment() *lexer.Comment {
	return d.Comment
}
func (d GoInterface) GetComment() *lexer.Comment {
	return d.Comment
}
func (d GoConstDecl) GetComment() *lexer.Comment {
	return d.Comment
}
func (d GoVarDecl) GetComment() *lexer.Comment {
	return d.Comment
}
func (d GoFuncDecl) GetComment() *lexer.Comment {
	return d.Comment
}

type InterMethod struct {
	Name string
	Args []GoType
	Ret  []GoType
}

////////////////////////////////////
// Expressions
////////////////////////////////////

type GoExpr interface {
	GetType() GoType
	GetPos() data.Pos
}

type GoConst struct {
	V    string
	Type GoType
	Pos  data.Pos
}

type GoVar struct {
	Name    string
	Package string
	Type    GoType
	Pos     data.Pos
}

type GoFunc struct {
	Args    map[string]GoType
	Returns []GoType
	Body    GoExpr
	Type    GoType
	Pos     data.Pos
}

type GoCall struct {
	Fn   GoExpr
	Args []GoExpr
	Type GoType
	Pos  data.Pos
}

type GoReturn struct {
	Exp GoExpr
	Pos data.Pos
}

type GoIf struct {
	Cond GoExpr
	Then GoExpr
	Else GoExpr
	Type GoType
	Pos  data.Pos
}

type GoVarDef struct {
	Name string
	Type GoType
	Pos  data.Pos
}

type GoLet struct {
	Binder   string
	BindExpr GoExpr
	Type     GoType
	Pos      data.Pos
}

type GoSetvar struct {
	Name string
	Exp  GoExpr
	Pos  data.Pos
}

type GoStmts struct {
	Exps []GoExpr
	Type GoType
	Pos  data.Pos
}

type GoUnit struct {
	Type GoType
	Pos  data.Pos
}

type GoWhile struct {
	Cond GoExpr
	Exps []GoExpr
	Type GoType
	Pos  data.Pos
}

type GoNil struct {
	Type GoType
	Pos  data.Pos
}

func (e GoConst) GetType() GoType {
	return e.Type
}
func (e GoVar) GetType() GoType {
	return e.Type
}
func (e GoFunc) GetType() GoType {
	return e.Type
}
func (e GoCall) GetType() GoType {
	return e.Type
}
func (e GoReturn) GetType() GoType {
	return e.Exp.GetType()
}
func (e GoIf) GetType() GoType {
	return e.Type
}
func (e GoLet) GetType() GoType {
	return e.Type
}
func (e GoVarDef) GetType() GoType {
	return e.Type
}
func (e GoSetvar) GetType() GoType {
	return nil
}
func (e GoStmts) GetType() GoType {
	return e.Type
}
func (e GoUnit) GetType() GoType {
	return e.Type
}
func (e GoWhile) GetType() GoType {
	return e.Type
}
func (e GoNil) GetType() GoType {
	return e.Type
}

func (e GoConst) GetPos() data.Pos {
	return e.Pos
}
func (e GoVar) GetPos() data.Pos {
	return e.Pos
}
func (e GoFunc) GetPos() data.Pos {
	return e.Pos
}
func (e GoCall) GetPos() data.Pos {
	return e.Pos
}
func (e GoReturn) GetPos() data.Pos {
	return e.Pos
}
func (e GoIf) GetPos() data.Pos {
	return e.Pos
}
func (e GoVarDef) GetPos() data.Pos {
	return e.Pos
}
func (e GoLet) GetPos() data.Pos {
	return e.Pos
}
func (e GoSetvar) GetPos() data.Pos {
	return e.Pos
}
func (e GoStmts) GetPos() data.Pos {
	return e.Pos
}
func (e GoUnit) GetPos() data.Pos {
	return e.Pos
}
func (e GoWhile) GetPos() data.Pos {
	return e.Pos
}
func (e GoNil) GetPos() data.Pos {
	return e.Pos
}

////////////////////////////////////
// Type
////////////////////////////////////

type GoType interface {
	goType()
}

type GoTConst struct {
	Name    string
	Package string
}

type GoTFunc struct {
	Arg GoType
	Ret GoType
}

func (_ GoTConst) goType() {}
func (_ GoTFunc) goType()  {}
