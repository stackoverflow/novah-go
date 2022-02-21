package frontend

import (
	"strings"
	"testing"

	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
	"github.com/stackoverflow/novah-go/frontend/parser"
	"github.com/stackoverflow/novah-go/frontend/typechecker"
)

// Constructors with the same name as the type are not allowed unless there's only one
func TestCtorName(t *testing.T) {
	code := `
	module test

	type Wrong = Wrong | NotWrong`

	ast := parseString(code, t)
	des := NewDesugar(ast, typechecker.NewTypechecker())
	_, err := des.Desugar()
	if err != nil {
		t.Error("fatal error during desugar: " + err.Error())
	}
	errs := des.errors
	if len(errs) != 1 {
		t.Error("expected 1 error")
	}
	if errs[0].Msg != "Multi constructor type cannot have the same name as their type: Wrong." {
		t.Errorf("Got wrong error message: %s", errs[0].Msg)
	}
}

func TestIndentation(t *testing.T) {
	code := `
	module indentationTest
            
	rec = { func: \x -> x, notfun: 0 }
	
	foo =
		\x ->
			x + 1
	
	fun () =
		let x =
			1
			2
		x
	
	fun2 x =
		while true do
			println "hello"
			x
	
	fun3 x =
		case x of
			Some _ -> 1
			None -> 0`

	ast := parseString(code, t)
	des := NewDesugar(ast, typechecker.NewTypechecker())
	_, err := des.Desugar()
	if err != nil {
		t.Error("fatal error during desugar: " + err.Error())
	}

	if len(des.errors) > 0 {
		t.Errorf("expected no errors, got %d", len(des.errors))
	}
}

func TestVisibilityIsSet(t *testing.T) {
	code := `
	module test
            
	pub+
	type AllVis = AllVis1 | AllVis2
	
	pub
	type NoVis = NoVis1 | NoVis2
	
	type Hidden = Hidden1
	
	pub
	x = 3
	
	y = true`

	astree := parseString(code, t)
	des := NewDesugar(astree, typechecker.NewTypechecker())
	mod, err := des.Desugar()
	if err != nil {
		t.Error("fatal error during desugar: " + err.Error())
	}

	allVis := mod.Decls[0].(ast.TypeDecl)
	if allVis.Visibility != ast.PUBLIC {
		t.Error("visibility should be public")
	}
	if allVis.DataCtors[0].Visibility != ast.PUBLIC || allVis.DataCtors[1].Visibility != ast.PUBLIC {
		t.Error("constructors should be public")
	}

	noVis := mod.Decls[1].(ast.TypeDecl)
	if noVis.Visibility != ast.PUBLIC {
		t.Error("visibility should be public")
	}
	if noVis.DataCtors[0].Visibility != ast.PRIVATE || noVis.DataCtors[1].Visibility != ast.PRIVATE {
		t.Error("constructors should be private")
	}

	hidden := mod.Decls[2].(ast.TypeDecl)
	if hidden.Visibility != ast.PRIVATE {
		t.Error("visibility should be private")
	}
	if hidden.DataCtors[0].Visibility != ast.PRIVATE {
		t.Error("constructors should be private")
	}

	x := mod.Decls[3].(ast.ValDecl)
	if x.Visibility != ast.PUBLIC {
		t.Error("visibility should be public")
	}

	y := mod.Decls[4].(ast.ValDecl)
	if y.Visibility != ast.PRIVATE {
		t.Error("visibility should be private")
	}
}

func parseString(code string, t *testing.T) ast.SModule {
	lexer := lexer.New("test.novah", strings.NewReader(code))
	parser := parser.NewParser(lexer)
	mod, errs := parser.ParseFullModule()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf(err.FormatToConsole())
		}
		panic("Error in parsing")
	}
	return mod
}
