package compiler

import (
	"io/ioutil"
	"testing"

	"github.com/stackoverflow/novah-go/compiler/ast"
	"github.com/stackoverflow/novah-go/compiler/typechecker"
	"github.com/stackoverflow/novah-go/data"
	"github.com/stretchr/testify/assert"
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

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "Int", ds["x"].Type.String())
	assert.Equal(t, "Float32", ds["y"].Type.String())
	assert.Equal(t, "Complex64", ds["z"].Type.String())
	assert.Equal(t, "String", ds["s"].Type.String())
	assert.Equal(t, "Rune", ds["c"].Type.String())
	assert.Equal(t, "Bool", ds["b"].Type.String())
}

func TestNumberConversions(t *testing.T) {
	bytes, _ := ioutil.ReadFile("../test_data/numberConversions.novah")
	code := string(bytes)

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "Int", ds["i"].Type.String())
	assert.Equal(t, "Byte", ds["b"].Type.String())
	assert.Equal(t, "Byte", ds["b2"].Type.String())
	assert.Equal(t, "Int16", ds["s"].Type.String())
	assert.Equal(t, "Int64", ds["l"].Type.String())
	assert.Equal(t, "Uint", ds["l2"].Type.String())
	assert.Equal(t, "Float32", ds["d"].Type.String())
	assert.Equal(t, "Float64", ds["f"].Type.String())
	assert.Equal(t, "Int", ds["bi"].Type.String())
	assert.Equal(t, "Byte", ds["bi2"].Type.String())
	assert.Equal(t, "Int", ds["hex"].Type.String())
	assert.Equal(t, "Uint64", ds["hex2"].Type.String())
	assert.Equal(t, "Int16", ds["hex3"].Type.String())
}

func TestIf(t *testing.T) {
	code := `
module test
	
f () = if false then 0 else 1`

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "Unit -> Int", simpleName(ds["f"].Type))
}

func TestSubsumedIf(t *testing.T) {
	code := `
module test

id x = x
	
f _ = if true then 10 else id 0
f2 a = if true then 10 else id a`

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "t1 -> t1", simpleName(ds["id"].Type))
	assert.Equal(t, "t1 -> Int", simpleName(ds["f"].Type))
	assert.Equal(t, "Int -> Int", simpleName(ds["f2"].Type))
}

func TestGenerics(t *testing.T) {
	code := `
module test

x = \y -> y

z = (\y -> y) false`

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "t1 -> t1", simpleName(ds["x"].Type))
	assert.Equal(t, "Bool", simpleName(ds["z"].Type))
}

func TestLet(t *testing.T) {
	code := `
module test

f () =
  let x = 1
	let y = x
	y

g () =
  let a = "bla"
  in a`

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "Unit -> Int", simpleName(ds["f"].Type))
	assert.Equal(t, "Unit -> String", simpleName(ds["g"].Type))
}

func TestPolymorphicLet(t *testing.T) {
	code := `
module test

f _ =
  let id x = x
	id 10
`

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "t1 -> Int", simpleName(ds["f"].Type))
}

func TestRecursiveFunction(t *testing.T) {
	code := `
module test

fact v = case v of
	0 -> 1
	1 -> 1
	x -> fact x
`

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "Int -> Int", simpleName(ds["fact"].Type))
}

func TestMutuallyRecursiveFunctions(t *testing.T) {
	code := `
module test

f1 : Int -> Int
f1 x =
	if true
	then f2 x
	else x

f2 : Int -> Int
f2 x =
	if false
	then f1 x
	else x`

	ds := compileCode(code, t).Env.Decls

	assert.Equal(t, "Int -> Int", simpleName(ds["f1"].Type))
	assert.Equal(t, "Int -> Int", simpleName(ds["f2"].Type))
}

func TestHighRankedTypes(t *testing.T) {
	code := `
module test

type Tuple a b = Tuple a b
            
//fun : (forall a. a -> a) -> Tuple Int String
fun f = Tuple (f 1) (f "a")`

	_, errs := compileCodeWithErrors(code)

	assert.Equal(t, 1, len(errs))
}

func TestRecursiveLets(t *testing.T) {
	code := `
module test

(==) : a -> a -> Bool
(==) _ _ = true

fun n s =
  let rec x y = if x == 1 then y else rec 1 y
  rec n s`

	ds := compileCode(code, t).Env.Decls
	assert.Equal(t, "Int -> t1 -> t1", simpleName(ds["fun"].Type))
}

// helpers

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

func simpleName(ty ast.Type) string {
	id := 0
	m := make(map[int]int)
	get := func(i int) int {
		if v, has := m[i]; has {
			return v
		} else {
			id++
			m[i] = id
			return id
		}
	}

	ast.EverywhereTypeUnit(ty, func(t ast.Type) {
		if tv, isTvar := t.(ast.TVar); isTvar {
			if tv.Tvar.Tag == ast.GENERIC || tv.Tvar.Tag == ast.UNBOUND {
				tv.Tvar.Id = get(tv.Tvar.Id)
			}
		}
	})
	return ty.String()
}
