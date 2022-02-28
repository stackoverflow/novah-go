package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/lexer"
	"golang.org/x/exp/slices"
)

const (
	tab      = "  " // 2 spaces
	max_cols = 120
)

type Formatter struct {
	builder strings.Builder
	tab     string
}

func NewFormatter() *Formatter {
	return &Formatter{tab: ""}
}

func Format(ast SModule) string {
	f := &Formatter{tab: ""}
	return f.ShowModule(ast)
}

func (f *Formatter) ShowModule(m SModule) string {
	var build strings.Builder
	if m.Comment != nil {
		build.WriteString(f.ShowComment(*m.Comment, true))
	}
	build.WriteString("module ")
	build.WriteString(m.Name.Val)

	if len(m.Imports) > 0 {
		build.WriteString("\n\n")
		imps := make([]Import, len(m.Imports))
		copy(imps, m.Imports)
		slices.SortStableFunc(imps, func(i, j Import) bool { return i.Module.Val < j.Module.Val })
		build.WriteString(data.JoinToStringFunc(imps, "\n", f.ShowImport))
	}

	if len(m.Decls) > 0 {
		for _, d := range m.Decls {
			build.WriteString("\n\n")
			build.WriteString(f.ShowDecl(d))
		}
	}
	return build.String()
}

func (f *Formatter) ShowImport(imp Import) string {
	cmt := ""
	if imp.Comment != nil {
		cmt = f.ShowComment(*imp.Comment, true)
	}
	exposes := ""
	if len(imp.Defs) > 0 {
		exposes = fmt.Sprintf("(%s)", data.JoinToStringFunc(imp.Defs, ", ", f.ShowDeclarationRef))
	}
	alias := ""
	if imp.Alias != "" {
		alias = " as " + imp.Alias
	}
	return fmt.Sprintf("%simport %s%s%s", cmt, imp.Module.Val, exposes, alias)
}

func (f *Formatter) ShowDeclarationRef(ref DeclarationRef) string {
	if ref.Tag == VAR {
		return ref.Name.Val
	}
	if ref.All {
		return fmt.Sprintf("%s(..)", ref.Name.Val)
	}
	if len(ref.Ctors) == 0 {
		return ref.Name.Val
	}
	return fmt.Sprintf("%s(%s)", ref.Name.Val, data.JoinToStringFunc(ref.Ctors, ", ", func(t Spanned[string]) string { return t.Val }))
}

func (f *Formatter) ShowDecl(d SDecl) string {
	cmt := ""
	if d.GetComment() != nil {
		cmt = f.ShowComment(*d.GetComment(), true)
	}
	vis := ""
	if d.GetVisibility() == PUBLIC {
		vis = "pub\n"
	}
	var body string
	switch dd := d.(type) {
	case STypeDecl:
		{
			visi := vis
			if dd.DataCtors[0].Visibility == PUBLIC {
				visi = "pub+\n"
			}
			body = fmt.Sprintf("%s%s", visi, f.ShowTypeDecl(dd))
		}
	case SValDecl:
		{
			visi := ""
			if dd.Visibility == PUBLIC && dd.IsInstance {
				visi = "pub instance\n"
			} else if dd.Visibility == PUBLIC {
				visi = "pub\n"
			} else if dd.IsInstance {
				visi = "instance\n"
			}
			body = fmt.Sprintf("%s%s", visi, f.ShowValDec(dd))
		}
	case STypeAliasDecl:
		body = fmt.Sprintf("%s%s", vis, f.ShowTypealiasDecl(dd))
	}
	return fmt.Sprintf("%s %s", cmt, body)
}

func (f *Formatter) ShowValDec(vd SValDecl) string {
	var prefix strings.Builder
	if vd.Signature != nil && vd.Signature.Type != nil {
		prefix.WriteString(f.showNameType(vd.Binder.Val, vd.Signature.Type, vd.IsOperator))
		prefix.WriteString("\n")
	}
	prefix.WriteString(vd.Binder.Val)
	prefix.WriteString(" ")
	prefix.WriteString(data.JoinToStringFunc(vd.Pats, " ", f.ShowPattern))
	prefix.WriteString(" =")

	if isSimpleExpr(vd.Exp) {
		prefix.WriteString(" ")
		prefix.WriteString(f.ShowExpr(vd.Exp))
	} else {
		prefix.WriteString(f.withIndentDef(func() string { return f.tab + f.ShowExpr(vd.Exp) }))
	}
	return prefix.String()
}

