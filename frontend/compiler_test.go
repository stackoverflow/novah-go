package frontend

import (
	"io/ioutil"
	"testing"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/typechecker"
)

func TestPrimitives(t *testing.T) {
	code := `
	module test

	x = 34
	y = 34.0
	z = 34.0i
	s = "string"
	c = 'a'
	b = true
	`

	env := compileCode(code, t)

	ds := env.Env.Decls
	if ds["x"].Type.String() != "Int" {
		t.Error("x should be an Int")
	}
	if ds["y"].Type.String() != "Float32" {
		t.Error("y should be a Float32")
	}
	if ds["z"].Type.String() != "Complex64" {
		t.Error("z should be a Complex64")
	}
	if ds["s"].Type.String() != "String" {
		t.Error("s should be a String")
	}
	if ds["c"].Type.String() != "Rune" {
		t.Error("c should be a Rune")
	}
	if ds["b"].Type.String() != "Bool" {
		t.Error("b should be a Bool")
	}
}

func TestNumberConversions(t *testing.T) {
	bytes, _ := ioutil.ReadFile("../test_data/numberConversions.novah")
	code := string(bytes)

	env := compileCode(code, t)
	t.Errorf("%v", env)
}

func compileCode(code string, t *testing.T) typechecker.FullModuleEnv {
	env, errs := compileCodeWithErrors(code)
	if len(errs) != 0 {
		for _, err := range errs {
			t.Error(err.FormatToConsole())
		}
	}
	return env
}

func compileCodeWithErrors(code string) (typechecker.FullModuleEnv, []data.CompilerProblem) {
	sources := []Source{{Path: "test", Str: code}}
	opts := Options{}
	comp := &Compiler{sources: sources, opts: opts, env: NewEnviroment(opts)}
	errs := comp.Run(".", true)
	return comp.env.modules["test"], errs
}
