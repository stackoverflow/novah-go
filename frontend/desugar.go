package frontend

import (
	"errors"
	"fmt"

	"github.com/huandu/go-clone"
	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
	tc "github.com/stackoverflow/novah-go/frontend/typechecker"
	"golang.org/x/exp/slices"
)

type Desugar struct {
	smod ast.SModule
	tc   *tc.Typechecker

	usedVars       data.Set[string]
	unusedVars     map[string]lexer.Span
	usedTypes      data.Set[string]
	usedImports    data.Set[string]
	declNames      data.Set[string]
	declVars       data.Set[string]
	imports        map[string]string
	modName        string
	synonyms       map[string]ast.STypeAliasDecl
	errors         []ast.CompilerProblem
	aliasedImports data.Set[string]
	varCount       int
}

func NewDesugar(smod ast.SModule, tc *tc.Typechecker) *Desugar {
	aliases := data.NewSet[string]()
	for _, imp := range smod.Imports {
		if imp.Alias != nil {
			aliases.Add(*imp.Alias)
		}
	}
	return &Desugar{
		smod:           smod,
		tc:             tc,
		usedVars:       data.NewSet[string](),
		unusedVars:     make(map[string]lexer.Span),
		usedTypes:      data.NewSet[string](),
		usedImports:    data.NewSet[string](),
		imports:        smod.ResolvedImports,
		modName:        smod.Name.Val,
		synonyms:       make(map[string]ast.STypeAliasDecl),
		aliasedImports: aliases,
		varCount:       0,
	}
}

// If error != nil a fatal error ocurred.
// Call Errors to get all errors.
func (d *Desugar) Desugar() (ast.Module, error) {
	d.declNames = data.NewSet[string]()
	for imp := range d.imports {
		d.declNames.Add(imp)
	}

	// TODO: validate type aliases
	desugaredDecls := make([]ast.Decl, 0, len(d.smod.Decls))
	for _, decl := range d.smod.Decls {
		res := d.desugarDecl(decl)
		if res != nil {
			desugaredDecls = append(desugaredDecls, res)
		}
	}
	decls, err := d.validateTopLevelValues(desugaredDecls)
	if err != nil {
		return ast.Module{}, err
	}

	// TODO: report unused imports
	return ast.Module{
		Name:          d.smod.Name,
		SourceName:    d.smod.SourceName,
		Decls:         decls,
		Imports:       d.smod.Imports,
		UnusedImports: make(map[string]lexer.Span),
		Comment:       d.smod.Comment,
	}, nil
}

func (d *Desugar) Errors() []ast.CompilerProblem {
	return d.errors
}

func (d *Desugar) desugarDecl(decl ast.SDecl) ast.Decl {
	switch de := decl.(type) {
	case ast.STypeDecl:
		{
			d.validateDataCtorNames(de)
			if _, imported := d.imports[de.Binder.Val]; imported {
				d.errors = append(d.errors, d.makeError(data.DuplicatedType(de.Binder.Val), de.Span))
				return nil
			} else {
				// TODO: auto derive
				ctors := data.MapSlice(de.DataCtors, d.desugarDataCtor)
				return ast.TypeDecl{Name: de.Binder, TyVars: de.TyVars, DataCtors: ctors, Visibility: de.Visibility, Span: de.Span, Comment: de.Comment}
			}
		}
	case ast.SValDecl:
		{
			name := de.Binder.Val
			if d.declNames.Contains(name) {
				d.errors = append(d.errors, d.makeError(data.DuplicatedDecl(name), de.Span))
				return nil
			}
			d.declNames.Add(name)
			d.declVars = data.NewSet[string]()
			d.checkShadow(name, de.Span)

			d.unusedVars = make(map[string]lexer.Span)
			vars := data.FlatMapSlice(de.Pats, func(t ast.SPattern) []CollectedVar { return d.collectVars(t, false) })
			for _, v := range vars {
				if !v.implicit && !v.instance {
					d.unusedVars[v.name] = v.span
				}
				d.checkShadow(v.name, v.span)
			}

			var stype ast.SType
			if de.Signature != nil {
				stype = de.Signature.Type
			}
			// hold the type variables as scoped typed variables
			typeVars := make(map[string]tc.Type)

			var expType tc.Type
			if stype != nil {
				expType = d.desugarType(stype, false, typeVars)
			}
			var sig *ast.Signature
			if expType != nil {
				sig = &ast.Signature{Type: expType, Span: de.Signature.Span}
			}

			// TODO: spread type annotations on parameters
			exp, err := d.desugarExp(de.Exp, data.NewSet[string](), typeVars)
			if err != nil {
				d.errors = append(d.errors, err.(ast.CompilerProblem))
				return nil
			}
			expr, err2 := d.nestLambdaPats(de.Pats, exp, data.NewSet[string](), typeVars)
			if err2 != nil {
				d.errors = append(d.errors, err2.(ast.CompilerProblem))
				return nil
			}

			if expType != nil {
				expr = ast.Ann{Exp: expr, AnnType: expType, Span: expr.GetSpan()}
			}

			if len(d.unusedVars) > 0 {
				d.addUnusedVars()
			}
			return ast.ValDecl{
				Name:       de.Binder,
				Exp:        expr,
				Recursive:  d.declVars.Contains(name),
				Span:       de.Span,
				Signature:  sig,
				Visibility: de.Visibility,
				IsInstance: de.IsInstance,
				IsOperator: de.IsOperator,
				Comment:    de.Comment,
			}
		}
	default:
		return nil
	}
}

