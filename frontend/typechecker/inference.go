package typechecker

import (
	"fmt"
	"math"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
)

type Inference struct {
	tc       *Typechecker
	errors   []data.CompilerProblem
	env      *Env
	uni      *Unification
	pvtTypes data.Set[string]
	mod      ast.Module
}

func NewInference(tc *Typechecker, uni *Unification) *Inference {
	return &Inference{
		tc:       tc,
		errors:   make([]data.CompilerProblem, 0),
		env:      NewEnv(),
		uni:      uni,
		pvtTypes: data.NewSet[string](),
	}
}

// Infer the whole module.
// If error != nil a fatal error ocurred.
// call Errors() to get all errors
func (i *Inference) inferModule(mod ast.Module) (ModuleEnv, error) {
	i.mod = mod
	i.errors = make([]data.CompilerProblem, 0)
	decls := make(map[string]DeclRef)
	types := make(map[string]TypeDeclRef)

	vvar := i.tc.NewGenVar()
	i.env.Extend("__fix", ast.TArrow{Args: []ast.Type{ast.TArrow{Args: []ast.Type{vvar}, Ret: vvar}}, Ret: vvar})

	datas := data.FilterSliceIsInstance[ast.Decl, ast.TypeDecl](mod.Decls)
	for _, d := range datas {
		ty, m := i.getDataType(d, mod.Name.Val)
		err := i.checkShadowType(i.env, d.Name.Val, d.Span)
		if err != nil {
			return ModuleEnv{}, err
		}
		typeName := fmt.Sprintf("%s.%s", mod.Name.Val, d.Name.Val)
		i.env.ExtendType(typeName, ty)

		if d.Visibility == ast.PRIVATE {
			i.pvtTypes.Add(typeName)
		}

		ctorNames := make([]string, 0, len(d.DataCtors))
		for _, dc := range d.DataCtors {
			dcname := dc.Name.Val
			ctorNames = append(ctorNames, dcname)
			dcty := getCtorType(dc, ty, m)
			err := i.checkShadow(i.env, dcname, dc.Span)
			if err != nil {
				return ModuleEnv{}, err
			}
			i.env.Extend(dcname, dcty)
			// TODO: cache constructor
			decls[dcname] = DeclRef{Type: dcty, Visibility: dc.Visibility, IsInstance: false, Comment: nil}
		}
		types[d.Name.Val] = TypeDeclRef{Type: ty, Visibility: d.Visibility, Ctors: ctorNames, Comment: d.Comment}
	}
	for _, d := range datas {
		for _, dc := range d.DataCtors {
			name, _ := i.env.Lookup(dc.Name.Val)
			err := i.tc.checkWellFormed(name, dc.Span)
			if err != nil {
				return ModuleEnv{}, err
			}
		}
	}

	// TODO: metadata

	vals := data.FilterSliceIsInstance[ast.Decl, ast.ValDecl](mod.Decls)
	for _, val := range vals {
		if ann, isAnn := val.Exp.(ast.Ann); isAnn {
			err := i.checkShadow(i.env, val.Name.Val, val.Span)
			if err != nil {
				return ModuleEnv{}, err
			}
			i.env.Extend(val.Name.Val, ann.AnnType)
			if val.IsInstance {
				i.env.ExtendInstance(val.Name.Val, ann.AnnType, false)
			}
		}
	}

	for _, decl := range vals {
		// TODO: implicits
		i.tc.context.decl = &decl
		name := decl.Name.Val
		_, isAnnotated := decl.Exp.(ast.Ann)
		if !isAnnotated {
			err := i.checkShadow(i.env, name, decl.Span)
			if err != nil {
				i.addError(err)
				continue
			}
		}

		newEnv := i.env.Fork()
		var ty ast.Type
		var err *data.CompilerProblem
		if decl.Recursive {
			newEnv.Remove(name)
			ty, err = i.inferRecursive(name, decl.Exp, newEnv, 0)
			if err != nil {
				i.addError(err)
				continue
			}
		} else {
			ty, err = i.infer(newEnv, 0, decl.Exp)
			if err != nil {
				i.addError(err)
				continue
			}
		}

		// TODO: check implicits
		genTy := i.generalize(-1, ty)
		i.env.Extend(name, genTy)
		if decl.IsInstance {
			i.env.ExtendInstance(name, genTy, false)
		}
		decls[name] = DeclRef{Type: genTy, Visibility: decl.Visibility, IsInstance: decl.IsInstance, Comment: decl.Comment}

		if decl.Visibility == ast.PUBLIC && i.pvtTypes.Size() != 0 {
			err = i.checkEscapePvtType(genTy, decl.Name.Span)
			if err != nil {
				i.addError(err)
				continue
			}
		}

		// TODO: check warnings
	}

	return ModuleEnv{Decls: decls, Types: types}, nil
}

