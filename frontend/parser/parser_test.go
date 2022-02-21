package parser

import (
	"io"
	"os"
	"testing"

	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
	"github.com/stackoverflow/novah-go/test"
)

var fmtt = ast.NewFormatter()

func TestOperators(t *testing.T) {
	mod := parseResource("../../test_data/operators.novah", t)

	pick := func(i int) string {
		return fmtt.ShowExpr(mod.Decls[i].(ast.SValDecl).Exp)
	}

	x := "w || r && x || p"
	fl := "(a >> b >> c) 1"
	fr := "(a << b << c) 1"
	l1 := "3 + 7 ^ 4 * 6 * 9"
	l2 := "3 ^ (7 * 4) * 6 + 9"
	r1 := "bla 3 $ df 4 $ pa"
	r2 := "3 :: 5 :: 7 :: Nil"
	ap := "fn 3 4 5"
	a2 := "fn (fn2 8)"
	co := "fn 'x' y (Some (3 + 4) 1)"

	test.Equals(t, pick(0), x)
	test.Equals(t, pick(1), fl)
	test.Equals(t, pick(2), fr)
	test.Equals(t, pick(3), l1)
	test.Equals(t, pick(4), l2)
	test.Equals(t, pick(5), r1)
	test.Equals(t, pick(6), r2)
	test.Equals(t, pick(7), ap)
	test.Equals(t, pick(8), a2)
	test.Equals(t, pick(9), co)
}

func TestLambdas(t *testing.T) {
	mod := parseResource("../../test_data/lambda.novah", t)

	l1 := mod.Decls[0].(ast.SValDecl).Exp
	l2 := mod.Decls[1].(ast.SValDecl).Exp

	expl1 := "\\x -> x"
	expl2 := "\\x y z -> 1"

	test.Equals(t, fmtt.ShowExpr(l1), expl1)
	test.Equals(t, fmtt.ShowExpr(l2), expl2)
}

func TestComments(t *testing.T) {
	mod := parseResource("../../test_data/comments.novah", t)

	data := mod.Decls[0].GetComment()
	typ := mod.Decls[1].GetComment()
	val := mod.Decls[2].GetComment()
	nocomm := mod.Decls[3].GetComment()

	test.Equals(t, data.Text, " comments on type definitions work")
	test.Equals(t, typ.Text, "\n comments on var\n types work\n")
	test.Equals(t, val.Text, " comments on var declaration work\n and are concatenated")
	test.Equals(t, nocomm, nil)
}

func TestTypeHints(t *testing.T) {
	mod := parseResource("../../test_data/hints.novah", t)

	x := mod.Decls[0].(ast.SValDecl).Exp.(ast.SAnn)

	test.Equals(t, fmtt.ShowType(x.Type), "String")
}

func TestRecords(t *testing.T) {
	parseResource("../../test_data/records.novah", t)
	// should not panic
}

func parseResource(input string, t *testing.T) ast.SModule {
	reader, _ := os.Open(input)
	defer reader.Close()

	return parseString(reader, input, t)
}

func parseString(reader io.Reader, name string, t *testing.T) ast.SModule {
	lexer := lexer.New(name, reader)
	parser := NewParser(lexer)
	mod, errs := parser.ParseFullModule()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf(err.FormatToConsole())
		}
		panic("Error in parsing")
	}
	return mod
}
