package typechecker

import "github.com/stackoverflow/novah-go/frontend/ast"

type InstanceEnv struct {
	Type        ast.Type
	IsLambdaVar bool
	TypeName    string
}

type Env struct {
	env       map[string]ast.Type
	types     map[string]ast.Type
	instances map[string]InstanceEnv
}

func NewEnv() *Env {
	return &Env{env: make(map[string]ast.Type), types: make(map[string]ast.Type), instances: make(map[string]InstanceEnv)}
}

func (e *Env) Extend(name string, typ ast.Type) {
	e.env[name] = typ
}

func (e *Env) Lookup(name string) (ast.Type, bool) {
	ty, found := e.env[name]
	return ty, found
}

func (e *Env) Remove(name string) {
	delete(e.env, name)
}

func (e *Env) ExtendType(name string, typ ast.Type) {
	e.types[name] = typ
}

func (e *Env) LookupType(name string) (ast.Type, bool) {
	ty, found := e.types[name]
	return ty, found
}

func (e *Env) ExtendInstance(name string, typ ast.Type, isLambdaVar bool) {
	e.instances[name] = InstanceEnv{Type: typ, IsLambdaVar: isLambdaVar}
}

func (e *Env) ForEachInstance(action func(string, InstanceEnv)) {
	for k, v := range e.instances {
		action(k, v)
	}
}

// Returns a copy of the original env.
// This will reallocate all the maps.
func (e *Env) Fork() *Env {
	env := make(map[string]ast.Type)
	types := make(map[string]ast.Type)
	instances := make(map[string]InstanceEnv)

	for k, v := range e.env {
		env[k] = v
	}
	for k, v := range e.types {
		types[k] = v
	}
	for k, v := range e.instances {
		instances[k] = v
	}
	return &Env{env: env, types: types, instances: instances}
}

// Default types

const (
	PrimInt        = "Int"
	PrimInt8       = "Int8"
	PrimInt16      = "Int16"
	PrimInt32      = "Int32"
	PrimInt64      = "Int64"
	PrimUint       = "Uint"
	PrimUint8      = "Uint8"
	PrimUint16     = "Uint16"
	PrimUint32     = "Uint32"
	PrimUint64     = "Uint64"
	PrimUintptr    = "Uintptr"
	PrimFloat32    = "Float32"
	PrimFloat64    = "Float64"
	PrimComplex64  = "Complex64"
	PrimComplex128 = "Complex128"
	PrimByte       = "Byte"
	PrimBool       = "Bool"
	PrimString     = "String"
	PrimRune       = "Rune"
	PrimUnit       = "Unit"
	PrimList       = "List"
	PrimSet        = "Set"
)

var tInt = ast.TConst{Name: PrimInt}
var tInt8 = ast.TConst{Name: PrimInt8}
var tInt16 = ast.TConst{Name: PrimInt16}
var tInt32 = ast.TConst{Name: PrimInt32}
var tInt64 = ast.TConst{Name: PrimInt64}
var tUint = ast.TConst{Name: PrimUint}
var tUint8 = ast.TConst{Name: PrimUint8}
var tUint16 = ast.TConst{Name: PrimUint16}
var tUint32 = ast.TConst{Name: PrimUint32}
var tUint64 = ast.TConst{Name: PrimUint64}
var tUintptr = ast.TConst{Name: PrimUintptr}
var tFloat32 = ast.TConst{Name: PrimFloat32}
var tFloat64 = ast.TConst{Name: PrimFloat64}
var tComplex64 = ast.TConst{Name: PrimComplex64}
var tComplex128 = ast.TConst{Name: PrimComplex128}
var tByte = ast.TConst{Name: PrimByte}
var tBool = ast.TConst{Name: PrimBool}
var tString = ast.TConst{Name: PrimString}
var tRune = ast.TConst{Name: PrimRune}

var tUnit = ast.TConst{Name: PrimUnit}

var primModuleEnv = ModuleEnv{
	Decls: make(map[string]DeclRef),
	Types: map[string]TypeDeclRef{
		"Byte":       tdecl(tByte),
		"Int":        tdecl(tInt),
		"Int8":       tdecl(tInt8),
		"Int16":      tdecl(tInt16),
		"Int32":      tdecl(tInt32),
		"Int64":      tdecl(tInt64),
		"Uint":       tdecl(tUint),
		"Uint8":      tdecl(tUint8),
		"Uint16":     tdecl(tUint16),
		"Uint32":     tdecl(tUint32),
		"Uint64":     tdecl(tUint64),
		"Uintptr":    tdecl(tUintptr),
		"Float32":    tdecl(tFloat32),
		"Float64":    tdecl(tFloat64),
		"Complex64":  tdecl(tComplex64),
		"Complex128": tdecl(tComplex128),
		"Bool":       tdecl(tBool),
		"Rune":       tdecl(tRune),
		"String":     tdecl(tString),
		"Unit":       tdecl(tUnit),
	},
}

func tdecl(ty ast.Type) TypeDeclRef {
	return TypeDeclRef{Type: ty, Visibility: ast.PUBLIC}
}