func (i *Inference) infer(env *Env, level ast.Level, expr ast.Expr) (ast.Type, *data.CompilerProblem) {
	switch e := expr.(type) {
	case ast.Int:
		return e.WithType(tInt), nil
	case ast.Float:
		return e.WithType(tFloat32), nil
	case ast.Complex:
		return e.WithType(tComplex64), nil
	case ast.Char:
		return e.WithType(tRune), nil
	case ast.String:
		return e.WithType(tString), nil
	case ast.Bool:
		return e.WithType(tBool), nil
	case ast.Unit:
		return e.WithType(tUnit), nil
	case ast.Nil:
		{
			// TODO
			panic("nil not yet supported")
		}
	case ast.Var:
		{
			ty, found := env.Lookup(e.Fullname())
			if found {
				return e.WithType(i.tc.instantiate(level, ty)), nil
			}
			return ast.TConst{}, i.tc.makeErrorRef(data.UndefinedVar(e.Name), e.Span)
		}
	case ast.Ctor:
		{
			ty, found := env.Lookup(e.Fullname())
			if found {
				return e.WithType(i.tc.instantiate(level, ty)), nil
			}
			return ast.TConst{}, i.tc.makeErrorRef(data.UndefinedVar(e.Name), e.Span)
		}
	case ast.ImplicitVar:
		{
			ty, found := env.Lookup(e.Fullname())
			if found {
				return e.WithType(i.tc.instantiate(level, ty)), nil
			}
			return ast.TConst{}, i.tc.makeErrorRef(data.UndefinedVar(e.Name), e.Span)
		}
	case ast.Lambda:
		{
			bind := e.Binder
			err := i.checkShadow(env, bind.Name, bind.Span)
			if err != nil {
				return ast.TConst{}, err
			}
			// if the binder is annotated, use it
			var par ast.Type
			if bind.Type != nil {
				par = *bind.Type
			} else {
				par = i.tc.NewVar(level)
			}
			param := par
			if bind.IsImplicit {
				param = ast.TImplicit{Type: par}
			}
			newEnv := env.Fork()
			newEnv.Extend(bind.Name, param)
			if bind.IsImplicit {
				newEnv.ExtendInstance(bind.Name, param, true)
			}
			returnTy, err := i.infer(newEnv, level, e.Body)
			if err != nil {
				return ast.TConst{}, err
			}
			ty := ast.TArrow{Args: []ast.Type{param}, Ret: returnTy}
			return e.WithType(ty), nil
		}
	case ast.Let:
		{
			name := e.Def.Binder.Name
			err := i.checkShadow(env, name, e.Def.Binder.Span)
			if err != nil {
				return ast.TConst{}, err
			}
			var varTy ast.Type
			if e.Def.Recursive {
				varTy, err = i.inferRecursive(name, e.Def.Expr, env, level+1)
			} else {
				varTy, err = i.infer(env, level+1, e.Def.Expr)
			}
			if err != nil {
				return ast.TConst{}, err
			}

			_, isArr := ast.RealType(varTy).(ast.TArrow)
			if e.Def.Recursive && !isArr {
				return ast.TConst{}, i.tc.makeErrorRef(data.RECURSIVE_LET, e.Def.Binder.Span)
			}

			genTy := i.generalize(level, varTy)
			newEnv := env.Fork()
			newEnv.Extend(name, genTy)
			if e.Def.IsInstance {
				newEnv.ExtendInstance(name, genTy, false)
			}
			ty, err := i.infer(newEnv, level, e.Body)
			if err != nil {
				return ast.TConst{}, err
			}
			return e.WithType(ty), nil
		}
	case ast.App:
		{
			retTy, err := i.infer(env, level, e.Fn)
			if err != nil {
				return nil, err
			}
			pars, retTy, err := i.matchFunType(1, retTy, e.Fn.GetSpan())
			if err != nil {
				return nil, err
			}
			argTy, err := i.infer(env, level, e.Arg)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(pars[0], argTy, e.Arg.GetSpan())
			if err != nil {
				return nil, err
			}
			return e.WithType(retTy), nil
		}
	case ast.Ann:
		{
			i.tc.context.types.Push(e.AnnType)
			typ := e.AnnType
			exp := e.Exp
			// this is a little `checking mode` for some base conversions
			// TODO: list and set literals
			var restTy ast.Type
			ii, isInt := exp.(ast.Int)
			if isInt && typ.Equals(tByte) && validInt8(ii) {
				restTy = tByte
			} else if isInt && typ.Equals(tInt8) && validInt8(ii) {
				restTy = tInt8
			} else if isInt && typ.Equals(tInt16) && validInt16(ii) {
				restTy = tInt16
			} else if isInt && typ.Equals(tInt32) && validInt32(ii) {
				restTy = tInt32
			} else if isInt && typ.Equals(tInt64) {
				restTy = tInt64
			} else if isInt && typ.Equals(tUint) && validUint(ii) {
				restTy = tUint
			} else if isInt && typ.Equals(tUint8) && validUint8(ii) {
				restTy = tUint8
			} else if isInt && typ.Equals(tUint16) && validUint16(ii) {
				restTy = tUint16
			} else if isInt && typ.Equals(tUint32) && validUint32(ii) {
				restTy = tUint32
			} else if isInt && typ.Equals(tUint64) {
				restTy = tUint64
			} else if f, isFloat := exp.(ast.Float); isFloat && typ.Equals(tFloat32) && validFloat32(f) {
				restTy = tFloat32
			} else {
				err := i.validateType(typ, env, exp.GetSpan())
				if err != nil {
					return nil, err
				}
				uniTy, err := i.infer(env, level, exp)
				if err != nil {
					return nil, err
				}
				err = i.uni.Unify(typ, uniTy, exp.GetSpan())
				if err != nil {
					return nil, err
				}
				restTy = typ
			}
			i.tc.context.types.Pop()
			exp.WithType(restTy)
			return e.WithType(restTy), nil
		}
	case ast.If:
		{
			condTy, err := i.infer(env, level, e.Cond)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(tBool, condTy, e.Cond.GetSpan())
			if err != nil {
				return nil, err
			}
			thenTy, err := i.infer(env, level, e.Then)
			if err != nil {
				return nil, err
			}
			elseTy, err := i.infer(env, level, e.Else)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(thenTy, elseTy, e.Span)
			if err != nil {
				return nil, err
			}
			return e.WithType(thenTy), nil
		}
	case ast.Do:
		{
			var ty ast.Type
			var err *data.CompilerProblem
			for _, exp := range e.Exps {
				ty, err = i.infer(env, level, exp)
				if err != nil {
					return nil, err
				}
			}
			return e.WithType(ty), nil
		}
	case ast.Match:
		{
			expsTy := make([]ast.Type, 0, len(e.Exps))
			for _, exp := range e.Exps {
				ty, err := i.infer(env, level, exp)
				if err != nil {
					return nil, err
				}
				expsTy = append(expsTy, ty)
			}
			resTy := i.tc.NewVar(level)

			for _, cas := range e.Cases {
				vars := make([]PatternVar, 0, len(cas.Patterns))
				for j, pat := range cas.Patterns {
					vs, err := i.inferPattern(env, level, pat, expsTy[j])
					if err != nil {
						return nil, err
					}
					vars = append(vars, vs...)
				}

				newEnv := env
				if len(vars) != 0 {
					newEnv = env.Fork()
					for _, v := range vars {
						err := i.checkShadow(newEnv, v.name, v.span)
						if err != nil {
							return nil, err
						}
						newEnv.Extend(v.name, v.typ)
					}
				}

				if cas.Guard != nil {
					ty, err := i.infer(newEnv, level, cas.Guard)
					if err != nil {
						return nil, err
					}
					err = i.uni.Unify(tBool, ty, cas.Guard.GetSpan())
					if err != nil {
						return nil, err
					}
				}

				ty, err := i.infer(newEnv, level, cas.Exp)
				if err != nil {
					return nil, err
				}
				err = i.uni.Unify(resTy, ty, cas.Exp.GetSpan())
				if err != nil {
					return nil, err
				}
			}
			return e.WithType(resTy), nil
		}
	case ast.RecordEmpty:
		return e.WithType(ast.TRecord{Row: ast.TRowEmpty{}}), nil
	case ast.RecordSelect:
		{
			rest := i.tc.NewVar(level)
			field := i.tc.NewVar(level)
			param := ast.TRecord{Row: ast.TRowExtend{Labels: data.LabelMapSingleton(e.Label.Val, field), Row: rest}}
			ty, err := i.infer(env, level, e.Exp)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(param, ty, e.Span)
			if err != nil {
				return nil, err
			}
			return e.WithType(field), nil
		}
	case ast.RecordRestrict:
		{
			rest := i.tc.NewVar(level)
			field := i.tc.NewVar(level)
			param := ast.TRecord{Row: ast.TRowExtend{Labels: data.LabelMapSingleton(e.Label, field), Row: rest}}
			returnTy := ast.TRecord{Row: rest}
			ty, err := i.infer(env, level, e.Exp)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(param, ty, e.Span)
			if err != nil {
				return nil, err
			}
			return e.WithType(returnTy), nil
		}
	case ast.RecordUpdate:
		{
			field, err := i.infer(env, level, e.Value)
			if err != nil {
				return nil, err
			}
			rest := i.tc.NewVar(level)
			var recTy ast.Type
			if e.IsSet {
				recTy = ast.TRecord{Row: ast.TRowExtend{Labels: data.LabelMapSingleton(e.Label.Val, field), Row: rest}}
			} else {
				actualField := i.tc.NewVar(level)
				err = i.uni.Unify(field, ast.TArrow{Args: []ast.Type{actualField}, Ret: actualField}, e.Value.GetSpan())
				if err != nil {
					return nil, err
				}
				recTy = ast.TRecord{Row: ast.TRowExtend{Labels: data.LabelMapSingleton(e.Label.Val, actualField), Row: rest}}
			}
			ty, err := i.infer(env, level, e.Exp)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(recTy, ty, e.Span)
			if err != nil {
				return nil, err
			}
			return e.WithType(recTy), nil
		}
	case ast.RecordExtend:
		{
			labelTys := make([]data.Entry[ast.Type], 0, len(e.Labels.Entries()))
			for _, ent := range e.Labels.Entries() {
				ty, err := i.infer(env, level, ent.Val)
				if err != nil {
					return nil, err
				}
				labelTys = append(labelTys, data.Entry[ast.Type]{Label: ent.Label, Val: ty})
			}

			rest := i.tc.NewVar(level)
			tmp, err := i.infer(env, level, e.Exp)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(ast.TRecord{Row: rest}, tmp, e.Span)
			if err != nil {
				return nil, err
			}
			ty := ast.TRecord{Row: ast.TRowExtend{Labels: data.LabelMapFrom(labelTys...), Row: rest}}
			return e.WithType(ty), nil
		}
	case ast.RecordMerge:
		{
			rest1 := i.tc.NewVar(level)
			rest2 := i.tc.NewVar(level)
			param1 := ast.TRecord{Row: ast.TRowExtend{Labels: data.LabelMapFrom[ast.Type](), Row: rest1}}
			param2 := ast.TRecord{Row: ast.TRowExtend{Labels: data.LabelMapFrom[ast.Type](), Row: rest2}}
			ty1, err := i.infer(env, level, e.Exp1)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(param1, ty1, e.Span)
			if err != nil {
				return nil, err
			}
			ty2, err := i.infer(env, level, e.Exp2)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(param2, ty2, e.Span)
			if err != nil {
				return nil, err
			}

			labels1, row1, err := i.uni.MatchRowType(rest1, e.Exp1.GetSpan())
			if err != nil {
				return nil, err
			}
			labels2, row2, err := i.uni.MatchRowType(rest2, e.Exp2.GetSpan())
			if err != nil {
				return nil, err
			}

			var row ast.Type
			if _, isEmpty := row1.(ast.TRowEmpty); isEmpty {
				row = row2
			} else if _, isEmpty := row2.(ast.TRowEmpty); isEmpty {
				row = row1
			} else {
				return nil, i.tc.makeErrorRef(data.RECORD_MERGE, e.Span)
			}
			ty := ast.TRecord{Row: ast.TRowExtend{Labels: labels2.Merge(labels1), Row: row}}
			return e.WithType(ty), nil
		}
	case ast.ListLiteral:
		{
			ty := i.tc.NewVar(level)
			for _, exp := range e.Exps {
				tt, err := i.infer(env, level, exp)
				if err != nil {
					return nil, err
				}
				err = i.uni.Unify(ty, tt, exp.GetSpan())
				if err != nil {
					return nil, err
				}
			}
			res := ast.TApp{Type: ast.TConst{Name: PrimList}, Types: []ast.Type{ty}}
			return e.WithType(res), nil
		}
	case ast.SetLiteral:
		{
			ty := i.tc.NewVar(level)
			for _, exp := range e.Exps {
				tt, err := i.infer(env, level, exp)
				if err != nil {
					return nil, err
				}
				err = i.uni.Unify(ty, tt, exp.GetSpan())
				if err != nil {
					return nil, err
				}
			}
			res := ast.TApp{Type: ast.TConst{Name: PrimSet}, Types: []ast.Type{ty}}
			return e.WithType(res), nil
		}
	case ast.Index:
		panic("index not yet supported")
	case ast.While:
		{
			typ, err := i.infer(env, level, e.Cond)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(tBool, typ, e.Cond.GetSpan())
			if err != nil {
				return nil, err
			}

			for _, exp := range e.Exps {
				_, err := i.infer(env, level, exp)
				if err != nil {
					return nil, err
				}
			}
			// while always returns unit
			return e.WithType(tUnit), nil
		}
	case ast.TypeCast:
		{
			_, err := i.infer(env, level, e.Exp)
			if err != nil {
				return nil, err
			}
			return e.WithType(e.Cast), nil
		}
	default:
		panic("Got unknow expression in infer")
	}
}

