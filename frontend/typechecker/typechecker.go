package typechecker

import (
	"fmt"
	"strings"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
)

type Typechecker struct {
	TypeVarMap map[int]string
	currentId  int
	env        *Env
	infer      *Inference
	uni        *Unification
	context    TypingContext
}

func NewTypechecker() *Typechecker {
	env := NewEnv()
	env.AddPrimitiveTypes()
	tc := &Typechecker{
		TypeVarMap: make(map[int]string),
		env:        env,
	}
	uni := NewUnification(tc)
	infer := NewInference(tc, uni)
	tc.infer = infer
	tc.uni = uni
	return tc
}

func (tc *Typechecker) NewVar(level ast.Level) ast.Type {
	tc.currentId++
	return ast.TVar{Tvar: &ast.TypeVar{Tag: ast.UNBOUND, Id: tc.currentId, Level: level}}
}

func (tc *Typechecker) NewGenVar() ast.TVar {
	tc.currentId++
	id := tc.currentId
	return ast.TVar{Tvar: &ast.TypeVar{Tag: ast.GENERIC, Id: id}}
}

func (tc *Typechecker) NewGenVarName(name string) ast.TVar {
	tc.currentId++
	id := tc.currentId
	tc.TypeVarMap[id] = name
	return ast.TVar{Tvar: &ast.TypeVar{Tag: ast.GENERIC, Id: id}}
}

func (tc *Typechecker) Infer(mod ast.Module) (ModuleEnv, error) {
	tc.context = TypingContext{mod: mod, types: data.NewStack[ast.Type]()}
	menv, err := tc.infer.inferModule(mod)
	return menv, err
}

func (tc *Typechecker) Errors() []data.CompilerProblem {
	return tc.infer.errors
}

func (tc *Typechecker) Env() *Env {
	return tc.env
}

func (tc *Typechecker) instantiate(level ast.Level, typ ast.Type) ast.Type {
	idVarMap := make(map[ast.Id]ast.Type)
	var f func(ast.Type) ast.Type
	f = func(ty ast.Type) ast.Type {
		switch t := ty.(type) {
		case ast.TConst:
			return ty
		case ast.TVar:
			{
				tv := t.Tvar
				if tv.Tag == ast.LINK {
					return f(tv.Type)
				} else if tv.Tag == ast.GENERIC {
					v, found := idVarMap[tv.Id]
					if found {
						return v
					} else {
						va := tc.NewVar(level)
						idVarMap[tv.Id] = va
						return va
					}
				} else { // unbound
					return ty
				}
			}
		case ast.TApp:
			return ast.TApp{Type: f(t.Type), Types: data.MapSlice(t.Types, f), Span: t.Span}
		case ast.TArrow:
			return ast.TArrow{Args: data.MapSlice(t.Args, f), Ret: f(t.Ret), Span: t.Span}
		case ast.TImplicit:
			return ast.TImplicit{Type: f(t.Type), Span: t.Span}
		case ast.TRecord:
			return ast.TRecord{Row: f(t.Row), Span: t.Span}
		case ast.TRowEmpty:
			return ty
		case ast.TRowExtend:
			return ast.TRowExtend{Labels: data.LabelMapValues(t.Labels, f), Row: f(t.Row), Span: t.Span}
		default:
			panic("got unknow type in instantiate")
		}
	}
	return f(typ)
}

func (tc *Typechecker) checkWellFormed(typ ast.Type, span data.Span) error {
	switch t := typ.(type) {
	case ast.TConst:
		{
			envType, found := tc.env.LookupType(t.Name)
			if !found {
				return tc.makeError(data.UndefinedType(t.Name), span)
			}

			if t.Kind != envType.GetKind() {
				return tc.makeError(data.WrongKind(t.Kind.String(), envType.GetKind().String()), span)
			}
		}
	case ast.TApp:
		{
			err := tc.checkWellFormed(t.Type, span)
			if err != nil {
				return err
			}
			for _, ty := range t.Types {
				err = tc.checkWellFormed(ty, span)
				if err != nil {
					return err
				}
			}
		}
	case ast.TArrow:
		{
			err := tc.checkWellFormed(t.Ret, span)
			if err != nil {
				return err
			}
			for _, ty := range t.Args {
				err = tc.checkWellFormed(ty, span)
				if err != nil {
					return err
				}
			}
		}
	case ast.TVar:
		{
			tv := t.Tvar
			if tv.Tag == ast.LINK {
				return tc.checkWellFormed(tv.Type, span)
			}
			if tv.Tag == ast.UNBOUND {
				return tc.makeError(data.UnusedVariable(t.String()), span)
			}
		}
	case ast.TRecord:
		return tc.checkWellFormed(t.Row, span)
	case ast.TRowExtend:
		{
			err := tc.checkWellFormed(t.Row, span)
			if err != nil {
				return err
			}
			for _, ty := range t.Labels.Values() {
				err = tc.checkWellFormed(ty, span)
				if err != nil {
					return err
				}
			}
		}
	case ast.TImplicit:
		return tc.checkWellFormed(t.Type, span)
	}
	return nil
}

func (tc *Typechecker) makeError(msg string, span data.Span) data.CompilerProblem {
	mod := tc.context.mod
	return data.CompilerProblem{
		Msg:           msg,
		Span:          span,
		Filename:      mod.SourceName,
		Module:        mod.Name.Val,
		Severity:      data.ERROR,
		TypingContext: formatTypingContext(tc.context),
	}
}

func (tc *Typechecker) makeErrorRef(msg string, span data.Span) *data.CompilerProblem {
	err := tc.makeError(msg, span)
	return &err
}

type TypingContext struct {
	mod   ast.Module
	decl  *ast.ValDecl
	types *data.Stack[ast.Type]
}

func formatTypingContext(ctx TypingContext) string {
	var sb strings.Builder
	if !ctx.types.IsEmpty() {
		typ := ctx.types.Peek()
		sb.WriteString("while checking type ")
		sb.WriteString(typ.String())
		sb.WriteString("\n")
	}
	decl := ctx.decl
	if decl != nil {
		sb.WriteString("in declaration ")
		sb.WriteString(decl.Name.Val)
	}
	str := sb.String()
	if str != "" {
		return fmt.Sprintf("%s\n\n", str)
	}
	return str
}
