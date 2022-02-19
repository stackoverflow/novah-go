package frontend

import (
	"fmt"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
	tc "github.com/stackoverflow/novah-go/frontend/typechecker"
)

type Desugar struct {
	smod ast.SModule
	tc   tc.Typechecker

	usedVars    map[string]bool
	usedTypes   map[string]bool
	usedImports map[string]bool
	declNames   []string
	imports     map[string]string
	modName     string
}

func NewDesugar(smod ast.SModule, tc tc.Typechecker) *Desugar {
	return &Desugar{
		smod:        smod,
		tc:          tc,
		usedVars:    make(map[string]bool),
		usedTypes:   make(map[string]bool),
		usedImports: make(map[string]bool),
		imports:     smod.ResolvedImports,
		modName:     smod.Name.Val,
	}
}

func (d *Desugar) Desugar() (ast.Module, []ast.CompilerProblem) {
	d.declNames = nil
	panic("TBD")
}

func (d *Desugar) desugarType(ty ast.SType, isCtor bool, vars map[string]tc.Type) tc.Type {
	// resolve aliases
	return d.goDesugarType(ty, isCtor, vars, 0)
}

func (d *Desugar) goDesugarType(ty ast.SType, isCtor bool, vars map[string]tc.Type, kindArity int) tc.Type {
	switch t := ty.(type) {
	case ast.STConst:
		{
			d.usedTypes[t.Name] = true
			kind := tc.STAR
			if kindArity > 0 {
				kind = tc.CTOR
			}
			if lexer.IsLower(t.Name) {
				if !isCtor {
					v, has := vars[t.Name]
					if has {
						return v.Clone().WithSpan(ty.GetSpan())
					} else {
						v = d.tc.NewGenVarName(t.Name).WithSpan(ty.GetSpan())
						vars[t.Name] = v
						return v
					}
				} else {
					return tc.TConst{Name: t.Name, Span: ty.GetSpan()}
				}
			} else {
				modName, has := d.imports[t.Fullname()]
				if has {
					d.usedImports[modName] = true
				} else {
					modName = d.modName
				}
				// TODO: check foreigns here
				varName := fmt.Sprintf("%s.%s", modName, t.Name)
				return tc.TConst{Name: varName, Kind: kind, Span: t.Span}
			}
		}
	case ast.STFun:
		return tc.TArrow{Args: []tc.Type{d.goDesugarType(t.Arg, isCtor, vars, 0)}, Ret: d.goDesugarType(t.Ret, isCtor, vars, 0)}
	case ast.STParens:
		return d.goDesugarType(t.Type, isCtor, vars, kindArity)
	case ast.STApp:
		return tc.TApp{
			Type:  d.goDesugarType(t.Type, isCtor, vars, len(t.Types)),
			Types: data.MapSlice(t.Types, func(tt ast.SType) tc.Type { return d.goDesugarType(tt, isCtor, vars, 0) }),
			Span:  t.Span,
		}
	case ast.STRecord:
		return tc.TRecord{Row: d.goDesugarType(t.Row, isCtor, vars, 0), Span: t.Span}
	case ast.STRowEmpty:
		return tc.TRowEmpty{Span: t.Span}
	case ast.STRowExtend:
		{

		}
	}
	panic("")
}
