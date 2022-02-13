package parser

import (
	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type Parser struct {
	lex        *lexer.Lexer
	sourceName string

	iter       PeekableIterator
	moduleName *string
	errors     []ast.CompilerProblem
}