func (f *Formatter) ShowTypeDecl(td STypeDecl) string {
	var str strings.Builder
	str.WriteString("type ")
	str.WriteString(td.Binder.Val)
	if len(td.TyVars) > 0 {
		str.WriteString(" ")
		str.WriteString(data.JoinToStringStr(td.TyVars, " "))
	}
	if len(td.DataCtors) == 1 {
		str.WriteString(" = ")
		str.WriteString(f.ShowDataCtor(td.DataCtors[0]))
	} else {
		str.WriteString(f.withIndentDef(func() string {
			ctors := f.tab + "= "
			ctors += data.JoinToStringFunc(td.DataCtors, "\n"+f.tab+"| ", f.ShowDataCtor)
			return ctors
		}))
	}
	return str.String()
}

func (f *Formatter) ShowTypealiasDecl(td STypeAliasDecl) string {
	vars := ""
	if len(td.TyVars) > 0 {
		vars = data.JoinToStringStr(td.TyVars, " ") + " "
	}
	return fmt.Sprintf("typealias %s %s= %s", td.Name, vars, f.ShowType(td.Type))
}

func (f *Formatter) showNameType(name string, ty SType, isOp bool) string {
	theName := name
	if isOp {
		theName = fmt.Sprintf("(%s)", name)
	}
	td := fmt.Sprintf("%s : %s", theName, f.ShowType(ty))
	if len(td) > max_cols {
		return fmt.Sprintf("%s%s", theName, f.withIndentDef(func() string { return f.showTypeIndented(ty, ":") }))
	} else {
		return td
	}
}

func (f *Formatter) ShowDataCtor(dc SDataCtor) string {
	args := data.JoinToStringFunc(dc.Args, " ", f.ShowType)
	return fmt.Sprintf("%s %s", dc.Name.Val, args)
}