type PatternVar struct {
	name string
	typ  ast.Type
	span data.Span
}

func (i *Inference) inferPattern(env *Env, level ast.Level, pat ast.Pattern, ty ast.Type) ([]PatternVar, *data.CompilerProblem) {
	var res []PatternVar
	switch p := pat.(type) {
	case ast.LiteralP:
		{
			ity, err := i.infer(env, level, p.Lit)
			if err != nil {
				return nil, err
			}
			err = i.uni.Unify(ty, ity, p.Span)
			if err != nil {
				return nil, err
			}
			res = []PatternVar{}
		}
	case ast.Wildcard:
		res = []PatternVar{}
	case ast.UnitP:
		{
			err := i.uni.Unify(ty, tUnit, p.Span)
			if err != nil {
				return nil, err
			}
			res = []PatternVar{}
		}
	case ast.VarP:
		res = []PatternVar{{name: p.V.Name, typ: ty, span: p.GetSpan()}}
	case ast.RegexP:
		{
			err := i.uni.Unify(tString, ty, p.Span)
			if err != nil {
				return nil, err
			}
			res = []PatternVar{}
		}
	case ast.CtorP:
		{
			cty, err := i.infer(env, level, p.Ctor)
			if err != nil {
				return nil, err
			}

			ctorTypes, ret := i.peelArgs([]ast.Type{}, cty)
			err = i.uni.Unify(ret, ty, p.Ctor.Span)
			if err != nil {
				return nil, err
			}

			if len(ctorTypes)-len(p.Fields) != 0 {
				return nil, i.tc.makeErrorRef(data.WrongArityCtorPattern(p.Ctor.Name, len(p.Fields), len(ctorTypes)), p.Span)
			}

			if len(ctorTypes) == 0 {
				res = []PatternVar{}
			} else {
				vars := make([]PatternVar, 0, len(ctorTypes))
				for _, tu := range data.ZipSlices(ctorTypes, p.Fields) {
					pvs, err := i.inferPattern(env, level, tu.V2, tu.V1)
					if err != nil {
						return nil, err
					}
					vars = append(vars, pvs...)
				}
				res = vars
			}
		}
	case ast.RecordP:
		{
			if p.Labels.IsEmpty() {
				err := i.uni.Unify(ast.TRecord{Row: i.tc.NewVar(level)}, ty, p.Span)
				return []PatternVar{}, err
			}

			labels := p.Labels.Entries()
			vars := make([]PatternVar, 0, len(labels))
			tvs := make([]data.Entry[ast.Type], 0, len(labels))
			for _, ent := range labels {
				rowTy := i.tc.NewVar(level)
				pvs, err := i.inferPattern(env, level, ent.Val, rowTy)
				if err != nil {
					return nil, err
				}
				vars = append(vars, pvs...)
				tvs = append(tvs, data.Entry[ast.Type]{Label: ent.Label, Val: rowTy})
			}
			labelMap := data.LabelMapFrom(tvs...)
			err := i.uni.Unify(ast.TRecord{Row: ast.TRowExtend{Labels: labelMap, Row: i.tc.NewVar(level)}}, ty, p.Span)
			if err != nil {
				return nil, err
			}
			res = vars
		}
	case ast.ListP:
		{
			if len(p.Elems) == 0 {
				err := i.uni.Unify(ast.TApp{Type: ast.TConst{Name: PrimList}, Types: []ast.Type{i.tc.NewVar(level)}}, ty, p.Span)
				return []PatternVar{}, err
			}

			elemTy := i.tc.NewVar(level)
			listTy := ast.TApp{Type: ast.TConst{Name: PrimList}, Types: []ast.Type{elemTy}}
			err := i.uni.Unify(listTy, ty, p.Span)
			if err != nil {
				return nil, err
			}

			vars := make([]PatternVar, 0, len(p.Elems))
			for _, pattern := range p.Elems {
				vs, err := i.inferPattern(env, level, pattern, elemTy)
				if err != nil {
					return nil, err
				}
				vars = append(vars, vs...)
			}
			if p.Tail != nil {
				vs, err := i.inferPattern(env, level, p.Tail, listTy)
				if err != nil {
					return nil, err
				}
				vars = append(vars, vs...)
			}
			res = vars
		}
	case ast.NamedP:
		{
			vars, err := i.inferPattern(env, level, p.Pat, ty)
			if err != nil {
				return nil, err
			}
			res = append(vars, PatternVar{name: p.Name.Val, typ: ty, span: p.Span})
		}
	case ast.TypeTest:
		panic("type test not yet supported")
	}
	pat.WithType(ty)
	return res, nil
}

