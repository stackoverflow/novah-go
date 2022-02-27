package typechecker

import (
	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type DeclRef struct {
	Type       ast.Type
	Visibility ast.Visibility
	IsInstance bool
	Comment    *lexer.Comment
}

type TypeDeclRef struct {
	Type       ast.Type
	Visibility ast.Visibility
	Ctors      []string
	Comment    *lexer.Comment
}

type ModuleEnv struct {
	Decls map[string]DeclRef
	Types map[string]TypeDeclRef
}

type FullModuleEnv struct {
	Env         ModuleEnv
	Ast         ast.Module
	Aliases     []ast.STypeAliasDecl
	TypeVarsMap map[int]string
	Comment     *lexer.Comment
	IsStdlib    bool
}