func (f *Formatter) ShowExpr(exp SExpr) string {
	cmt := ""
	if exp.GetComment() != nil {
		cmt = f.ShowComment(*exp.GetComment(), false)
	}
	var estr string
	switch e := exp.(type) {
	case SDo:
		estr = data.JoinToStringFunc(e.Exps, "\n"+f.tab, f.ShowExpr)
	case SMatch:
		{
			exps := data.JoinToStringFunc(e.Exprs, ", ", f.ShowExpr)
			cases := f.withIndentDef(func() string {
				return data.JoinToStringFunc(e.Cases, "\n"+f.tab, f.ShowCase)
			})
			estr = fmt.Sprintf("case %s of%s%s", exps, f.tab, cases)
		}
	case SLet:
		{
			var str string
			if isSimpleExpr(e.Body) {
				str = fmt.Sprintf("in %s", f.ShowExpr(e.Body))
			} else {
				str = fmt.Sprintf("in %s", f.withIndentDef(func() string { return f.tab + f.ShowExpr(e.Body) }))
			}
			estr = fmt.Sprintf("let %s\n%s%s", f.ShowLetDef(e.Def, false), f.tab, str)
		}
	case SDoLet:
		estr = fmt.Sprintf("let %s", f.ShowLetDef(e.Def, false))
	case SLetBang:
		if e.Body == nil {
			estr = fmt.Sprintf("let! %s", f.ShowLetDef(e.Def, false))
		} else {
			estr = fmt.Sprintf("let! %s in %s", f.ShowLetDef(e.Def, false), f.ShowExpr(e.Body))
		}
	case SFor:
		estr = fmt.Sprintf("for %s, do %s", f.ShowLetDef(e.Def, true), f.ShowExpr(e.Body))
	case SDoBang:
		estr = fmt.Sprintf("do! %s", f.ShowExpr(e.Exp))
	case SIf:
		if e.Else != nil {
			if isSimpleExpr(e.Then) && isSimpleExpr(e.Else) {
				estr = fmt.Sprintf("if %s then %s else %s", f.ShowExpr(e.Cond), f.ShowExpr(e.Then), f.ShowExpr(e.Else))
			} else {
				estr = fmt.Sprintf("if %s\n%sthen %s\n%selse %s", f.ShowExpr(e.Cond), f.tab, f.ShowExpr(e.Then), f.tab, f.ShowExpr(e.Else))
			}
		} else {
			if isSimpleExpr(e.Then) {
				estr = fmt.Sprintf("if %s then %s", f.ShowExpr(e.Cond), f.ShowExpr(e.Then))
			} else {
				estr = fmt.Sprintf("if %s\n%sthen %s", f.ShowExpr(e.Cond), f.tab, f.ShowExpr(e.Then))
			}
		}
	case SReturn:
		estr = fmt.Sprintf("return %s", f.ShowExpr(e.Exp))
	case SYield:
		estr = fmt.Sprintf("yield %s", f.ShowExpr(e.Exp))
	case SApp:
		estr = fmt.Sprintf("%s %s", f.ShowExpr(e.Fn), f.ShowExpr(e.Arg))
	case SLambda:
		{
			var show string
			if isSimpleExpr(e.Body) {
				show = fmt.Sprintf(" -> %s", f.ShowExpr(e.Body))
			} else {
				show = fmt.Sprintf(" ->%s", f.withIndentDef(func() string { return f.tab + f.ShowExpr(e.Body) }))
			}
			pats := data.JoinToStringFunc(e.Pats, " ", f.ShowPattern)
			estr = fmt.Sprintf("\\%s%s", pats, show)
		}
	case SVar:
		if e.Alias != nil {
			estr = fmt.Sprintf("%s.%s", *e.Alias, e.Name)
		} else {
			estr = e.Name
		}
	case SOperator:
		{
			var val string
			if e.Alias != nil {
				val = fmt.Sprintf("%s.%s", *e.Alias, e.Name)
			} else {
				val = e.Name
			}
			if e.IsPrefix {
				estr = fmt.Sprintf("`%s`", val)
			} else {
				estr = val
			}
		}
	case SImplicitVar:
		if e.Alias != nil {
			estr = fmt.Sprintf("{{%s.%s}}", *e.Alias, e.Name)
		} else {
			estr = fmt.Sprintf("{{%s}}", e.Name)
		}
	case SConstructor:
		if e.Alias != nil {
			estr = fmt.Sprintf("%s.%s", *e.Alias, e.Name)
		} else {
			estr = e.Name
		}
	case SInt:
		estr = e.Text
	case SFloat:
		estr = e.Text
	case SComplex:
		estr = e.Text
	case SChar:
		estr = fmt.Sprintf("'%s'", e.Raw)
	case SString:
		if e.Multi {
			estr = fmt.Sprintf(`"""%s"""`, e.Raw)
		} else {
			estr = fmt.Sprintf(`"%s"`, e.Raw)
		}
	case SBool:
		estr = strconv.FormatBool(e.V)
	case SPatternLiteral:
		estr = fmt.Sprintf(`#"%s"`, e.Raw)
	case SParens:
		estr = fmt.Sprintf("(%s)", f.ShowExpr(e.Exp))
	case SUnit:
		estr = "()"
	case SUnderscore:
		estr = "_"
	case SRecordEmpty:
		estr = "{}"
	case SRecordSelect:
		{
			labels := data.JoinToStringFunc(e.Labels, ".", func(t Spanned[string]) string { return t.Val })
			estr = fmt.Sprintf("%s.%s", f.ShowExpr(e.Exp), labels)
		}
	case SRecordRestrict:
		{
			labels := data.JoinToStringFunc(e.Labels, ", ", func(l string) string { return data.ShowLabel(l) })
			estr = fmt.Sprintf("{ - %s | %s }", labels, f.ShowExpr(e.Exp))
		}
	case SRecordUpdate:
		{
			labels := data.JoinToStringFunc(e.Labels, ".", func(l Spanned[string]) string { return data.ShowLabel(l.Val) })
			estr = fmt.Sprintf("{ .%s = %s | %s }", labels, f.ShowExpr(e.Val), f.ShowExpr(e.Exp))
		}
	case SRecordExtend:
		{
			labels := e.Labels.Show(f.showLabelExpr)
			_, isEmpty := e.Exp.(SRecordEmpty)
			if isEmpty {
				estr = fmt.Sprintf("{ %s }", labels)
			} else {
				estr = fmt.Sprintf("{ %s | %s }", labels, f.ShowExpr(e.Exp))
			}
		}
	case SRecordMerge:
		estr = fmt.Sprintf("{ + %s, %s }", f.ShowExpr(e.Exp1), f.ShowExpr(e.Exp2))
	case SListLiteral:
		estr = fmt.Sprintf("[%s]", data.JoinToStringFunc(e.Exps, ", ", f.ShowExpr))
	case SSetLiteral:
		estr = fmt.Sprintf("[%s]", data.JoinToStringFunc(e.Exps, ", ", f.ShowExpr))
	case SIndex:
		estr = fmt.Sprintf("%s.[%s]", f.ShowExpr(e.Exp), f.ShowExpr(e.Index))
	case SBinApp:
		estr = fmt.Sprintf("%s %s %s", f.ShowExpr(e.Left), f.ShowExpr(e.Op), f.ShowExpr(e.Right))
	case SWhile:
		{
			body := f.withIndentDef(func() string {
				return data.JoinToStringFunc(e.Exps, "\n"+f.tab, f.ShowExpr)
			})
			estr = fmt.Sprintf("while %s do%s", f.ShowExpr(e.Cond), body)
		}
	case SComputation:
		{
			body := f.withIndentDef(func() string {
				return data.JoinToStringFunc(e.Exps, "\n"+f.tab, f.ShowExpr)
			})
			estr = fmt.Sprintf("do.%s%s", e.Builder.Name, body)
		}
	case SNil:
		estr = "nil"
	case STypeCast:
		estr = fmt.Sprintf("%s as %s", f.ShowExpr(e.Exp), f.ShowType(e.Cast))
	}
	return fmt.Sprintf("%s%s", cmt, estr)
}