func (i *Inference) matchFunType(numParams int, typ ast.Type, span data.Span) ([]ast.Type, ast.Type, *data.CompilerProblem) {
	switch t := typ.(type) {
	case ast.TArrow:
		{
			if numParams != len(t.Args) {
				panic(fmt.Sprintf("unexpected number of arguments to function: %d", numParams))
			}
			return t.Args, t.Ret, nil
		}
	case ast.TVar:
		if t.Tvar.Tag == ast.LINK {
			return i.matchFunType(numParams, t.Tvar.Type, span)
		} else if t.Tvar.Tag == ast.UNBOUND {
			unb := t.Tvar
			pars := []ast.Type{i.tc.NewVar(unb.Level)}
			ret := i.tc.NewVar(unb.Level)
			t.Tvar.Tag = ast.LINK
			t.Tvar.Type = ast.TArrow{Args: pars, Ret: ret, Span: t.Span}
			return pars, ret, nil
		} else {
			return nil, nil, i.tc.makeErrorRef(data.NOT_A_FUNCTION, span)
		}
	default:
		return nil, nil, i.tc.makeErrorRef(data.NOT_A_FUNCTION, span)
	}
}

func (i *Inference) peelArgs(args []ast.Type, typ ast.Type) ([]ast.Type, ast.Type) {
	switch t := typ.(type) {
	case ast.TArrow:
		{
			if _, isArr := t.Ret.(ast.TArrow); isArr {
				return i.peelArgs(append(args, t.Args...), t.Ret)
			}
			return append(args, t.Args...), t.Ret
		}
	case ast.TVar:
		{
			if t.Tvar.Tag == ast.LINK {
				return i.peelArgs(args, t.Tvar.Type)
			}
			return args, typ
		}
	default:
		return args, typ
	}
}

