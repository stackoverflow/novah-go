package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/stackoverflow/novah-go/compiler/ast"
	"github.com/stackoverflow/novah-go/data"
)

// Generates go ast
type Optimizer struct {
	mod  ast.Module
	init map[string]ast.Expr
}

func NewOptimizer(mod ast.Module) *Optimizer {
	return &Optimizer{mod: mod, init: make(map[string]ast.Expr)}
}

func (o *Optimizer) Convert() ast.GoPackage {
	decls := make([]ast.GoDecl, 0, len(o.mod.Decls))
	for _, decl := range o.mod.Decls {
		decls = append(decls, o.convertDecl(decl)...)
	}

	return ast.GoPackage{
		Name:       o.mod.Name.Val,
		SourceName: o.mod.SourceName,
		Decls:      decls,
		Pos:        o.mod.Name.Span.Start,
		Comment:    o.mod.Comment,
	}
}

func (o *Optimizer) convertDecl(decl ast.Decl) []ast.GoDecl {
	decls := make([]ast.GoDecl, 0, 1)
	switch d := decl.(type) {
	case ast.TypeDecl:
		if len(d.DataCtors) == 1 {
			ctor := d.DataCtors[0]
			fields := make(map[string]ast.GoType)
			for i, f := range ctor.Args {
				fields[fmt.Sprintf("v%d", i)] = o.convertType(f)
			}
			decls = append(decls, ast.GoStruct{
				Name:    ctor.Name.Val,
				Fields:  fields,
				Pos:     d.Span.Start,
				Comment: d.Comment,
			})
		} else {
			// interface
			method := ast.InterMethod{Name: fmt.Sprintf("__is_%s", d.Name.Val)}
			decls = append(decls, ast.GoInterface{
				Name:    d.Name.Val,
				Methods: []ast.InterMethod{method},
				Pos:     d.Span.Start,
				Comment: d.Comment,
			})
			//structs
			for _, ctor := range d.DataCtors {
				fields := make(map[string]ast.GoType)
				for i, f := range ctor.Args {
					fields[fmt.Sprintf("v%d", i)] = o.convertType(f)
				}
				decls = append(decls, ast.GoStruct{
					Name:    ctor.Name.Val,
					Fields:  fields,
					Pos:     d.Span.Start,
					Comment: d.Comment,
				})

				// implement the interface
				impl := ast.GoFuncDecl{Name: method.Name, Pos: d.Span.Start}
				decls = append(decls, impl)
			}
		}
	case ast.ValDecl:
		if ast.IsConst(d.Exp) {
			decls = append(decls, ast.GoConstDecl{
				Name:    d.Name.Val,
				Val:     o.convertExpr(d.Exp, false).(ast.GoConst),
				Pos:     d.Span.Start,
				Comment: d.Comment,
			})
		} else if lam, ok := d.Exp.(ast.Lambda); ok {
			ty := o.convertType(lam.Type.Type)
			tfun, ok := ty.(ast.GoTFunc)
			if !ok {
				panic("got wrong type for lambda expression")
			}
			params := make(map[string]ast.GoType)
			params[lam.Binder.Name] = tfun.Arg
			body := o.convertExpr(lam.Body, true)
			decls = append(decls, ast.GoFuncDecl{
				Name:    d.Name.Val,
				Params:  params,
				Returns: []ast.GoType{tfun.Ret},
				Body:    &body,
				Pos:     d.Span.Start,
				Comment: d.Comment,
			})
		} else {
			decls = append(decls, ast.GoVarDecl{
				Name:    d.Name.Val,
				Type:    o.convertType(d.Exp.GetType()),
				Pos:     d.Span.Start,
				Comment: d.Comment,
			})
			o.init[d.Name.Val] = d.Exp
		}
	default:
		panic("unknow declaration in optimizer")
	}
	return decls
}