func (f *Formatter) ShowLetDef(l SLetDef, isFor bool) string {
	sep := "="
	if isFor {
		sep = "in"
	}
	var prefix string
	switch def := l.(type) {
	case SLetBind:
		{
			typ := ""
			if def.Type != nil {
				typ = fmt.Sprintf("%s : %s\n%s", def.Name.Name, f.ShowType(def.Type), f.tab)
			}
			pats := data.JoinToStringFunc(def.Pats, " ", func(p SPattern) string { return f.ShowPattern(p) })
			prefix = fmt.Sprintf("%s%s %s %s", typ, def.Name.Name, pats, sep)
		}
	case SLetPat:
		prefix = fmt.Sprintf("%s %s", f.ShowPattern(def.Pat), sep)
	}
	if isSimpleExpr(l.GetExpr()) {
		return fmt.Sprintf("%s %s", prefix, f.ShowExpr(l.GetExpr()))
	}
	return prefix + f.withIndentDef(func() string { return fmt.Sprintf("%s%s", f.tab, f.ShowExpr(l.GetExpr())) })
}

func (f *Formatter) ShowComment(c lexer.Comment, newline bool) string {
	if !c.IsMulti {
		return fmt.Sprintf("// %s\n", c.Text)
	}
	end := ""
	if newline {
		end = "\n"
	}
	return fmt.Sprintf("/*%s*/%s", c.Text, end)
}

func (f *Formatter) ShowCase(cas SCase) string {
	guard := ""
	if cas.Guard != nil {
		guard = fmt.Sprintf(" if %s", f.ShowExpr(cas.Guard))
	}
	pats := data.JoinToStringFunc(cas.Pats, ", ", func(p SPattern) string { return f.ShowPattern(p) })
	return fmt.Sprintf("%s%s -> %s", pats, guard, f.ShowExpr(cas.Exp))
}