func (i *Inference) generalize(level ast.Level, typ ast.Type) ast.Type {
	switch t := typ.(type) {
	case ast.TVar:
		{
			tv := t.Tvar
			if tv.Tag == ast.LINK {
				return i.generalize(level, tv.Type)
			} else if tv.Tag == ast.UNBOUND && tv.Level > level {
				return ast.TVar{Tvar: &ast.TypeVar{Tag: ast.GENERIC, Id: tv.Id}}
			} else {
				return typ
			}
		}
	case ast.TApp:
		{
			t.Type = i.generalize(level, t.Type)
			t.Types = data.MapSlice(t.Types, func(ty ast.Type) ast.Type { return i.generalize(level, ty) })
			return t
		}
	case ast.TArrow:
		{
			t.Args = data.MapSlice(t.Args, func(ty ast.Type) ast.Type { return i.generalize(level, ty) })
			t.Ret = i.generalize(level, t.Ret)
			return t
		}
	case ast.TImplicit:
		{
			t.Type = i.generalize(level, t.Type)
			return t
		}
	case ast.TRecord:
		{
			t.Row = i.generalize(level, t.Row)
			return t
		}
	case ast.TRowExtend:
		{
			t.Row = i.generalize(level, t.Row)
			t.Labels = data.LabelMapValues(t.Labels, func(ty ast.Type) ast.Type { return i.generalize(level, ty) })
			return t
		}
	default:
		return typ
	}
}