func (d *Desugar) desugarDataCtor(ctor ast.SDataCtor) ast.DataCtor {
	return ast.DataCtor{
		Name:       ctor.Name,
		Args:       data.MapSlice(ctor.Args, func(t ast.SType) tc.Type { return d.desugarType(t, true, make(map[string]tc.Type)) }),
		Visibility: ctor.Visibility,
		Span:       ctor.Span,
	}
}

func (d *Desugar) desugarExp(sexp ast.SExpr, locals data.Set[string], tvars map[string]tc.Type) (ast.Expr, error) {
	switch e := sexp.(type) {
	case ast.SInt:
		return ast.Int{V: e.V, Span: e.Span}, nil
	case ast.SFloat:
		return ast.Float{V: e.V, Span: e.Span}, nil
	case ast.SComplex:
		return ast.Complex{V: e.V, Span: e.Span}, nil
	case ast.SBool:
		return ast.Bool{V: e.V, Span: e.Span}, nil
	case ast.SChar:
		return ast.Char{V: e.V, Span: e.Span}, nil
	case ast.SString:
		return ast.String{V: e.V, Span: e.Span}, nil
	case ast.SPatternLiteral:
		panic("pattern literals not supported yet")
	case ast.SVar:
		{
			d.declVars.Add(e.Fullname())
			if e.Alias == nil {
				delete(d.unusedVars, e.Name)
				d.usedVars.Add(e.Name)
			}
			if e.Alias == nil && locals.Contains(e.Name) {
				return ast.Var{Name: e.Name, Span: e.Span}, nil
			} else {
				if e.Alias != nil {
					d.checkAlias(*e.Alias, e.Span)
				}
				importedModule, has := d.imports[e.Fullname()]
				if has {
					d.usedImports.Add(importedModule)
				}
				return ast.Var{Name: e.Name, Span: e.Span, ModuleName: &importedModule}, nil
			}
		}
	case ast.SImplicitVar:
		{
			if e.Alias == nil {
				delete(d.unusedVars, e.Name)
				d.usedVars.Add(e.Name)
			}
			if e.Alias != nil {
				d.checkAlias(*e.Alias, e.Span)
			}
			importedModule, has := d.imports[e.Fullname()]
			if has {
				d.usedImports.Add(importedModule)
			}
			mname := &importedModule
			if locals.Contains(e.Name) {
				mname = nil
			}
			return ast.ImplicitVar{Name: e.Name, Span: e.Span, ModuleName: mname}, nil
		}
	case ast.SOperator:
		{
			d.declVars.Add(e.Fullname())
			exp := e
			if e.Name == ";" {
				exp.Name = "Tuple"
			}
			if exp.Name == "<-" {
				return nil, d.makeError(data.NOT_A_FIELD, e.Span)
			}
			if exp.Alias == nil {
				delete(d.unusedVars, exp.Name)
				d.usedVars.Add(exp.Name)
			}
			if exp.Alias != nil {
				d.checkAlias(*exp.Alias, e.Span)
			}
			importedModule, has := d.imports[exp.Fullname()]
			if has {
				d.usedImports.Add(importedModule)
			}
			if lexer.IsUpper(exp.Name) {
				return ast.Ctor{Name: exp.Name, Span: exp.Span, ModuleName: &importedModule}, nil
			} else {
				return ast.Var{Name: exp.Name, Span: exp.Span, ModuleName: &importedModule, IsOp: true}, nil
			}
		}
	case ast.SConstructor:
		{
			if e.Alias == nil {
				d.usedVars.Add(e.Name)
			}
			if e.Alias != nil {
				d.checkAlias(*e.Alias, e.Span)
			}
			importedModule, has := d.imports[e.Fullname()]
			if has {
				d.usedImports.Add(importedModule)
			}
			return ast.Ctor{Name: e.Name, Span: e.Span, ModuleName: &importedModule}, nil
		}
	case ast.SLambda:
		{
			vars := data.FlatMapSlice(e.Pats, func(t ast.SPattern) []CollectedVar { return d.collectVars(t, false) })
			for _, v := range vars {
				if !v.implicit && !v.instance {
					d.unusedVars[v.name] = v.span
				}
				d.checkShadow(v.name, v.span)
			}
			newlocals := locals.Copy()
			for _, v := range vars {
				newlocals.Add(v.name)
			}
			body, err := d.desugarExp(e.Body, newlocals, tvars)
			if err != nil {
				return nil, err
			}
			return d.nestLambdaPats(e.Pats, body, locals, tvars)
		}
	case ast.SApp:
		{
			fn, err := d.desugarExp(e.Fn, locals, tvars)
			if err != nil {
				return nil, err
			}
			arg, err2 := d.desugarExp(e.Arg, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return ast.App{Fn: fn, Arg: arg, Span: e.Span}, nil
		}
	case ast.SParens:
		return d.desugarExp(e.Exp, locals, tvars)
	case ast.SIf:
		{
			els := e.Else
			if els == nil {
				els = ast.SUnit{}
			}
			// TODO: nest lambdas
			cond, err := d.desugarExp(e.Cond, locals, tvars)
			if err != nil {
				return nil, err
			}
			then, err2 := d.desugarExp(e.Then, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			_else, err3 := d.desugarExp(e.Else, locals, tvars)
			if err3 != nil {
				return nil, err3
			}
			return ast.If{Cond: cond, Then: then, Else: _else, Span: e.Span}, nil
		}
	case ast.SLet:
		{
			vars := d.collectVarsLdef(e.Def)
			for _, v := range vars {
				if !v.implicit && !v.instance {
					d.unusedVars[v.name] = v.span
				}
				d.checkShadow(v.name, v.span)
			}
			newlocals := locals.Copy()
			for _, v := range vars {
				newlocals.Add(v.name)
			}
			body, err := d.desugarExp(e.Body, newlocals, tvars)
			if err != nil {
				return nil, err
			}
			return d.nestLets(e.Def, body, e.Span, locals, tvars)
		}
	case ast.SMatch:
		{
			// TODO: nest lambdas
			exps, err := data.MapSliceError(e.Exprs, func(t ast.SExpr) (ast.Expr, error) { return d.desugarExp(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			cases, err2 := data.MapSliceError(e.Cases, func(t ast.SCase) (ast.Case, error) { return d.desugarCase(t, locals, tvars) })
			if err2 != nil {
				return nil, err2
			}
			return ast.Match{Exps: exps, Cases: cases, Span: e.Span}, nil
		}
	case ast.SAnn:
		{
			exp, err := d.desugarExp(e.Exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			vars := clone.Clone(tvars).(map[string]tc.Type)
			typ := d.desugarType(e.Type, false, vars)
			return ast.Ann{Exp: exp, Type: typ, Span: e.Span}, nil
		}
	case ast.SDo:
		{
			if exp, isDoLet := e.Exps[len(e.Exps)-1].(ast.SDoLet); isDoLet {
				return nil, d.makeError(data.LET_DO_LAST, exp.Span)
			}
			converted := d.convertDoLets(e.Exps)
			exps, err := data.MapSliceError(converted, func(t ast.SExpr) (ast.Expr, error) { return d.desugarExp(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			return ast.Do{Exps: exps, Span: e.Span}, nil
		}
	case ast.SDoLet:
		return nil, d.makeError(data.LET_IN, e.Span)
	case ast.SUnit:
		return ast.Unit{Span: e.Span}, nil
	case ast.SRecordEmpty:
		return ast.RecordEmpty{Span: e.Span}, nil
	case ast.SRecordSelect:
		{
			// TODO: nest lambdas
			exp, err := d.desugarExp(e.Exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			return d.nestRecordSelects(exp, e.Labels, e.Span), nil
		}
	case ast.SRecordExtend:
		{
			// TODO: nest lambdas
			exp, err := d.desugarExp(e.Exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			labels, err2 := data.LabelMapValuesErr(e.Labels, func(t ast.SExpr) (ast.Expr, error) { return d.desugarExp(t, locals, tvars) })
			if err2 != nil {
				return nil, err2
			}
			return ast.RecordExtend{Labels: labels, Exp: exp, Span: e.Span}, nil
		}
	case ast.SRecordRestrict:
		{
			// TODO: nest lambdas
			exp, err := d.desugarExp(e.Exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			return d.nestRecordRestrictions(exp, e.Labels, e.Span), nil
		}
	case ast.SRecordUpdate:
		{
			// TODO: nest lambdas
			exp, err := d.desugarExp(e.Exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			val, err2 := d.desugarExp(e.Val, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return d.nestRecordUpdates(exp, e.Labels, val, e.IsSet, e.Span), nil
		}
	case ast.SRecordMerge:
		{
			// TODO: nest lambdas
			exp1, err := d.desugarExp(e.Exp1, locals, tvars)
			if err != nil {
				return nil, err
			}
			exp2, err2 := d.desugarExp(e.Exp2, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return ast.RecordMerge{Exp1: exp1, Exp2: exp2, Span: e.Span}, nil
		}
	case ast.SListLiteral:
		{
			// TODO: range literals
			exps, err := data.MapSliceError(e.Exps, func(t ast.SExpr) (ast.Expr, error) { return d.desugarExp(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			return ast.ListLiteral{Exps: exps, Span: e.Span}, nil
		}
	case ast.SSetLiteral:
		{
			// TODO: range literals
			exps, err := data.MapSliceError(e.Exps, func(t ast.SExpr) (ast.Expr, error) { return d.desugarExp(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			return ast.SetLiteral{Exps: exps, Span: e.Span}, nil
		}
	case ast.SIndex:
		{
			// TODO: nest lambdas
			exp, err := d.desugarExp(e.Exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			index, err2 := d.desugarExp(e.Index, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return ast.Index{Exp: exp, Index: index, Span: e.Span}, nil
		}
	case ast.SBinApp:
		if op, isOp := e.Op.(ast.SOperator); isOp && op.Name == "<-" {
			// TODO: Go field setter
			panic("setter not yet implemented")
		} else {
			// TODO: nest lambdas
			left, err := d.desugarExp(e.Left, locals, tvars)
			if err != nil {
				return nil, err
			}
			right, err2 := d.desugarExp(e.Right, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			op, err3 := d.desugarExp(e.Op, locals, tvars)
			if err3 != nil {
				return nil, err3
			}
			inner := ast.App{Fn: op, Arg: left, Span: lexer.NewSpan(left.GetSpan(), op.GetSpan())}
			return ast.App{Fn: inner, Arg: right, Span: lexer.NewSpan(inner.Span, right.GetSpan())}, nil
		}
	case ast.SUnderscore:
		return nil, d.makeError(data.ANONYMOUS_FUNCTION_ARGUMENT, e.Span)
	case ast.SWhile:
		{
			if exp, isDoLet := e.Exps[len(e.Exps)-1].(ast.SDoLet); isDoLet {
				return nil, d.makeError(data.LET_DO_LAST, exp.Span)
			}
			converted := d.convertDoLets(e.Exps)
			exps, err := data.MapSliceError(converted, func(t ast.SExpr) (ast.Expr, error) { return d.desugarExp(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			cond, err2 := d.desugarExp(e.Cond, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return ast.While{Cond: cond, Exps: exps, Span: e.Span}, nil
		}
	case ast.SComputation:
		// TODO: computation
		panic("Computation not supported yet")
	case ast.SNil:
		return ast.Nil{Span: e.Span}, nil
	case ast.STypeCast:
		{
			exp, err := d.desugarExp(e.Exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			vars := clone.Clone(tvars).(map[string]tc.Type)
			typ := d.desugarType(e.Cast, false, vars)
			return ast.TypeCast{Exp: exp, Cast: typ, Span: e.Span}, nil
		}
	case ast.SReturn:
		return nil, d.makeError(data.RETURN_EXPR, e.Span)
	case ast.SYield:
		return nil, d.makeError(data.YIELD_EXPR, e.Span)
	case ast.SLetBang:
		return nil, d.makeError(data.LET_BANG, e.Span)
	case ast.SDoBang:
		return nil, d.makeError(data.DO_BANG, e.Span)
	case ast.SFor:
		return nil, d.makeError(data.FOR_EXPR, e.Span)
	default:
		panic("Got unknow expression in desugarExpr: " + e.String())
	}
}

func (d *Desugar) desugarDefBind(bind ast.SLetBind, locals data.Set[string], tvars map[string]tc.Type) (ast.LetDef, error) {
	e, err := d.desugarExp(bind.Expr, locals, tvars)
	if err != nil {
		return ast.LetDef{}, err
	}
	vars := collectVars(e)
	recursive := slices.Contains(vars, bind.Name.Name)
	if len(bind.Pats) == 0 {
		if bind.Type != nil {
			e = ast.Ann{Exp: e, Type: d.desugarTypeDef(bind.Type), Span: bind.Expr.GetSpan()}
		}
		return ast.LetDef{Binder: d.desugarBinder(bind.Name), Expr: e, Recursive: recursive, IsInstance: bind.IsInstance}, nil
	} else {
		binder, err2 := d.nestLambdaPats(bind.Pats, e, locals, tvars)
		if err2 != nil {
			return ast.LetDef{}, err2
		}
		if bind.Type != nil {
			binder = ast.Ann{Exp: binder, Type: d.desugarTypeDef(bind.Type), Span: bind.Expr.GetSpan()}
		}
		return ast.LetDef{Binder: d.desugarBinder(bind.Name), Expr: binder, Recursive: recursive, IsInstance: bind.IsInstance}, nil
	}
}

func (d *Desugar) desugarBinder(b ast.SBinder) ast.Binder {
	return ast.Binder{Name: b.Name, Span: b.Span, IsImplicit: b.IsImplicit}
}

func (d *Desugar) desugarCase(cas ast.SCase, locals data.Set[string], tvars map[string]tc.Type) (ast.Case, error) {
	vars := data.FlatMapSlice(cas.Pats, func(t ast.SPattern) []CollectedVar { return d.collectVars(t, false) })
	for _, v := range vars {
		if !v.implicit && !v.instance {
			d.unusedVars[v.name] = v.span
		}
		d.checkShadow(v.name, v.span)
	}

	pats, err := data.MapSliceError(cas.Pats, func(t ast.SPattern) (ast.Pattern, error) { return d.desugarPattern(t, locals, tvars) })
	if err != nil {
		return ast.Case{}, err
	}
	exp, err2 := d.desugarExp(cas.Exp, locals, tvars)
	if err2 != nil {
		return ast.Case{}, err2
	}
	var guard ast.Expr
	if cas.Guard != nil {
		guard, err = d.desugarExp(cas.Guard, locals, tvars)
	}
	return ast.Case{Patterns: pats, Exp: exp, Guard: guard}, nil
}

func (d *Desugar) desugarPattern(sp ast.SPattern, locals data.Set[string], tvars map[string]tc.Type) (ast.Pattern, error) {
	switch pat := sp.(type) {
	case ast.SWildcard:
		return ast.Wildcard{Span: pat.Span}, nil
	case ast.SLiteralP:
		{
			lit, err := d.desugarExp(pat.Lit, locals, tvars)
			if err != nil {
				return nil, err
			}
			return ast.LiteralP{Lit: lit, Span: pat.Span}, nil
		}
	case ast.SVarP:
		return ast.VarP{V: ast.Var{Name: pat.V.Name, Span: pat.V.Span}}, nil
	case ast.SCtorP:
		{
			fields, err := data.MapSliceError(pat.Fields, func(t ast.SPattern) (ast.Pattern, error) { return d.desugarPattern(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			ctor, err2 := d.desugarExp(pat.Ctor, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return ast.CtorP{Ctor: ctor.(ast.Ctor), Fields: fields, Span: pat.Span}, nil
		}
	case ast.SParensP:
		return d.desugarPattern(pat.Pat, locals, tvars)
	case ast.SRecordP:
		{
			labels, err := data.LabelMapValuesErr(pat.Labels, func(t ast.SPattern) (ast.Pattern, error) { return d.desugarPattern(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			return ast.RecordP{Labels: labels, Span: pat.Span}, nil
		}
	case ast.SListP:
		{
			var tail ast.Pattern
			var err error
			if pat.Tail != nil {
				tail, err = d.desugarPattern(pat.Tail, locals, tvars)
				if err != nil {
					return nil, err
				}
			}
			elems, err2 := data.MapSliceError(pat.Elems, func(t ast.SPattern) (ast.Pattern, error) { return d.desugarPattern(t, locals, tvars) })
			if err2 != nil {
				return nil, err2
			}
			return ast.ListP{Elems: elems, Tail: tail, Span: pat.Span}, nil
		}
	case ast.SNamed:
		{
			patt, err := d.desugarPattern(pat.Pat, locals, tvars)
			if err != nil {
				return nil, err
			}
			return ast.NamedP{Pat: patt, Name: pat.Name, Span: pat.Span}, nil
		}
	case ast.SUnitP:
		return ast.UnitP{Span: pat.Span}, nil
	case ast.STypeTest:
		return ast.TypeTest{Test: d.desugarTypeDef(pat.Type), Alias: pat.Alias, Span: pat.Span}, nil
	case ast.STupleP:
		{
			p1, err1 := d.desugarPattern(pat.P1, locals, tvars)
			if err1 != nil {
				return nil, err1
			}
			p2, err2 := d.desugarPattern(pat.P2, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			mname := "novah.core"
			ctor := ast.Ctor{Name: "Tuple", Span: pat.Span, ModuleName: &mname}
			return ast.CtorP{Ctor: ctor, Fields: []ast.Pattern{p1, p2}, Span: pat.Span}, nil
		}
	case ast.SRegexP:
		return ast.RegexP{Regex: pat.Regex.Regex, Span: pat.Regex.Span}, nil
	case ast.SImplicitP:
		return nil, d.makeError(data.IMPLICIT_PATTERN, pat.Span)
	case ast.STypeAnnotationP:
		return nil, d.makeError(data.ANNOTATION_PATTERN, pat.Span)
	default:
		panic("Unknow pattern in desugarPattern: " + pat.String())
	}
}

func (d *Desugar) desugarTypeDef(ty ast.SType) tc.Type {
	return d.desugarType(ty, false, make(map[string]tc.Type))
}

func (d *Desugar) desugarType(ty ast.SType, isCtor bool, vars map[string]tc.Type) tc.Type {
	typ := d.resolveAliases(ty)
	return d.goDesugarType(typ, isCtor, vars, 0)
}

func (d *Desugar) goDesugarType(ty ast.SType, isCtor bool, vars map[string]tc.Type, kindArity int) tc.Type {
	switch t := ty.(type) {
	case ast.STConst:
		{
			d.usedTypes.Add(t.Name)
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
					d.usedImports.Add(modName)
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
			labels := data.LabelMapValues(t.Labels, func(tt ast.SType) tc.Type { return d.goDesugarType(tt, isCtor, vars, 0) })
			return tc.TRowExtend{Labels: labels, Row: d.goDesugarType(t.Row, isCtor, vars, 0), Span: t.Span}
		}
	case ast.STImplicit:
		return tc.TImplicit{Type: d.goDesugarType(t.Type, isCtor, vars, 0), Span: t.Span}
	default:
		panic("unknow type in desugaring: " + ty.String())
	}
}

func (d *Desugar) nestLambdaPats(pats []ast.SPattern, exp ast.Expr, locals data.Set[string], tvars map[string]tc.Type) (ast.Expr, error) {
	if len(pats) <= 0 {
		return exp, nil
	}
	switch pat := pats[0].(type) {
	case ast.SVarP:
		{
			body, err := d.nestLambdaPats(pats[1:], exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			return ast.Lambda{
				Binder: ast.Binder{Name: pat.V.Name, Span: pat.V.Span},
				Body:   body,
				Span:   lexer.NewSpan(pat.GetSpan(), exp.GetSpan()),
			}, nil
		}
	case ast.SImplicitP:
		if p, isVar := pat.Pat.(ast.SVarP); isVar {
			body, err := d.nestLambdaPats(pats[1:], exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			return ast.Lambda{
				Binder: ast.Binder{Name: p.V.Name, Span: p.GetSpan(), IsImplicit: true},
				Body:   body,
				Span:   lexer.NewSpan(pat.GetSpan(), exp.GetSpan()),
			}, nil
		} else {
			name := d.newVar()
			vars := []ast.Expr{ast.Var{Name: name, Span: pat.Span}}
			pattern, err := d.desugarPattern(pat.Pat, locals, tvars)
			if err != nil {
				return nil, err
			}
			expr := ast.Match{Exps: vars, Cases: []ast.Case{{Patterns: []ast.Pattern{pattern}, Exp: exp}}, Span: exp.GetSpan()}
			body, err2 := d.nestLambdaPats(pats[1:], expr, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return ast.Lambda{
				Binder: ast.Binder{Name: name, Span: pat.Span, IsImplicit: true},
				Body:   body,
				Span:   lexer.NewSpan(pat.Span, exp.GetSpan()),
			}, nil
		}
	case ast.SParensP:
		{
			pats[0] = pat.Pat
			return d.nestLambdaPats(pats, exp, locals, tvars)
		}
	case ast.STypeAnnotationP:
		{
			bind := ast.Binder{Name: pat.Par.Name, Span: pat.Par.Span}
			vars := clone.Clone(tvars).(map[string]tc.Type)
			bind.Type = d.desugarType(pat.Type, false, vars)
			body, err := d.nestLambdaPats(pats[1:], exp, locals, tvars)
			if err != nil {
				return nil, err
			}
			return ast.Lambda{
				Binder: bind,
				Body:   body,
				Span:   lexer.NewSpan(pat.Span, exp.GetSpan()),
			}, nil
		}
	default:
		{
			vars := data.MapSlice(pats, func(t ast.SPattern) ast.Expr { return ast.Var{Name: d.newVar(), Span: t.GetSpan()} })
			span := lexer.NewSpan(pat.GetSpan(), exp.GetSpan())
			dpats, err := data.MapSliceError(pats, func(t ast.SPattern) (ast.Pattern, error) { return d.desugarPattern(t, locals, tvars) })
			if err != nil {
				return nil, err
			}
			expr := ast.Match{Exps: vars, Cases: []ast.Case{{Patterns: dpats, Exp: exp}}, Span: span}
			binders := data.MapSlice(vars, func(t ast.Expr) ast.Binder { return ast.Binder{Name: t.(ast.Var).Name, Span: t.GetSpan()} })
			return d.nestLambdas(binders, expr), nil
		}
	}
}

func (d *Desugar) nestLambdas(binders []ast.Binder, exp ast.Expr) ast.Expr {
	if len(binders) <= 0 {
		return exp
	}
	return ast.Lambda{Binder: binders[0], Body: d.nestLambdas(binders[1:], exp), Span: exp.GetSpan()}
}

func (d *Desugar) nestLets(ld ast.SLetDef, exp ast.Expr, span lexer.Span, locals data.Set[string], tvars map[string]tc.Type) (ast.Expr, error) {
	switch l := ld.(type) {
	case ast.SLetBind:
		{
			def, err := d.desugarDefBind(l, locals, tvars)
			if err != nil {
				return nil, err
			}
			return ast.Let{Def: def, Body: exp, Span: span}, nil
		}
	case ast.SLetPat:
		{
			pat, err := d.desugarPattern(l.Pat, locals, tvars)
			if err != nil {
				return nil, err
			}
			cas := ast.Case{Patterns: []ast.Pattern{pat}, Exp: exp}
			expr, err2 := d.desugarExp(l.Expr, locals, tvars)
			if err2 != nil {
				return nil, err2
			}
			return ast.Match{Exps: []ast.Expr{expr}, Cases: []ast.Case{cas}, Span: span}, nil
		}
	default:
		panic("Got unknow letdef in nestLets")
	}
}

func (d *Desugar) nestRecordSelects(exp ast.Expr, labels []ast.Spanned[string], span lexer.Span) ast.Expr {
	if len(labels) == 0 {
		return exp
	}
	return d.nestRecordSelects(ast.RecordSelect{Exp: exp, Label: labels[0], Span: span}, labels[1:], span)
}

func (d *Desugar) nestRecordRestrictions(exp ast.Expr, labels []string, span lexer.Span) ast.Expr {
	if len(labels) == 0 {
		return exp
	}
	return d.nestRecordRestrictions(ast.RecordRestrict{Exp: exp, Label: labels[0], Span: span}, labels[1:], span)
}

func (d *Desugar) nestRecordUpdates(exp ast.Expr, labels []ast.Spanned[string], value ast.Expr, isSet bool, span lexer.Span) ast.Expr {
	if len(labels) == 0 {
		return exp
	}
	tail := labels[1:]
	shouldSet := isSet || len(tail) > 0
	selec := value
	if len(tail) > 0 {
		selec = ast.RecordSelect{Exp: exp, Label: labels[0], Span: value.GetSpan()}
	}
	return ast.RecordUpdate{
		Exp:   exp,
		Label: labels[0],
		Value: d.nestRecordUpdates(selec, tail, value, isSet, span),
		IsSet: shouldSet,
		Span:  span,
	}
}

func (d *Desugar) convertDoLets(exps []ast.SExpr) []ast.SExpr {
	doLetIndex := slices.IndexFunc(exps, func(e ast.SExpr) bool {
		_, isDolet := e.(ast.SDoLet)
		return isDolet
	})
	if doLetIndex == -1 {
		return exps
	}
	exp := exps[0]
	if do, isDolet := exp.(ast.SDoLet); isDolet {
		body := d.convertDoLets(exps[1:])
		bodyExp := body[0]
		if len(body) > 1 {
			bodyExp = ast.SDo{Exps: body, Span: do.Span}
		}
		return []ast.SExpr{ast.SLet{Def: do.Def, Body: bodyExp, Span: lexer.NewSpan(do.Span, bodyExp.GetSpan())}}
	} else {
		res := []ast.SExpr{exp}
		return append(res, d.convertDoLets(exps[1:])...)
	}
}

// Resolve all type aliases in this type
func (d *Desugar) resolveAliases(ty ast.SType) ast.SType {
	switch t := ty.(type) {
	case ast.STConst:
		{
			syn, has := d.synonyms[t.Name]
			if has {
				if exp := syn.Expanded; exp != nil {
					return exp.Clone().WithSpan(t.Span)
				} else {
					return ty
				}
			}
			return ty
		}
	case ast.STParens:
		{
			par := t.Clone().(ast.STParens)
			par.Type = d.resolveAliases(par.Type)
			return par
		}
	case ast.STFun:
		{
			fun := t.Clone().(ast.STFun)
			fun.Arg = d.resolveAliases(fun.Arg)
			fun.Ret = d.resolveAliases(fun.Ret)
			return fun
		}
	case ast.STApp:
		{
			typ := t.Type.(ast.STConst)
			syn, has := d.synonyms[typ.Name]
			if has {
				synTy := syn.Expanded
				if synTy == nil {
					panic("Got unexpanded type alias: " + syn.Name)
				}
				synTy = synTy.Clone().WithSpan(t.Type.GetSpan())
				zip := data.ZipSlices(syn.TyVars, t.Types)

				res := synTy
				for _, tu := range zip {
					res = ast.SubstVar(res, tu.V1, d.resolveAliases(tu.V2))
				}
				return res
			}
			app := t.Clone().(ast.STApp)
			app.Types = data.MapSlice(app.Types, d.resolveAliases)
			return app
		}
	case ast.STRecord:
		{
			rec := t.Clone().(ast.STRecord)
			rec.Row = d.resolveAliases(rec.Row)
			return rec
		}
	case ast.STRowEmpty:
		return t
	case ast.STRowExtend:
		{
			rec := t.Clone().(ast.STRowExtend)
			rec.Row = d.resolveAliases(rec.Row)
			rec.Labels = data.LabelMapValues(rec.Labels, d.resolveAliases)
			return rec
		}
	case ast.STImplicit:
		{
			imp := t.Clone().(ast.STImplicit)
			imp.Type = d.resolveAliases(imp.Type)
			return imp
		}
	default:
		panic("Unknow type in resolveAliases: " + t.String())
	}
}

// Make sure variables are not co-dependent (form cycles)
// and order them by dependency
func (d *Desugar) validateTopLevelValues(desugared []ast.Decl) ([]ast.Decl, error) {
	reportCycle := func(cycle []data.DagNode[string, ast.ValDecl]) {
		first := cycle[0]
		vars := data.MapSlice(cycle, func(t data.DagNode[string, ast.ValDecl]) string { return t.Val })
		if len(cycle) == 1 {
			d.errors = append(d.errors, d.makeError(data.CycleInValues(vars), first.Data.Span))
		} else if data.AnySlice(cycle, func(t data.DagNode[string, ast.ValDecl]) bool { return isVariable(t.Data.Exp) }) {
			for _, node := range cycle {
				d.errors = append(d.errors, d.makeError(data.CycleInValues(vars), node.Data.Span))
			}
		} else {
			for _, node := range cycle {
				d.errors = append(d.errors, d.makeError(data.CycleInFunctions(vars), node.Data.Span))
			}
		}
	}

	collectDependencies := func(exp ast.Expr) data.Set[string] {
		deps := data.NewSet[string]()
		ast.EverywhereExprUnit(exp, func(expr ast.Expr) {
			switch e := expr.(type) {
			case ast.Var:
				deps.Add(e.Fullname())
			case ast.ImplicitVar:
				deps.Add(e.Fullname())
			case ast.Ctor:
				deps.Add(e.Fullname())
			}
		})
		return deps
	}

	var decls []ast.ValDecl
	var types []ast.TypeDecl
	for _, decl := range desugared {
		switch d := decl.(type) {
		case ast.ValDecl:
			decls = append(decls, d)
		case ast.TypeDecl:
			types = append(types, d)
		}
	}

	deps := make(map[string]data.Set[string])
	nodes := make(map[string]*data.DagNode[string, ast.ValDecl])
	dag := data.NewDag[string, ast.ValDecl](len(decls))
	for _, d := range decls {
		deps[d.Name.Val] = collectDependencies(d.Exp)
		node := data.NewDagNode(d.Name.Val, d)
		nodes[d.Name.Val] = node
		dag.AddNodes(node)
	}

	for _, node := range nodes {
		set, has := deps[node.Val]
		if !has {
			continue
		}
		for name := range set.Inner() {
			dep, has := nodes[name]
			if !has {
				continue
			}
			d1, d2 := dep.Data, node.Data
			if isVariable(d1.Exp) || isVariable(d2.Exp) {
				// variables cannot have cycles
				dep.Link(node)
			} else if d1.Name.Val == d2.Name.Val {
				// functions can be recursive
			} else if (d1.Signature == nil || d1.Signature.Type == nil) || (d2.Signature == nil || d2.Signature.Type == nil) {
				// functions can only be mutually recursive if they have type annotations
				dep.Link(node)
			}
		}
	}

	cycle, hasCycle := dag.FindCycle()
	if hasCycle {
		reportCycle(cycle)
		return nil, errors.New("module has cycles")
	}

	ordered := dag.Toposort().ToSlice()
	res := make([]ast.Decl, 0, len(types)+len(ordered))
	for _, typ := range types {
		res = append(res, typ)
	}
	for _, decl := range ordered {
		res = append(res, decl.Data)
	}
	return res, nil
}

func isVariable(exp ast.Expr) bool {
	switch e := exp.(type) {
	case ast.Lambda:
		return false
	case ast.Ann:
		return isVariable(e.Exp)
	default:
		return true
	}
}

type CollectedVar struct {
	name     string
	span     lexer.Span
	implicit bool
	instance bool
}

func (d *Desugar) collectVarsLdef(ldef ast.SLetDef) []CollectedVar {
	switch ld := ldef.(type) {
	case ast.SLetBind:
		return []CollectedVar{{name: ld.Name.Name, span: ld.Name.Span, implicit: ld.Name.IsImplicit, instance: ld.IsInstance}}
	case ast.SLetPat:
		return d.collectVars(ld.Pat, false)
	default:
		panic("Unknow letdef type in collectVarsLdef")
	}
}

func (d *Desugar) collectVars(pat ast.SPattern, implicit bool) []CollectedVar {
	switch p := pat.(type) {
	case ast.SVarP:
		return []CollectedVar{{name: p.V.Name, span: p.V.Span, implicit: implicit}}
	case ast.SParensP:
		return d.collectVars(p.Pat, implicit)
	case ast.SCtorP:
		return data.FlatMapSlice(p.Fields, func(t ast.SPattern) []CollectedVar { return d.collectVars(t, implicit) })
	case ast.SRecordP:
		return data.LabelFlatMapValues(p.Labels, func(t ast.SPattern) []CollectedVar { return d.collectVars(t, implicit) })
	case ast.SListP:
		{
			tail := []CollectedVar{}
			if p.Tail != nil {
				tail = d.collectVars(p.Tail, implicit)
			}
			res := data.FlatMapSlice(p.Elems, func(t ast.SPattern) []CollectedVar { return d.collectVars(t, implicit) })
			res = append(res, tail...)
			return res
		}
	case ast.SNamed:
		{
			vars := d.collectVars(p.Pat, implicit)
			return append(vars, CollectedVar{name: p.Name.Val, span: p.Name.Span})
		}
	case ast.SImplicitP:
		return d.collectVars(p.Pat, true)
	case ast.SWildcard:
		return []CollectedVar{}
	case ast.SLiteralP:
		return []CollectedVar{}
	case ast.SUnitP:
		return []CollectedVar{}
	case ast.SRegexP:
		return []CollectedVar{}
	case ast.STypeTest:
		if p.Alias != nil {
			return []CollectedVar{{name: *p.Alias, span: p.Span, implicit: implicit}}
		} else {
			return []CollectedVar{}
		}
	case ast.STypeAnnotationP:
		return []CollectedVar{{name: p.Par.Name, span: p.Par.Span, implicit: implicit}}
	case ast.STupleP:
		{
			vars := d.collectVars(p.P1, implicit)
			return append(vars, d.collectVars(p.P2, implicit)...)
		}
	default:
		panic("Got unknow pattern in collectVars: " + p.String())
	}
}

func collectVars(e ast.Expr) []string {
	return ast.EverywhereExprAcc(e, func(e ast.Expr) []string {
		if v, isVar := e.(ast.Var); isVar {
			return []string{v.Fullname()}
		} else {
			return []string{}
		}
	})
}

func (d *Desugar) validateDataCtorNames(dd ast.STypeDecl) {
	if len(dd.DataCtors) > 1 {
		tyName := dd.Binder.Val
		for _, ct := range dd.DataCtors {
			if ct.Name.Val == tyName {
				d.errors = append(d.errors, d.makeError(data.WrongConstructorName(tyName), dd.Span))
			}
		}
	}
}

func (d *Desugar) newVar() string {
	d.varCount++
	return fmt.Sprintf("__var%d", d.varCount)
}

func (d *Desugar) addUnusedVars() {
	for name, span := range d.unusedVars {
		d.errors = append(d.errors, d.makeWarn(data.UnusedVariable(name), span))
	}
}

func (d *Desugar) checkAlias(alias string, span lexer.Span) {
	if !d.aliasedImports.Contains(alias) {
		d.errors = append(d.errors, d.makeError(data.NoAliasFound(alias), span))
	}
}

func (d *Desugar) checkShadow(name string, span lexer.Span) {
	_, has := d.imports[name]
	if has {
		err := d.makeError(data.ShadowedVariable(name), span)
		d.errors = append(d.errors, err)
	}
}

func (d *Desugar) makeError(msg string, span lexer.Span) ast.CompilerProblem {
	return ast.CompilerProblem{Msg: msg, Span: span, Filename: d.smod.SourceName, Module: &d.modName, Severity: ast.ERROR}
}

func (d *Desugar) makeWarn(msg string, span lexer.Span) ast.CompilerProblem {
	return ast.CompilerProblem{Msg: msg, Span: span, Filename: d.smod.SourceName, Module: &d.modName, Severity: ast.WARN}
}