func (o *Optimizer) convertExpr(expr ast.Expr, retur bool) ast.GoExpr {
	switch e := expr.(type) {
	case ast.Int:
		return _return(retur, ast.GoConst{V: e.Raw, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.Float:
		return _return(retur, ast.GoConst{V: e.Raw, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.Complex:
		return _return(retur, ast.GoConst{V: e.Raw, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.Bool:
		return _return(retur, ast.GoConst{V: strconv.FormatBool(e.V), Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.Char:
		return _return(retur, ast.GoConst{V: e.Raw, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.String:
		return _return(retur, ast.GoConst{V: e.Raw, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.Var:
		return _return(retur, ast.GoVar{Name: e.Name, Package: e.ModuleName, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.Ctor:
		return _return(retur, ast.GoVar{Name: e.Name, Package: e.ModuleName, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.ImplicitVar:
		return _return(retur, ast.GoVar{Name: e.Name, Package: e.ModuleName, Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.Lambda:
		{
			ty := o.convertType(e.Type.Type)
			tfun, ok := ty.(ast.GoTFunc)
			if !ok {
				panic("got wrong type for lambda expression")
			}

			args := map[string]ast.GoType{e.Binder.Name: tfun.Arg}
			return _return(retur, ast.GoFunc{
				Args:    args,
				Returns: []ast.GoType{tfun.Ret},
				Body:    o.convertExpr(e.Body, true),
				Type:    ty,
				Pos:     e.Span.Start,
			})
		}
	case ast.App:
		return _return(retur, ast.GoCall{
			Fn:   o.convertExpr(e.Fn, false),
			Args: []ast.GoExpr{o.convertExpr(e.Arg, false)},
			Type: o.convertType(e.Type.Type),
			Pos:  e.Span.Start,
		})
	case ast.If:
		{
			then := o.convertExpr(e.Then, true)
			els := o.convertExpr(e.Else, true)
			return _return(retur, ast.GoCall{
				Fn: ast.GoVar{Name: "__if"},
				Args: []ast.GoExpr{
					o.convertExpr(e.Cond, false),
					ast.GoFunc{
						Args:    nil,
						Returns: []ast.GoType{then.GetType()},
						Body:    then,
						Pos:     then.GetPos(),
					},
					ast.GoFunc{
						Args:    nil,
						Returns: []ast.GoType{els.GetType()},
						Body:    els,
						Pos:     els.GetPos(),
					},
				},
				Type: o.convertType(e.Type.Type),
				Pos:  e.Span.Start,
			})
		}
	case ast.Let:
		{
			stmts := make([]ast.GoExpr, 0, 2)
			typ := o.convertType(e.Def.Expr.GetType())
			varname := e.Def.Binder.Name
			if _if, ok := e.Def.Expr.(ast.If); ok {
				stmts = append(stmts, ast.GoVarDef{Name: varname, Type: typ, Pos: e.Def.Binder.Span.Start})
				stmts = append(stmts, ast.GoIf{
					Cond: o.convertExpr(_if.Cond, false),
					Then: ast.GoSetvar{Name: varname, Exp: o.convertExpr(_if.Then, false)},
					Else: ast.GoSetvar{Name: varname, Exp: o.convertExpr(_if.Else, false)},
					Type: typ,
					Pos:  e.Def.Binder.Span.Start,
				})
			} else {
				stmts = append(stmts, ast.GoLet{
					Binder:   e.Def.Binder.Name,
					BindExpr: o.convertExpr(e.Def.Expr, false),
					Type:     o.convertType(e.Type.Type),
					Pos:      e.Span.Start,
				})
			}
			stmts = append(stmts, o.convertExpr(e.Body, retur))
			return ast.GoStmts{Exps: stmts, Type: o.convertType(e.Type.Type), Pos: e.Span.Start}
		}
	case ast.Do:
		{
			size := len(e.Exps)
			exps := make([]ast.GoExpr, 0, size)
			for i, exp := range e.Exps {
				isRet := false
				if i == size-1 {
					isRet = retur
				}
				exps = append(exps, o.convertExpr(exp, isRet))
			}
			return ast.GoStmts{
				Exps: exps,
				Type: o.convertType(e.Type.Type),
				Pos:  e.Span.Start,
			}
		}
	case ast.Unit:
		return _return(retur, ast.GoUnit{Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	case ast.While:
		return ast.GoWhile{
			Cond: o.convertExpr(e.Cond, false),
			Exps: data.MapSlice(e.Exps, func(ex ast.Expr) ast.GoExpr { return o.convertExpr(ex, false) }),
			Type: o.convertType(e.Type.Type),
			Pos:  e.Span.Start,
		}
	case ast.Nil:
		return _return(retur, ast.GoNil{Type: o.convertType(e.Type.Type), Pos: e.Span.Start})
	default:
		panic("unsuported expression")
	}
}

func (o *Optimizer) convertType(typ ast.Type) ast.GoType {
	switch t := typ.(type) {
	case ast.TConst:
		{
			strs := strings.Split(t.Name, ".")
			pack := ""
			if len(strs) > 1 {
				pack = strs[0]
			}
			name := convertGoType(strs[len(strs)-1])
			return ast.GoTConst{Name: name, Package: pack}
		}
	case ast.TVar:
		switch t.Tvar.Tag {
		case ast.UNBOUND:
			return ast.GoTConst{Name: "any"}
		case ast.GENERIC:
			return ast.GoTConst{Name: "any"}
		case ast.LINK:
			return o.convertType(t.Tvar.Type)
		default:
			panic("impossible")
		}
	case ast.TArrow:
		return ast.GoTFunc{Arg: o.convertType(t.Args[0]), Ret: o.convertType(t.Ret)}
	case ast.TApp:
		return o.convertType(t.Type)
	case ast.TImplicit:
		return o.convertType(t.Type)
	default:
		panic("unknow type " + typ.String())
	}
}

func _return(retur bool, exp ast.GoExpr) ast.GoExpr {
	if !retur {
		return exp
	}
	return ast.GoReturn{Exp: exp, Pos: exp.GetPos()}
}

var primTypes = data.NewSet("Int", "Int8", "Int16", "Int32", "Int64",
	"Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Byte", "Float32", "Float64",
	"Complex64", "Complex128", "Rune", "Uintptr", "Bool", "String")

func convertGoType(name string) string {
	if primTypes.Contains(name) {
		return strings.ToLower(name)
	}
	return name
}