// Use a fixpoint operator to infer a recursive function
func (i *Inference) inferRecursive(name string, exp ast.Expr, env *Env, level ast.Level) (ast.Type, *data.CompilerProblem) {
	newName, newExp := funToFixpoint(name, exp)
	recTy, err := i.infer(env, level, newExp)
	if err != nil {
		return recTy, err
	}
	env.Extend(newName, recTy)
	fix := ast.App{
		Fn:  ast.Var{Name: "__fix", Span: exp.GetSpan(), Type: &ast.Typed{}},
		Arg: ast.Var{Name: newName, Span: exp.GetSpan(), Type: &ast.Typed{}}, Span: exp.GetSpan(),
		Type: &ast.Typed{},
	}
	ty, err := i.infer(env, level, fix)
	if err != nil {
		return ty, err
	}
	return exp.WithType(ty), nil
}

/*
 Change this recursive expression in a way that
 it can be infered.

 Ex.:

 fun x = fun x

 __fun __funrec x = __funrec x
*/
func funToFixpoint(name string, exp ast.Expr) (string, ast.Expr) {
	binder := "__rec" + name
	return binder, ast.Lambda{Binder: ast.Binder{Name: name, Span: exp.GetSpan()}, Body: exp, Span: exp.GetSpan(), Type: &ast.Typed{}}
}

