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
	tc := &Typechecker{
		TypeVarMap: make(map[int]string),
		env:        NewEnv(),
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

func (tc *Typechecker) instantiate(level ast.Level, typ ast.Type) ast.Type {
	idVarMap := make(map[ast.Id]ast.Type)
	var f func(ast.Type) ast.Type
	f = func(typp ast.Type) ast.Type {
		switch t := typp.(type) {
		case ast.TConst:
			return t
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
					return t
				}
			}
		case ast.TApp:
			{
				t.Type = f(t.Type)
				t.Types = data.MapSlice(t.Types, f)
				return t
			}
		case ast.TArrow:
			{
				t.Args = data.MapSlice(t.Args, f)
				t.Ret = f(t.Ret)
				return t
			}
		case ast.TImplicit:
			{
				t.Type = f(t.Type)
				return t
			}
		case ast.TRecord:
			{
				t.Row = f(t.Row)
				return t
			}
		case ast.TRowEmpty:
			return t
		case ast.TRowExtend:
			{
				t.Row = f(t.Row)
				t.Labels = data.LabelMapValues(t.Labels, f)
				return t
			}
		default:
			panic("got unknow type in instantiate")
		}
	}
	return f(typ.Clone())
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