func (f *Formatter) ShowPattern(pat SPattern) string {
	switch p := pat.(type) {
	case SWildcard:
		return "_"
	case SVarP:
		return f.ShowExpr(p.V)
	case SCtorP:
		{
			fields := data.JoinToStringFunc(p.Fields, " ", func(pa SPattern) string { return f.ShowPattern(pa) })
			return fmt.Sprintf("%s %s", f.ShowExpr(p.Ctor), fields)
		}
	case SLiteralP:
		return f.ShowExpr(p.Lit)
	case SParensP:
		return fmt.Sprintf("(%s)", f.ShowPattern(p.Pat))
	case SRecordP:
		{
			labels := p.Labels.Show(func(l string, pat SPattern) string { return fmt.Sprintf("%s: %s", l, f.ShowPattern(pat)) })
			return fmt.Sprintf("{ %s }", labels)
		}
	case SListP:
		{
			tail := ""
			if p.Tail != nil {
				tail = fmt.Sprintf(" :: %s", f.ShowPattern(p.Tail))
			}
			pats := data.JoinToStringFunc(p.Elems, ", ", func(pa SPattern) string { return f.ShowPattern(pa) })
			return fmt.Sprintf("[%s%s]", pats, tail)
		}
	case SNamed:
		return fmt.Sprintf("%s as %s", f.ShowPattern(p.Pat), p.Name.Val)
	case SUnitP:
		return "()"
	case STypeTest:
		{
			alias := ""
			if p.Alias != nil {
				alias = fmt.Sprintf(" as %s", *p.Alias)
			}
			return fmt.Sprintf(":? %s%s", f.ShowType(p.Type), alias)
		}
	case SImplicitP:
		return fmt.Sprintf("{{%s}}", f.ShowPattern(p.Pat))
	case STypeAnnotationP:
		return fmt.Sprintf("%s : %s", f.ShowExpr(p.Par), f.ShowType(p.Type))
	case STupleP:
		return fmt.Sprintf("%s ; %s", f.ShowPattern(p.P1), f.ShowPattern(p.P2))
	case SRegexP:
		return fmt.Sprintf(`#"%s"`, f.ShowExpr(p.Regex))
	default:
		panic("received unknow pattern " + p.String())
	}
}

func (f *Formatter) ShowType(ty SType) string {
	switch t := ty.(type) {
	case STFun:
		{
			var tlist []string
			fn := ty
			for isFun(fn) {
				fnf := fn.(STFun)
				tlist = append(tlist, f.ShowType(fnf.Arg))
				fn = fnf.Ret
			}
			tlist = append(tlist, f.ShowType(fn))
			return data.JoinToStringStr(tlist, " -> ")
		}
	case STApp:
		return fmt.Sprintf("%s %s", f.ShowType(t.Type), data.JoinToStringFunc(t.Types, " ", func(typ SType) string {
			return f.ShowType(typ)
		}))
	case STParens:
		return fmt.Sprintf("(%s)", f.ShowType(t.Type))
	case STConst:
		return t.Fullname()
	case STRecord:
		return f.ShowType(t.Row)
	case STRowEmpty:
		return "{}"
	case STRowExtend:
		{
			labels := t.Labels.Show(f.showLabelType)
			_, isEmpty := t.Row.(STRowEmpty)
			if isEmpty {
				return fmt.Sprintf("{ %s }", labels)
			} else {
				return fmt.Sprintf("{ %s | %s }", labels, f.ShowType(t.Row))
			}
		}
	case STImplicit:
		return fmt.Sprintf("{{ %s }}", f.ShowType(t.Type))
	default:
		panic("unknow type in formatter")
	}
}

func (f *Formatter) showTypeIndented(ty SType, prefix string) string {
	switch ty.(type) {
	case STFun:
		{
			var tlist []string
			fn := ty
			for isFun(fn) {
				fnf := fn.(STFun)
				tlist = append(tlist, f.ShowType(fnf.Arg))
				fn = fnf.Ret
			}
			tlist = append(tlist, f.ShowType(fn))
			return fmt.Sprintf("%s%s %s", f.tab, prefix, data.JoinToStringStr(tlist, fmt.Sprintf("\n%s-> ", f.tab)))
		}
	default:
		return f.ShowType(ty)
	}
}

func (f *Formatter) showLabelExpr(l string, e SExpr) string {
	return fmt.Sprintf("%s: %s", l, f.ShowExpr(e))
}

func (f *Formatter) showLabelType(l string, ty SType) string {
	return fmt.Sprintf("%s : %s", l, f.ShowType(ty))
}

func isSimpleExpr(exp SExpr) bool {
	switch exp.(type) {
	case SInt, SFloat, SComplex, SString, SChar, SBool, SVar, SOperator, SPatternLiteral:
		return true
	default:
		return false
	}
}

func (f *Formatter) withIndentDef(fun func() string) string {
	return f.withIndent(true, fun)
}

func (f *Formatter) withIndent(shouldBreak bool, fun func() string) string {
	oldIndent := f.tab
	f.tab = fmt.Sprintf("%s%s", f.tab, tab)
	res := fun()
	f.tab = oldIndent
	if shouldBreak {
		return fmt.Sprintf("\n%s", res)
	}
	return res
}

func isFun(ty SType) bool {
	switch ty.(type) {
	case STFun:
		return true
	default:
		return false
	}
}