func (i *Inference) validateType(typ ast.Type, env *Env, span data.Span) *data.CompilerProblem {
	var err *data.CompilerProblem
	ast.EverywhereTypeUnit(typ, func(t ast.Type) {
		if ty, isConst := t.(ast.TConst); isConst {
			if _, inEnv := env.LookupType(ty.Name); inEnv {
				if !ty.Span.IsEmpty() {
					span = ty.Span
				}
				err = i.tc.makeErrorRef(data.UndefinedType(ty.String()), span)
			}
		}
	})
	return err
}

func validInt8(i ast.Int) bool {
	return i.V >= math.MinInt8 && i.V <= math.MaxInt8
}
func validInt16(i ast.Int) bool {
	return i.V >= math.MinInt16 && i.V <= math.MaxInt16
}
func validInt32(i ast.Int) bool {
	return i.V >= math.MinInt32 && i.V <= math.MaxInt32
}
func validUint(i ast.Int) bool {
	if i.V < 0 {
		return false
	}
	var v uint
	v = uint(i.V)
	return v <= math.MaxUint
}
func validUint8(i ast.Int) bool {
	return i.V >= 0 && i.V <= math.MaxUint8
}
func validUint16(i ast.Int) bool {
	return i.V >= 0 && i.V <= math.MaxUint16
}
func validUint32(i ast.Int) bool {
	return i.V >= 0 && i.V <= math.MaxUint32
}
func validFloat32(f ast.Float) bool {
	return f.V >= math.SmallestNonzeroFloat32 && f.V <= math.MaxFloat32
}

// Checks if a public function doesn't have a private type.
// So the type doesn't escape its module.
func (i *Inference) checkEscapePvtType(typ ast.Type, span data.Span) *data.CompilerProblem {
	found := ""
	ast.EverywhereTypeUnit(typ, func(t ast.Type) {
		if tc, isTc := t.(ast.TConst); isTc && i.pvtTypes.Contains(tc.Name) {
			found = tc.Name
			if !tc.Span.IsEmpty() {
				span = tc.Span
			}
		}
	})
	if found != "" {
		return i.tc.makeErrorRef(data.EscapeType(found), span)
	}
	return nil
}

func (i *Inference) addError(err *data.CompilerProblem) {
	i.errors = append(i.errors, *err)
}

// Check if `name` is shadowing some variable
// and throw an error if that's the case.
func (i *Inference) checkShadow(env *Env, name string, span data.Span) *data.CompilerProblem {
	if _, has := env.Lookup(name); has {
		return i.tc.makeErrorRef(data.ShadowedVariable(name), span)
	}
	return nil
}

// Check if `type` is shadowing some other type
// and throw an error if that's the case.
func (i *Inference) checkShadowType(env *Env, typ string, span data.Span) *data.CompilerProblem {
	if _, has := env.LookupType(typ); has {
		return i.tc.makeErrorRef(data.DuplicatedType(typ), span)
	}
	return nil
}

func (i *Inference) getDataType(d ast.TypeDecl, moduleName string) (ast.Type, map[string]ast.Type) {
	kind := ast.Kind{Type: ast.STAR}
	if len(d.TyVars) > 0 {
		kind = ast.Kind{Type: ast.CTOR, Arity: len(d.TyVars)}
	}
	raw := ast.TConst{Name: fmt.Sprintf("%s.%s", moduleName, d.Name.Val), Kind: kind, Span: d.Span}

	if len(d.TyVars) == 0 {
		return raw, make(map[string]ast.Type)
	}
	m := make(map[string]ast.Type)
	vars := make([]ast.Type, 0, len(d.TyVars))
	for _, v := range d.TyVars {
		gv := i.tc.NewGenVar()
		m[v] = gv
		vars = append(vars, gv)
	}
	return ast.TApp{Type: raw, Types: vars, Span: d.Span}, m
}

func getCtorType(dc ast.DataCtor, dataType ast.Type, m map[string]ast.Type) ast.Type {
	switch t := dataType.(type) {
	case ast.TConst:
		if len(dc.Args) == 0 {
			return dataType
		} else {
			arr := ast.NestArrows(dc.Args, dataType)
			return arr.WithSpan(dc.Span)
		}
	case ast.TApp:
		{
			args := data.MapSlice(dc.Args, func(t ast.Type) ast.Type { return ast.SubstConst(t, m) })
			return ast.NestArrows(args, dataType).WithSpan(dc.Span)
		}
	default:
		panic("Got absurd type for data constructor " + t.String())
	}
}
