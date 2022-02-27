package parser

import (
	"math"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type parser struct {
	sourceName string

	iter       *PeekableIterator
	moduleName string
	errors     []data.CompilerProblem
}

func NewParser(tokens *lexer.Lexer) *parser {
	return &parser{
		sourceName: tokens.Name,
		iter:       newPeekableIterator(tokens, throwMismatchedIdentation),
	}
}

func (p *parser) ParseFullModule() (res ast.SModule, errs []data.CompilerProblem) {
	defer func() {
		if r := recover(); r != nil {
			var msg string
			var span data.Span
			switch e := r.(type) {
			case lexer.LexerError:
				{
					msg = e.Msg
					span = e.Span
				}
			case ParserError:
				{
					msg = e.msg
					span = e.span
				}
			default:
				panic("Got unexpected error in parseFullModule")
			}
			errs = append(errs, p.errors...)
			errs = append(errs, data.CompilerProblem{Msg: msg, Span: span, Filename: p.sourceName, Module: p.moduleName, Severity: data.FATAL})
		}
	}()

	res = p.parseFullModule()
	errs = append(errs, p.errors...)
	return
}

func (p *parser) parseFullModule() ast.SModule {
	mdef := p.parseModule()
	p.moduleName = mdef.name.Val

	var imports []ast.Import
	next := p.iter.peek().Type
	// TODO: parse foreign imports
	if next == lexer.IMPORT {
		imports = p.parseImports()
	}
	// TODO: add default imports

	decls := make([]ast.SDecl, 0, 5)
	for p.iter.peek().Type != lexer.EOF {
		func() {
			defer func() {
				if r := recover(); r != nil {
					var msg string
					var span data.Span
					switch e := r.(type) {
					case lexer.LexerError:
						{
							msg = e.Msg
							span = e.Span
						}
					case ParserError:
						{
							msg = e.msg
							span = e.span
						}
					default:
						panic("Got unexpected error in parseFullModule")
					}
					p.errors = append(p.errors, data.CompilerProblem{Msg: msg, Span: span, Filename: p.sourceName, Module: p.moduleName, Severity: data.ERROR})
					p.fastForward()
				}
			}()

			decls = append(decls, p.parseDecl())
		}()
	}

	return ast.SModule{
		Name:       mdef.name,
		SourceName: p.sourceName,
		Imports:    imports,
		Decls:      decls,
		Span:       mdef.span,
		Comment:    mdef.comment,
	}
}

func (p *parser) parseModule() ModuleDef {
	m := p.expect(lexer.MODULE, withError(data.MODULE_DEFINITION))
	return ModuleDef{name: p.parseModuleName(), span: span(m.Span, p.iter.current.Span), comment: m.Comment}
}

func (p *parser) parseModuleName() ast.Spanned[string] {
	idents := between(p, lexer.DOT, func() lexer.Token {
		return p.expect(lexer.IDENT, withError(data.MODULE_NAME))
	})
	for _, it := range idents {
		v := *it.Text
		if v[len(v)-1] == '?' || v[len(v)-1] == '!' {
			throwError2(data.MODULE_NAME, it.Span)
		}
	}
	span := span(idents[0].Span, idents[len(idents)-1].Span)
	name := data.JoinToStringFunc(idents, ".", func(t lexer.Token) string { return *t.Text })
	return ast.Spanned[string]{Val: name, Span: span}
}

func (p *parser) parseImports() []ast.Import {
	imps := make([]ast.Import, 0, 2)
	for {
		if p.iter.peek().Type == lexer.IMPORT {
			imps = append(imps, p.parseImport())
		} else {
			break
		}
	}
	return imps
}

func (p *parser) parseImport() ast.Import {
	impTk := p.expect(lexer.IMPORT, noErr())
	mod := p.parseModuleName()
	var impor ast.Import
	switch p.iter.peek().Type {
	case lexer.LPAREN:
		{
			imp := p.parseDeclarationRefs()
			if p.iter.peek().Type == lexer.AS {
				p.iter.next()
				alias := p.expect(lexer.UPPERIDENT, withError(data.IMPORT_ALIAS))
				impor = ast.Import{Module: mod, Defs: imp, Alias: alias.Text, Span: span(impTk.Span, p.iter.current.Span)}
			} else {
				impor = ast.Import{Module: mod, Defs: imp, Span: span(impTk.Span, p.iter.current.Span)}
			}
		}
	case lexer.AS:
		{
			p.iter.next()
			alias := p.expect(lexer.UPPERIDENT, withError(data.IMPORT_ALIAS))
			impor = ast.Import{Module: mod, Alias: alias.Text, Span: span(impTk.Span, p.iter.current.Span)}
		}
	}
	impor.Comment = impTk.Comment
	return impor
}

func (p *parser) parseDeclarationRefs() []ast.DeclarationRef {
	p.expect(lexer.LPAREN, withError(data.LParensExpected("import")))
	if p.iter.peek().Type == lexer.RPAREN {
		throwError(withError(data.EmptyImport("Import"))(p.iter.peek()))
	}

	exps := between(p, lexer.COMMA, func() ast.DeclarationRef { return p.parseDeclarationRef() })

	p.expect(lexer.RPAREN, withError(data.RParensExpected("import")))
	return exps
}

func (p *parser) parseDeclarationRef() ast.DeclarationRef {
	sp := p.iter.next()
	switch sp.Type {
	case lexer.IDENT:
		return ast.DeclarationRef{Tag: ast.VAR, Name: spanned(*sp.Text, sp.Span)}
	case lexer.OP:
		return ast.DeclarationRef{Tag: ast.VAR, Name: spanned(*sp.Text, sp.Span)}
	case lexer.UPPERIDENT:
		{
			binder := spanned(*sp.Text, sp.Span)
			if p.iter.peek().Type == lexer.LPAREN {
				p.expect(lexer.LPAREN, noErr())
				var ctors []ast.Spanned[string]
				all := false
				if p.iter.peek().Type == lexer.OP {
					op := p.expect(lexer.OP, withError(data.DECLARATION_REF_ALL))
					if *op.Text != ".." {
						throwError(withError(data.DECLARATION_REF_ALL)(op))
					}
					all = true
				} else {
					ctors = between(p, lexer.COMMA, func() ast.Spanned[string] {
						ident := p.expect(lexer.UPPERIDENT, withError(data.CTOR_NAME))
						return spanned(*ident.Text, ident.Span)
					})
				}
				end := p.expect(lexer.RPAREN, withError(data.DECLARATION_REF_ALL))
				return ast.DeclarationRef{Tag: ast.TYPE, Name: binder, Span: span(sp.Span, end.Span), Ctors: ctors, All: all}
			} else {
				return ast.DeclarationRef{Tag: ast.TYPE, Name: binder, Span: sp.Span}
			}
		}
	default:
		return throwError(withError(data.IMPORT_REFER)(sp)).(ast.DeclarationRef)
	}
}

func (p *parser) parseDecl() ast.SDecl {
	// meta := parseMetadata()
	tk := p.iter.peek()
	comment := tk.Comment
	var visibility *lexer.TokenType
	isInstance := false
	offside := math.MaxInt32
	if tk.Type == lexer.PUBLIC || tk.Type == lexer.PUBLICPLUS {
		vis := p.iter.next().Type
		visibility = &vis
		if tk.Offside() < offside {
			offside = tk.Offside()
		}
		tk = p.iter.peek()
	}
	if tk.Type == lexer.INSTANCE {
		p.iter.next()
		isInstance = true
		if tk.Offside() < offside {
			offside = tk.Offside()
		}
		tk = p.iter.peek()
	}
	if tk.Offside() < offside {
		offside = tk.Offside()
	}
	var decl ast.SDecl
	switch tk.Type {
	case lexer.TYPE:
		{
			if isInstance {
				throwError2(data.INSTANCE_ERROR, tk.Span)
			}
			tdecl := *p.parseTypeDecl(visibility, offside)
			tdecl.Comment = comment
			decl = tdecl
		}
	case lexer.IDENT:
		{
			vdecl := *p.parseVarDecl(visibility, isInstance, offside, false)
			vdecl.Comment = comment
			decl = vdecl
		}
	case lexer.LPAREN:
		{
			vdecl := *p.parseVarDecl(visibility, isInstance, offside, true)
			vdecl.Comment = comment
			decl = vdecl
		}
	case lexer.TYPEALIAS:
		{
			if isInstance {
				throwError2(data.INSTANCE_ERROR, tk.Span)
			}
			tdecl := *p.parseTypeAlias(visibility, offside)
			tdecl.Comment = comment
			decl = tdecl
		}
	default:
		throwError2(data.TOPLEVEL_IDENT, tk.Span)
	}
	return decl
}

func (p *parser) parseTypeDecl(visibility *lexer.TokenType, offside int) *ast.STypeDecl {
	vis := ast.PUBLIC
	if visibility == nil {
		vis = ast.PRIVATE
	}
	typ := p.expect(lexer.TYPE, noErr())
	return withOffside(p, offside+1, func() *ast.STypeDecl {
		nameTk := p.expect(lexer.UPPERIDENT, withError(data.DATA_NAME))
		name := ast.Spanned[string]{Val: *nameTk.Text, Span: nameTk.Span}

		tyVars := parseListOf(p, p.parseTypeVar, func(t lexer.Token) bool { return t.Type == lexer.IDENT })

		p.expect(lexer.EQUALS, withError(data.DATA_EQUALS))

		ctors := make([]ast.SDataCtor, 0)
		for true {
			ctors = append(ctors, *p.parseDataConstructor(tyVars, visibility))
			if p.iter.peekIsOffside() || p.iter.peek().Type == lexer.EOF {
				break
			}
			p.expect(lexer.PIPE, withError(data.PipeExpected("constructor")))
		}
		return &ast.STypeDecl{Binder: name, Visibility: vis, TyVars: tyVars, DataCtors: ctors,
			Span: span(typ.Span, p.iter.current.Span)}
	})
}

func (p *parser) parseVarDecl(visibility *lexer.TokenType, isInstance bool, offside int, isOperator bool) *ast.SValDecl {
	parseName := func(name string) ast.Spanned[string] {
		if isOperator {
			tk := p.expect(lexer.LPAREN, withError(data.INVALID_OPERATOR_DECL))
			op := p.expect(lexer.OP, withError(data.INVALID_OPERATOR_DECL))
			end := p.expect(lexer.RPAREN, withError(data.INVALID_OPERATOR_DECL))
			return ast.Spanned[string]{Val: op.Value.(string), Span: span(tk.Span, end.Span)}
		} else {
			id := p.expect(lexer.IDENT, withError(data.ExpectedDefinition(name)))
			return ast.Spanned[string]{Val: *id.Text, Span: id.Span}
		}
	}

	vis := ast.PRIVATE
	if visibility != nil {
		if *visibility == lexer.PUBLICPLUS {
			throwError2(data.PUB_PLUS, p.iter.current.Span)
		}
		vis = ast.PUBLIC
	}
	nameTk := parseName("")
	name := nameTk.Val

	return withOffside(p, offside+1, func() *ast.SValDecl {
		var sig ast.SSignature
		nameTk2 := nameTk
		if p.iter.peek().Type == lexer.COLON {
			sig = ast.SSignature{Type: p.parseTypeSignature(), Span: nameTk.Span}
			withOffside(p, nameTk.Offside(), func() any {
				nameTk2 = parseName(name)
				if name != nameTk2.Val {
					throwError2(data.ExpectedDefinition(name), nameTk2.Span)
				}
				return 0
			})
		}
		vars := tryParseListOfDef(p, func() (ast.SPattern, bool) { return p.tryParsePattern(true) })

		eq := p.expect(lexer.EQUALS, withError(data.EqualsExpected("function parameters/patterns")))

		var exp ast.SExpr
		if eq.Span.SameLine(p.iter.peek().Span) {
			exp = p.parseExpression(false)
		} else {
			exp = p.parseDo()
		}
		binder := ast.Spanned[string]{Val: name, Span: nameTk2.Span}
		if isOperator && len(name) > 3 {
			throwError2(data.OpTooLong(name), binder.Span)
		}
		return &ast.SValDecl{Binder: binder, Pats: vars, Exp: exp, Signature: &sig, Visibility: vis,
			IsInstance: isInstance, IsOperator: isOperator, Span: span(nameTk.Span, exp.GetSpan())}
	})
}

func (p *parser) parseTypeAlias(visibility *lexer.TokenType, offside int) *ast.STypeAliasDecl {
	vis := ast.PRIVATE
	if visibility != nil {
		if *visibility == lexer.PUBLICPLUS {
			throwError2(data.PUB_PLUS, p.iter.current.Span)
		}
		vis = ast.PUBLIC
	}
	p.expect(lexer.TYPEALIAS, noErr())
	name := p.expect(lexer.UPPERIDENT, withError(data.TYPEALIAS_NAME))
	return withOffside(p, offside+1, func() *ast.STypeAliasDecl {
		tyVars := tryParseListOfNil(p, false, p.tryParseTypeVar)
		end := p.iter.current.Span
		p.expect(lexer.EQUALS, withError(data.TYPEALIAS_EQUALS))
		ty := p.parseType(false)
		return &ast.STypeAliasDecl{Name: name.Value.(string),
			TyVars: tyVars, Type: ty, Visibility: vis,
			Span: span(name.Span, end),
		}
	})
}

func (p *parser) parseDataConstructor(tyVars []string, visibility *lexer.TokenType) *ast.SDataCtor {
	ctorTk := p.expect(lexer.UPPERIDENT, withError(data.CTOR_NAME))
	ctor := ast.Spanned[string]{Val: *ctorTk.Text, Span: ctorTk.Span}

	vis := ast.PRIVATE
	if visibility != nil && *visibility == lexer.PUBLICPLUS {
		vis = ast.PUBLIC
	}

	pars := tryParseListOfDef(p, func() (ast.SType, bool) { return p.parseTypeAtom(true) })

	freeVars := data.FlatMapSlice(pars, func(par ast.SType) []string { return ast.FindFreeVars(par, tyVars) })
	if len(freeVars) > 0 {
		throwError2(data.UndefinedVarInCtor(ctor.Val, freeVars), span(ctor.Span, p.iter.current.Span))
	}
	return &ast.SDataCtor{Name: ctor, Args: pars, Visibility: vis, Span: span(ctor.Span, p.iter.current.Span)}
}

func (p *parser) parseExpression(inDo bool) ast.SExpr {
	if p.iter.peekIsOffside() {
		throwMismatchedIdentation(p.iter.peek())
	}
	tk := p.iter.peek()
	exps := make([]ast.SExpr, 0, 1)
	atom, success := p.tryParseAtom()
	if !success {
		throwError2(data.MALFORMED_EXPR, tk.Span)
	}
	exps = append(exps, atom)

	offside := p.iter.offside
	if inDo {
		offside++
	}

	return withOffside(p, offside, func() ast.SExpr {
		atoms := tryParseListOfDef(p, func() (ast.SExpr, bool) { return p.tryParseAtom() })
		exps = append(exps, atoms...)

		// sanity check
		if len(exps) > 1 {
			doLets := data.FilterSliceIsInstance[ast.SExpr, ast.SDoLet](exps)
			if len(doLets) > 0 {
				throwError2(data.APPLIED_DO_LET, doLets[0].Span)
			}
		}

		unrolled := parseApplication(exps)
		if unrolled == nil {
			throwError2(data.MALFORMED_EXPR, tk.Span)
		}

		// type signatures and casts have the lowest precedence
		typedExpr := unrolled
		if p.iter.peek().Type == lexer.COLON {
			p.iter.next()
			pt := p.parseType(false)
			typedExpr = ast.SAnn{Exp: unrolled, Type: pt, Span: span(unrolled.GetSpan(), p.iter.current.Span), Comment: unrolled.GetComment()}
		}

		if p.iter.peek().Type == lexer.AS {
			p.iter.next()
			ty := p.parseType(false)
			return ast.STypeCast{Exp: typedExpr, Cast: ty, Span: span(typedExpr.GetSpan(), p.iter.current.Span), Comment: typedExpr.GetComment()}
		}
		return typedExpr
	})
}

func (p *parser) tryParseAtom() (ast.SExpr, bool) {
	var exp ast.SExpr
	switch p.iter.peek().Type {
	case lexer.INT:
		exp = p.parseInt()
	case lexer.FLOAT:
		exp = p.parseFloat()
	case lexer.COMPLEX:
		exp = p.parseComplex()
	case lexer.STRING:
		exp = p.parseString()
	case lexer.MULTILINESTRING:
		exp = p.parseMultilineString()
	case lexer.PATTERNSTRING:
		exp = p.parsePatternString()
	case lexer.CHAR:
		exp = p.parseChar()
	case lexer.BOOL:
		exp = p.parseBool()
	case lexer.IDENT:
		exp = p.parseVar()
	case lexer.OP:
		exp = p.parseOperator()
	case lexer.NIL:
		exp = p.parseNil()
	case lexer.SEMICOLON:
		{
			tk := p.iter.next()
			exp = ast.SOperator{Name: ";", IsPrefix: false, Span: tk.Span, Comment: tk.Comment}
		}
	case lexer.UNDERLINE:
		{
			tk := p.iter.next()
			under := ast.SUnderscore{Span: tk.Span, Comment: tk.Comment}
			// a _!! unwrap anonymous function
			if p.iter.peek().Type == lexer.BANGBANG {
				bb := p.iter.next()
				unwrap := ast.SVar{Name: "unwrapOption", Span: bb.Span}
				v := ast.SVar{Name: "__unw", Span: tk.Span}
				body := ast.SApp{Fn: unwrap, Arg: v, Span: span(tk.Span, bb.Span)}
				pat := ast.SVarP{V: v}
				exp = ast.SLambda{Pats: []ast.SPattern{pat}, Body: body, Span: span(tk.Span, bb.Span), Comment: tk.Comment}
			} else {
				exp = under
			}
		}
	case lexer.LPAREN:
		{
			exp = withIgnoreOffside(p, true, func() ast.SExpr {
				tk := p.iter.next()
				if p.iter.peek().Type == lexer.RPAREN {
					end := p.iter.next()
					return ast.SUnit{Span: span(tk.Span, end.Span), Comment: tk.Comment}
				}
				expr := p.parseExpression(false)
				end := p.expect(lexer.RPAREN, withError(data.RParensExpected("expression")))
				return ast.SParens{Exp: expr, Span: span(tk.Span, end.Span), Comment: tk.Comment}
			})
		}
	// an aliased operator like `MyModule.==`
	case lexer.UPPERIDENT:
		{
			uident := p.expect(lexer.UPPERIDENT, noErr())
			peek := p.iter.peek()
			if peek.Type == lexer.DOT {
				exp = p.parseAliasedVar(uident)
			} else if peek.IsDotStart() {
				op := p.expect(lexer.OP, noErr())
				name := op.Value.(string)
				exp = ast.SOperator{Name: name[1:], IsPrefix: false, Alias: uident.Text, Span: span(uident.Span, op.Span), Comment: uident.Comment}
			} else if peek.Type == lexer.HASH {
				// TODO: foreign functions
				panic("foreign functions not supported")
			} else if peek.Type == lexer.HASHDASH {
				// TODO: foreign functions
				panic("foreign functions not supported")
			} else {
				exp = ast.SConstructor{Name: *uident.Text, Span: uident.Span, Comment: uident.Comment}
			}
		}
	case lexer.BACKSLASH:
		exp = p.parseLambda()
	case lexer.IF:
		exp = p.parseIf()
	case lexer.LET:
		exp = p.parseLet()
	case lexer.LETBANG:
		exp = p.parseLetBang()
	case lexer.DODOT:
		exp = p.parseComputation()
	case lexer.DOBANG:
		{
			dobang := p.iter.next()
			exp = withOffsideDef(p, func() ast.SExpr { return p.parseExpression(false) })
			exp = ast.SDoBang{Exp: exp, Span: span(dobang.Span, exp.GetSpan()), Comment: dobang.Comment}
		}
	case lexer.CASE:
		exp = p.parseMatch()
	case lexer.LBRACKET:
		exp = p.parseRecordOrImplicit()
	case lexer.LSBRACKET:
		{
			tk := p.iter.next()
			if p.iter.peek().Type == lexer.RSBRACKET {
				end := p.iter.next()
				exp = ast.SListLiteral{Exps: []ast.SExpr{}, Span: span(tk.Span, end.Span), Comment: tk.Comment}
			} else {
				exps := between(p, lexer.COMMA, func() ast.SExpr { return p.parseExpression(false) })
				end := p.expect(lexer.RSBRACKET, withError(data.RSBracketExpected("list literal")))
				exp = ast.SListLiteral{Exps: exps, Span: span(tk.Span, end.Span), Comment: tk.Comment}
			}
		}
	case lexer.SETBRACKET:
		{
			tk := p.iter.next()
			if p.iter.peek().Type == lexer.RSBRACKET {
				end := p.iter.next()
				exp = ast.SSetLiteral{Exps: []ast.SExpr{}, Span: span(tk.Span, end.Span), Comment: tk.Comment}
			} else {
				exps := between(p, lexer.COMMA, func() ast.SExpr { return p.parseExpression(false) })
				end := p.expect(lexer.RSBRACKET, withError(data.RSBracketExpected("set literal")))
				exp = ast.SSetLiteral{Exps: exps, Span: span(tk.Span, end.Span), Comment: tk.Comment}
			}
		}
	case lexer.WHILE:
		exp = p.parseWhile()
	case lexer.RETURN:
		{
			ret := p.iter.next()
			exp := withOffsideDef(p, func() ast.SExpr { return p.parseExpression(false) })
			exp = ast.SReturn{Exp: exp, Span: span(ret.Span, exp.GetSpan()), Comment: ret.Comment}
		}
	case lexer.YIELD:
		{
			ret := p.iter.next()
			exp := withOffsideDef(p, func() ast.SExpr { return p.parseExpression(false) })
			exp = ast.SYield{Exp: exp, Span: span(ret.Span, exp.GetSpan()), Comment: ret.Comment}
		}
	case lexer.FOR:
		exp = p.parseFor()
	default:
		return nil, false
	}

	// record selection, index (.[x]), unwrap (!!) and method/field call have the highest precedence
	return p.parseSelection(exp), true
}

func (p *parser) parseSelection(exp ast.SExpr) ast.SExpr {
	switch p.iter.peek().Type {
	case lexer.DOT:
		{
			p.iter.next()
			labelsTk := between(p, lexer.DOT, func() lexer.Token { return p.parseLabel() })
			labels := data.MapSlice(labelsTk, func(t lexer.Token) ast.Spanned[string] {
				return ast.Spanned[string]{Val: t.Value.(string), Span: t.Span}
			})
			res := ast.SRecordSelect{Exp: exp, Labels: labels, Span: span(exp.GetSpan(), labels[len(labels)-1].Span)}
			return p.parseSelection(res)
		}
	case lexer.DOTBRACKET:
		{
			p.iter.next()
			index := p.parseExpression(false)
			end := p.expect(lexer.RSBRACKET, withError(data.RSBracketExpected("index")))
			res := ast.SIndex{Exp: exp, Index: index, Span: span(exp.GetSpan(), end.Span)}
			return p.parseSelection(res)
		}
	case lexer.HASH:
		{
			// TODO: support foreigns
			panic("foreigns not supported yet")
		}
	case lexer.HASHDASH:
		{
			// TODO: support foreigns
			panic("foreigns not supported yet")
		}
	case lexer.BANGBANG:
		{
			sp := p.iter.next().Span
			unwrap := ast.SVar{Name: "unwrapOption", Span: sp}
			res := ast.SApp{Fn: unwrap, Arg: exp, Span: span(exp.GetSpan(), sp), Comment: exp.GetComment()}
			return p.parseSelection(res)
		}
	default:
		return exp
	}
}

func (p *parser) parseInt() ast.SExpr {
	num := p.expect(lexer.INT, withError(data.LiteralExpected("integer")))
	v := num.Value.(int64)
	return ast.SInt{V: v, Text: *num.Text, Span: num.Span, Comment: num.Comment}
}

func (p *parser) parseFloat() ast.SExpr {
	num := p.expect(lexer.FLOAT, withError(data.LiteralExpected("float")))
	v := num.Value.(float64)
	return ast.SFloat{V: v, Text: *num.Text, Span: num.Span, Comment: num.Comment}
}

func (p *parser) parseComplex() ast.SExpr {
	num := p.expect(lexer.COMPLEX, withError(data.LiteralExpected("complex")))
	v := num.Value.(complex128)
	return ast.SComplex{V: v, Text: *num.Text, Span: num.Span, Comment: num.Comment}
}

func (p *parser) parseString() ast.SExpr {
	str := p.expect(lexer.STRING, withError(data.LiteralExpected("string")))
	return ast.SString{V: str.Value.(string), Raw: *str.Text, Span: str.Span, Comment: str.Comment}
}

func (p *parser) parseMultilineString() ast.SExpr {
	str := p.expect(lexer.MULTILINESTRING, withError(data.LiteralExpected("multiline string")))
	v := str.Value.(string)
	return ast.SString{V: v, Raw: v, Multi: true, Span: str.Span, Comment: str.Comment}
}

func (p *parser) parsePatternString() ast.SPatternLiteral {
	str := p.expect(lexer.PATTERNSTRING, withError(data.LiteralExpected("pattern string")))
	return ast.SPatternLiteral{Regex: str.Value.(string), Raw: *str.Text, Span: str.Span, Comment: str.Comment}
}

func (p *parser) parseChar() ast.SExpr {
	char := p.expect(lexer.CHAR, withError(data.LiteralExpected("char")))
	return ast.SChar{V: char.Value.(rune), Raw: *char.Text, Span: char.Span, Comment: char.Comment}
}

func (p *parser) parseBool() ast.SExpr {
	booll := p.expect(lexer.BOOL, withError(data.LiteralExpected("bool")))
	return ast.SBool{V: booll.Value.(bool), Span: booll.Span, Comment: booll.Comment}
}

func (p *parser) parseVar() ast.SVar {
	v := p.expect(lexer.IDENT, withError(data.VARIABLE))
	return ast.SVar{Name: *v.Text, Span: v.Span, Comment: v.Comment}
}

func (p *parser) parseOperator() ast.SExpr {
	v := p.expect(lexer.OP, withError(data.OPERATOR))
	return ast.SOperator{Name: *v.Text, Span: v.Span, Comment: v.Comment}
}

func (p *parser) parseNil() ast.SExpr {
	v := p.expect(lexer.NIL, noErr())
	return ast.SNil{Span: v.Span, Comment: v.Comment}
}

func (p *parser) parseAliasedVar(alias lexer.Token) ast.SExpr {
	p.expect(lexer.DOT, noErr())
	if p.iter.peek().Type == lexer.IDENT {
		ident := p.expect(lexer.IDENT, withError(data.IMPORTED_DOT))
		return ast.SVar{Name: *ident.Text, Alias: alias.Text, Span: span(alias.Span, ident.Span), Comment: alias.Comment}
	}
	ident := p.expect(lexer.UPPERIDENT, withError(data.IMPORTED_DOT))
	return ast.SConstructor{Name: *ident.Text, Alias: alias.Text, Span: span(alias.Span, ident.Span), Comment: alias.Comment}
}

func (p *parser) parseLambda() ast.SExpr {
	begin := p.iter.peek()
	p.expect(lexer.BACKSLASH, withError(data.LAMBDA_BACKSLASH))

	vars := tryParseListOfDef(p, func() (ast.SPattern, bool) { return p.tryParsePattern(true) })
	if len(vars) == 0 {
		throwError(withError(data.LAMBDA_VAR)(*p.iter.current))
	}

	arr := p.expect(lexer.ARROW, withError(data.LAMBDA_ARROW))

	var exp ast.SExpr
	if arr.Span.SameLine(p.iter.peek().Span) {
		exp = p.parseExpression(false)
	} else {
		exp = withOffsideDef(p, func() ast.SExpr { return p.parseDo() })
	}
	return ast.SLambda{Pats: vars, Body: exp, Span: span(begin.Span, exp.GetSpan()), Comment: begin.Comment}
}

func (p *parser) parseIf() ast.SExpr {
	ifTk := p.expect(lexer.IF, noErr())
	offside := p.iter.offside + 1

	hasElse := false
	var els *lexer.Token = nil
	condThens := withIgnoreOffside(p, true, func() data.Tuple[ast.SExpr, ast.SExpr] {
		cond := p.parseExpression(false)

		th := p.expect(lexer.THEN, withError(data.THEN))

		thens := withIgnoreOffside(p, false, func() ast.SExpr {
			return withOffside(p, offside, func() ast.SExpr {
				if th.Span.SameLine(p.iter.peek().Span) {
					return p.parseExpression(false)
				} else {
					return p.parseDo()
				}
			})
		})

		if p.iter.peek().Type == lexer.ELSE {
			*els = p.expect(lexer.ELSE, withError(data.ELSE))
			hasElse = true
		}
		return data.Tuple[ast.SExpr, ast.SExpr]{V1: cond, V2: thens}
	})

	var elses ast.SExpr
	if hasElse {
		if els.Span.SameLine(p.iter.peek().Span) {
			elses = p.parseExpression(false)
		} else {
			elses = p.parseDo()
		}
	}

	var end data.Span
	if elses != nil {
		end = elses.GetSpan()
	} else {
		end = condThens.V2.GetSpan()
	}
	return ast.SIf{Cond: condThens.V1, Then: condThens.V2, Else: elses, Span: span(ifTk.Span, end), Comment: ifTk.Comment}
}

func (p *parser) parseDo() ast.SExpr {
	spanned := p.iter.peek()
	if spanned.Type == lexer.DODOT {
		return p.parseExpression(false)
	}

	if p.iter.peekIsOffside() {
		throwMismatchedIdentation(spanned)
	}
	align := spanned.Offside()
	return withIgnoreOffside(p, false, func() ast.SExpr {
		return withOffside(p, align, func() ast.SExpr {
			var tk lexer.Token
			exps := make([]ast.SExpr, 0, 2)
			run := true
			for run {
				exps = append(exps, p.parseExpression(true))
				tk = p.iter.peek()
				run = !p.iter.peekIsOffside() && !statementEnding[tk.Type]
			}

			if len(exps) == 1 {
				return exps[0]
			} else {
				return ast.SDo{Exps: exps, Span: span(spanned.Span, p.iter.current.Span), Comment: spanned.Comment}
			}
		})
	})
}

func (p *parser) parseWhile() ast.SExpr {
	whil := p.expect(lexer.WHILE, noErr())
	cond := withIgnoreOffside(p, true, func() ast.SExpr {
		exp := p.parseExpression(false)
		p.expect(lexer.DO, withError(data.DO_WHILE))
		return exp
	})
	if !ast.IsSimple(&cond) {
		throwError2(data.EXP_SIMPLE, cond.GetSpan())
	}

	if p.iter.peekIsOffside() {
		throwMismatchedIdentation(p.iter.peek())
	}
	align := p.iter.peek().Offside()
	return withIgnoreOffside(p, false, func() ast.SExpr {
		return withOffside(p, align, func() ast.SExpr {
			var tk lexer.Token
			exps := make([]ast.SExpr, 0, 2)
			run := true
			for run {
				exps = append(exps, p.parseExpression(true))
				tk = p.iter.peek()
				run = !p.iter.peekIsOffside() && !statementEnding[tk.Type]
			}

			return ast.SWhile{Cond: cond, Exps: exps, Span: span(whil.Span, p.iter.current.Span), Comment: whil.Comment}
		})
	})
}

func (p *parser) parseComputation() ast.SExpr {
	doo := p.expect(lexer.DODOT, noErr())
	builder := p.parseVar()

	if p.iter.peekIsOffside() {
		throwMismatchedIdentation(p.iter.peek())
	}
	align := p.iter.peek().Offside()
	return withIgnoreOffside(p, false, func() ast.SExpr {
		return withOffside(p, align, func() ast.SExpr {
			var tk lexer.Token
			exps := make([]ast.SExpr, 0, 2)
			run := true
			for run {
				exps = append(exps, p.parseExpression(true))
				tk = p.iter.peek()
				run = !p.iter.peekIsOffside() && !statementEnding[tk.Type]
			}

			return ast.SComputation{Builder: builder, Exps: exps, Span: span(doo.Span, p.iter.current.Span), Comment: doo.Comment}
		})
	})
}

func (p *parser) parsePattern(isDestructuring bool) ast.SPattern {
	pat, success := p.tryParsePattern(isDestructuring)
	if !success {
		throwError(withError(data.PATTERN)(p.iter.peek()))
	}
	return pat
}

func (p *parser) parseLet() ast.SExpr {
	let := p.expect(lexer.LET, noErr())
	isInstance := false
	if p.iter.peek().Type == lexer.INSTANCE {
		p.iter.next()
		isInstance = true
	}

	var def ast.SLetDef
	withOffsideDef(p, func() bool {
		if isInstance || p.iter.peek().Type == lexer.IDENT {
			def = p.parseLetDefBind(isInstance)
		} else {
			def = p.parseLetDefPattern(false)
		}
		return false
	})

	if p.iter.peek().Type != lexer.IN {
		return ast.SDoLet{Def: def, Span: span(let.Span, p.iter.current.Span), Comment: let.Comment}
	}
	withIgnoreOffside(p, true, func() lexer.Token { return p.expect(lexer.IN, withError(data.LET_IN)) })

	exp := p.parseExpression(false)
	return ast.SLet{Def: def, Body: exp, Span: span(let.Span, exp.GetSpan()), Comment: let.Comment}
}

func (p *parser) parseLetBang() ast.SExpr {
	let := p.expect(lexer.LETBANG, noErr())

	def := withOffsideDef(p, func() ast.SLetPat { return p.parseLetDefPattern(false) })

	if p.iter.peek().Type != lexer.IN {
		return ast.SLetBang{Def: def, Span: span(let.Span, p.iter.current.Span), Comment: let.Comment}
	}
	withIgnoreOffside(p, true, func() lexer.Token {
		return p.expect(lexer.IN, withError(data.LET_IN))
	})

	exp := p.parseExpression(false)
	return ast.SLetBang{Def: def, Body: exp, Span: span(let.Span, exp.GetSpan()), Comment: let.Comment}
}

func (p *parser) parseFor() ast.SExpr {
	forr := p.expect(lexer.FOR, noErr())

	def := withOffsideDef(p, func() ast.SLetPat { return p.parseLetDefPattern(true) })
	withIgnoreOffside(p, true, func() lexer.Token {
		return p.expect(lexer.DO, withError(data.FOR_DO))
	})

	exp := p.parseDo()
	return ast.SFor{Def: def, Body: exp, Span: span(forr.Span, exp.GetSpan()), Comment: forr.Comment}
}

func (p *parser) parseLetDefBind(isInstance bool) ast.SLetDef {
	ident := p.expect(lexer.IDENT, withError(data.LET_DECL))
	name := *ident.Text

	var ty ast.SType
	if p.iter.peek().Type == lexer.COLON {
		ty = withOffside(p, ident.Offside()+1, func() ast.SType {
			p.iter.next()
			return p.parseType(false)
		})
		newIdent := p.expect(lexer.IDENT, withError(data.ExpectedLetDefinition(name)))
		if newIdent.Value.(string) != name {
			throwError(withError(data.ExpectedLetDefinition(name))(newIdent))
		}
	}

	vars := tryParseListOfDef(p, func() (ast.SPattern, bool) { return p.tryParsePattern(true) })
	eq := p.expect(lexer.EQUALS, withError(data.LET_EQUALS))
	var exp ast.SExpr
	if eq.Span.SameLine(p.iter.peek().Span) {
		exp = p.parseExpression(false)
	} else {
		exp = p.parseDo()
	}
	return ast.SLetBind{Expr: exp, Name: ast.SBinder{Name: name, Span: ident.Span}, Pats: vars, IsInstance: isInstance, Type: ty}
}

func (p *parser) parseLetDefPattern(isFor bool) ast.SLetPat {
	pat := p.parsePattern(true)

	var tk lexer.Token
	if isFor {
		tk = p.expect(lexer.IN, withError(data.FOR_IN))
	} else {
		tk = p.expect(lexer.EQUALS, withError(data.LET_EQUALS))
	}
	var exp ast.SExpr
	if tk.Span.SameLine(p.iter.peek().Span) {
		exp = p.parseExpression(false)
	} else {
		exp = p.parseDo()
	}

	return ast.SLetPat{Expr: exp, Pat: pat}
}

func (p *parser) parseRecordOrImplicit() ast.SExpr {
	return withIgnoreOffside(p, true, func() ast.SExpr {
		begin := p.expect(lexer.LBRACKET, noErr())
		nex := p.iter.peek()

		if nex.Type == lexer.LBRACKET {
			p.iter.next()
			var alias *string
			if p.iter.peek().Type == lexer.UPPERIDENT {
				*alias = *p.expect(lexer.UPPERIDENT, noErr()).Text
				p.expect(lexer.DOT, withError(data.ALIAS_DOT))
			}
			exp := *p.expect(lexer.IDENT, withError(data.INSTANCE_VAR)).Text
			p.expect(lexer.RBRACKET, withError(data.INSTANCE_VAR))
			end := p.expect(lexer.RBRACKET, withError(data.INSTANCE_VAR))
			return ast.SImplicitVar{Name: exp, Alias: alias, Span: span(begin.Span, end.Span), Comment: begin.Comment}
		}
		if nex.Type == lexer.RBRACKET {
			end := p.iter.next()
			return ast.SRecordEmpty{Span: span(begin.Span, end.Span), Comment: begin.Comment}
		}
		if nex.Type == lexer.DOT {
			p.iter.next()
			return p.parseRecordSetOrUpdate(begin)
		}
		if nex.Type == lexer.OP && nex.Value.(string) == "-" {
			p.iter.next()
			return p.parseRecordRestriction(begin)
		}
		if nex.Type == lexer.OP && nex.Value.(string) == "+" {
			p.iter.next()
			return p.parseRecordMerge(begin)
		}
		rows := between(p, lexer.COMMA, func() data.Entry[ast.SExpr] { return p.parseRecordRow() })
		var exp ast.SExpr
		if p.iter.peek().Type == lexer.PIPE {
			p.iter.next()
			exp = p.parseExpression(false)
		} else {
			exp = ast.SRecordEmpty{}
		}
		end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected("record")))
		return ast.SRecordExtend{Labels: data.LabelMapFrom(rows...), Exp: exp, Span: span(begin.Span, end.Span), Comment: begin.Comment}
	})
}

func (p *parser) parseRecordRow() data.Entry[ast.SExpr] {
	label := p.parseLabel()
	if p.iter.peek().Type != lexer.COLON && label.Type == lexer.IDENT {
		exp := ast.SVar{Name: label.Value.(string), Span: label.Span, Comment: label.Comment}
		return data.Entry[ast.SExpr]{Label: label.Value.(string), Val: exp}
	}
	p.expect(lexer.COLON, withError(data.RECORD_COLON))
	exp := p.parseExpression(false)
	return data.Entry[ast.SExpr]{Label: label.Value.(string), Val: exp}
}

func (p *parser) parseRecordSetOrUpdate(begin lexer.Token) ast.SExpr {
	labels := between(p, lexer.DOT, func() lexer.Token {
		return p.parseLabel()
	})
	isSet := true
	if p.iter.peek().Type == lexer.EQUALS {
		p.expect(lexer.EQUALS, withError(data.RECORD_EQUALS))
	} else {
		isSet = false
		p.expect(lexer.ARROW, withError(data.RECORD_EQUALS))
	}
	var ctx string
	if isSet {
		ctx = "record set"
	} else {
		ctx = "record update"
	}

	value := p.parseExpression(false)
	p.expect(lexer.PIPE, withError(data.PipeExpected(ctx)))
	record := p.parseExpression(false)
	end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected(ctx)))
	spanneds := data.MapSlice(labels, func(x lexer.Token) ast.Spanned[string] {
		return ast.Spanned[string]{Val: x.Value.(string), Span: x.Span}
	})
	return ast.SRecordUpdate{Exp: record, Labels: spanneds, Val: value, IsSet: isSet, Span: span(begin.Span, end.Span), Comment: begin.Comment}
}

func (p *parser) parseRecordRestriction(begin lexer.Token) ast.SExpr {
	labels := between(p, lexer.COMMA, func() lexer.Token { return p.parseLabel() })
	p.expect(lexer.PIPE, withError(data.PipeExpected("record restriction")))
	record := p.parseExpression(false)
	end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected("record restriction")))
	spanneds := data.MapSlice(labels, func(x lexer.Token) string { return x.Value.(string) })
	return ast.SRecordRestrict{Exp: record, Labels: spanneds, Span: span(begin.Span, end.Span), Comment: begin.Comment}
}

func (p *parser) parseRecordMerge(begin lexer.Token) ast.SExpr {
	exp1 := p.parseExpression(false)
	p.expect(lexer.COMMA, withError(data.CommaExpected("expression in record merge")))
	exp2 := p.parseExpression(false)
	end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected("record merge")))
	return ast.SRecordMerge{Exp1: exp1, Exp2: exp2, Span: span(begin.Span, end.Span), Comment: begin.Comment}
}

func (p *parser) parseMatch() ast.SExpr {
	caseTk := p.expect(lexer.CASE, noErr())

	exps := withIgnoreOffside(p, true, func() []ast.SExpr {
		exps := between(p, lexer.COMMA, func() ast.SExpr {
			return p.parseExpression(false)
		})
		p.expect(lexer.OF, withError(data.CASE_OF))
		return exps
	})
	arity := len(exps)

	align := p.iter.peek().Offside()
	return withIgnoreOffside(p, false, func() ast.SExpr {
		return withOffside(p, align, func() ast.SExpr {
			cases := make([]ast.SCase, 0, 1)
			first := p.parseCase()
			if len(first.Pats) != arity {
				throwError2(data.WrongArityToCase(len(first.Pats), arity), first.PatternSpan())
			}
			cases = append(cases, first)

			tk := p.iter.peek()
			for !p.iter.peekIsOffside() && !statementEnding[tk.Type] {
				cas := p.parseCase()
				if len(cas.Pats) != arity {
					throwError2(data.WrongArityToCase(len(cas.Pats), arity), cas.PatternSpan())
				}
				cases = append(cases, cas)
				tk = p.iter.peek()
			}

			return ast.SMatch{Exprs: exps, Cases: cases, Span: span(caseTk.Span, p.iter.current.Span), Comment: caseTk.Comment}
		})
	})
}

func (p *parser) parseCase() ast.SCase {
	var guard ast.SExpr
	pats := withIgnoreOffside(p, true, func() []ast.SPattern {
		pats := between(p, lexer.COMMA, func() ast.SPattern {
			return p.parsePattern(false)
		})
		if p.iter.peek().Type == lexer.IF {
			p.iter.next()
			guard = p.parseExpression(false)
		}
		p.expect(lexer.ARROW, withError(data.CASE_ARROW))
		return pats
	})
	return withOffsideDef(p, func() ast.SCase {
		return ast.SCase{Pats: pats, Exp: p.parseDo(), Guard: guard}
	})
}

func (p *parser) tryParsePattern(isDestructuring bool) (ast.SPattern, bool) {
	tk := p.iter.peek()
	var pat ast.SPattern
	switch tk.Type {
	case lexer.UNDERLINE:
		{
			p.iter.next()
			pat = ast.SWildcard{Span: tk.Span}
		}
	case lexer.IDENT:
		{
			p.iter.next()
			pat = ast.SVarP{V: ast.SVar{Name: *tk.Text, Span: tk.Span}}
		}
	case lexer.BOOL:
		{
			p.iter.next()
			pat = ast.SLiteralP{Lit: ast.SBool{V: tk.Value.(bool), Span: tk.Span}, Span: tk.Span}
		}
	case lexer.INT:
		{
			p.iter.next()
			pat = ast.SLiteralP{Lit: ast.SInt{V: tk.Value.(int64), Span: tk.Span}, Span: tk.Span}
		}
	case lexer.FLOAT:
		{
			p.iter.next()
			pat = ast.SLiteralP{Lit: ast.SFloat{V: tk.Value.(float64), Span: tk.Span}, Span: tk.Span}
		}
	case lexer.COMPLEX:
		{
			p.iter.next()
			pat = ast.SLiteralP{Lit: ast.SComplex{V: tk.Value.(complex128), Span: tk.Span}, Span: tk.Span}
		}
	case lexer.CHAR:
		{
			p.iter.next()
			pat = ast.SLiteralP{Lit: ast.SChar{V: tk.Value.(rune), Span: tk.Span}, Span: tk.Span}
		}
	case lexer.STRING:
		{
			p.iter.next()
			pat = ast.SLiteralP{Lit: ast.SString{V: tk.Value.(string), Span: tk.Span}, Span: tk.Span}
		}
	case lexer.LPAREN:
		{
			p.iter.next()
			if p.iter.peek().Type == lexer.RPAREN {
				end := p.iter.next()
				pat = ast.SUnitP{Span: span(tk.Span, end.Span)}
			} else {
				pat := p.parsePattern(false)
				end := p.expect(lexer.RPAREN, withError(data.RParensExpected("pattern declaration")))
				pat = ast.SParensP{Pat: pat, Span: span(tk.Span, end.Span)}
			}
		}
	case lexer.UPPERIDENT:
		{
			ctor := p.parseConstructor()
			if isDestructuring {
				pat = ast.SCtorP{Ctor: ctor, Fields: []ast.SPattern{}, Span: span(tk.Span, ctor.Span)}
			} else {
				fields := tryParseListOfDef(p, func() (ast.SPattern, bool) { return p.tryParsePattern(false) })
				var end data.Span
				if len(fields) == 0 {
					end = ctor.Span
				} else {
					end = fields[len(fields)-1].GetSpan()
				}
				pat = ast.SCtorP{Ctor: ctor, Fields: fields, Span: span(tk.Span, end)}
			}
		}
	case lexer.LBRACKET:
		{
			p.iter.next()
			if p.iter.peek().Type == lexer.LBRACKET {
				p.iter.next()
				pat := p.parsePattern(false)
				p.expect(lexer.RBRACKET, withError(data.INSTANCE_VAR))
				end := p.expect(lexer.RBRACKET, withError(data.INSTANCE_VAR))
				pat = ast.SImplicitP{Pat: pat, Span: span(tk.Span, end.Span)}
			} else {
				rows := between(p, lexer.COMMA, func() data.Entry[ast.SPattern] {
					return p.parsePatternRow()
				})
				end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected("record pattern"))).Span
				pat = ast.SRecordP{Labels: data.LabelMapFrom(rows...), Span: span(tk.Span, end)}
			}
		}
	case lexer.LSBRACKET:
		{
			p.iter.next()
			if p.iter.next().Type == lexer.RSBRACKET {
				end := p.iter.next().Span
				pat = ast.SListP{Elems: []ast.SPattern{}, Span: span(tk.Span, end)}
			} else {
				elems := between(p, lexer.COMMA, func() ast.SPattern { return p.parsePattern(false) })
				if p.iter.peek().IsDoubleColon() {
					p.iter.next()
					tail := p.parsePattern(false)
					end := p.expect(lexer.RSBRACKET, withError(data.RSBracketExpected("list pattern"))).Span
					pat = ast.SListP{Elems: elems, Tail: tail, Span: span(tk.Span, end)}
				} else {
					end := p.expect(lexer.RSBRACKET, withError(data.RSBracketExpected("list pattern"))).Span
					pat = ast.SListP{Elems: elems, Span: span(tk.Span, end)}
				}
			}
		}
	case lexer.PATTERNSTRING:
		pat = ast.SRegexP{Regex: p.parsePatternString()}
	case lexer.OP:
		{
			if tk.Value.(string) == ":?" {
				p.iter.next()
				tk2 := p.expect(lexer.UPPERIDENT, withError(data.TYPE_TEST_TYPE))
				ty := *tk2.Text
				var alias *string
				if p.iter.peek().Type == lexer.DOT {
					p.iter.next()
					*alias = ty
					ty = p.expect(lexer.UPPERIDENT, withError(data.TYPEALIAS_DOT)).Value.(string)
				}
				typ := ast.STConst{Name: ty, Alias: alias, Span: span(tk2.Span, p.iter.current.Span)}

				var name *string
				end := typ.Span
				if p.iter.peek().Type == lexer.AS {
					p.iter.next()
					ident := p.expect(lexer.IDENT, withError(data.VARIABLE))
					*name = *ident.Text
					end = ident.Span
				}
				pat = ast.STypeTest{Type: typ, Alias: name, Span: span(tk.Span, end)}
			} else {
				return nil, false
			}
		}
	default:
		return nil, false
	}

	// named, tuple and annotation patterns
	token := p.iter.peek().Type
	if token == lexer.AS {
		p.iter.next()
		name := p.expect(lexer.IDENT, withError(data.VARIABLE))
		return ast.SNamed{Pat: pat, Name: ast.Spanned[string]{Val: *name.Text, Span: name.Span}, Span: span(pat.GetSpan(), name.Span)}, true
	}
	if token == lexer.COLON {
		varr, isVar := pat.(ast.SVarP)
		if isVar {
			p.iter.next()
			ty := p.parseType(false)
			return ast.STypeAnnotationP{Par: varr.V, Type: ty, Span: span(pat.GetSpan(), ty.GetSpan())}, true
		} else {
			return pat, true
		}
	}
	if token == lexer.SEMICOLON {
		p.iter.next()
		p2 := p.parsePattern(isDestructuring)
		return ast.STupleP{P1: pat, P2: p2, Span: span(pat.GetSpan(), p2.GetSpan())}, true
	}
	return pat, true
}

func (p *parser) parseConstructor() ast.SConstructor {
	tk := p.expect(lexer.UPPERIDENT, noErr())
	var alias *string
	ident := tk
	if p.iter.peek().Type == lexer.DOT {
		p.iter.next()
		alias = tk.Text
		ident = p.expect(lexer.UPPERIDENT, withError(data.IMPORTED_DOT))
	}

	return ast.SConstructor{Name: *ident.Text, Alias: alias, Span: span(tk.Span, ident.Span), Comment: tk.Comment}
}

func (p *parser) parsePatternRow() data.Entry[ast.SPattern] {
	label := p.parseLabel()
	if p.iter.peek().Type != lexer.COLON && label.Type == lexer.IDENT {
		v := ast.SVar{Name: label.Value.(string), Span: label.Span}
		return data.Entry[ast.SPattern]{Label: label.Value.(string), Val: ast.SVarP{V: v}}
	}
	p.expect(lexer.COLON, withError(data.RECORD_COLON))
	pat := p.parsePattern(false)
	return data.Entry[ast.SPattern]{Label: label.Value.(string), Val: pat}
}

func (p *parser) parseTypeSignature() ast.SType {
	p.expect(lexer.COLON, withError(data.TYPE_COLON))
	return p.parseType(false)
}

func (p *parser) parseType(inCtor bool) ast.SType {
	ty, success := p.parseTypeAtom(inCtor)
	if !success {
		throwError(withError(data.TYPE_DEF)(p.iter.peek()))
	}
	return ty
}

func (p *parser) parseTypeAtom(inCtor bool) (ast.SType, bool) {
	if p.iter.peekIsOffside() {
		return nil, false
	}

	parseRowExtend := func() ast.STRowExtend {
		tk := p.iter.current.Span
		labels := between(p, lexer.COMMA, func() data.Entry[ast.SType] { return p.parseRecordTypeRow() })
		var rowInner ast.SType
		if p.iter.peek().Type == lexer.PIPE {
			p.iter.next()
			rowInner = p.parseType(false)
		} else {
			rowInner = ast.STRowEmpty{Span: span(tk, p.iter.current.Span)}
		}
		return ast.STRowExtend{Labels: data.LabelMapFrom(labels...), Row: rowInner, Span: span(tk, p.iter.current.Span)}
	}

	tk := p.iter.peek()
	var ty ast.SType
	switch tk.Type {
	case lexer.LPAREN:
		{
			p.iter.next()
			ty = withIgnoreOffside(p, true, func() ast.SType {
				typ := p.parseType(false)
				end := p.expect(lexer.RPAREN, withError(data.RParensExpected("type definition")))
				return ast.STParens{Type: typ, Span: span(tk.Span, end.Span)}
			})
		}
	case lexer.IDENT:
		ty = ast.STConst{Name: p.parseTypeVar(), Span: span(tk.Span, p.iter.current.Span)}
	case lexer.UPPERIDENT:
		{
			typ := *p.parseUpperIdent().Text
			var alias *string
			if p.iter.peek().Type == lexer.DOT {
				p.iter.next()
				*alias = typ
				typ = *p.expect(lexer.UPPERIDENT, withError(data.TYPEALIAS_DOT)).Text
			}
			if !inCtor {
				ty = ast.STConst{Name: typ, Alias: alias, Span: span(tk.Span, p.iter.current.Span)}
			} else {
				tconst := ast.STConst{Name: typ, Alias: alias, Span: span(tk.Span, p.iter.current.Span)}
				pars := tryParseListOf(p, true, func() (ast.SType, bool) { return p.parseTypeAtom(true) })

				if len(pars) == 0 {
					ty = tconst
				} else {
					ty = ast.STApp{Type: tconst, Types: pars, Span: span(tk.Span, p.iter.current.Span)}
				}
			}
		}
	case lexer.LSBRACKET:
		{
			p.iter.next()
			ty = withIgnoreOffside(p, true, func() ast.SType {
				if p.iter.peek().Type == lexer.RSBRACKET {
					end := p.iter.next()
					return ast.STRowEmpty{Span: span(tk.Span, end.Span)}
				} else {
					rextend := parseRowExtend()
					p.expect(lexer.RSBRACKET, withError(data.RSBracketExpected("row type")))
					return rextend
				}
			})
		}
	case lexer.LBRACKET:
		{
			ty = withIgnoreOffside(p, true, func() ast.SType {
				p.iter.next()
				switch p.iter.peek().Type {
				case lexer.LBRACKET:
					{
						p.iter.next()
						typ := p.parseType(false)
						p.expect(lexer.RBRACKET, withError(data.INSTANCE_TYPE))
						end := p.expect(lexer.RBRACKET, withError(data.INSTANCE_TYPE))
						return ast.STImplicit{Type: typ, Span: span(tk.Span, end.Span)}
					}
				case lexer.RBRACKET:
					{
						sp := span(tk.Span, p.iter.next().Span)
						return ast.STRecord{Row: ast.STRowEmpty{Span: sp}, Span: sp}
					}
				case lexer.PIPE:
					{
						p.iter.next()
						typ := p.parseType(false)
						end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected("record type")))
						sp := span(tk.Span, end.Span)
						return ast.STRecord{Row: ast.STRowExtend{Labels: data.EmptyLabelMap[ast.SType](), Row: typ, Span: sp}, Span: sp}
					}
				default:
					{
						rextend := parseRowExtend()
						end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected("record type")))
						return ast.STRecord{Row: rextend, Span: span(tk.Span, end.Span)}
					}
				}
			})
		}
	default:
		return nil, false
	}

	if inCtor {
		return ty, true
	}

	if p.iter.peek().Type == lexer.ARROW {
		p.iter.next()
		ret := p.parseType(false)
		return ast.STFun{Arg: ty, Ret: ret, Span: span(ty.GetSpan(), ret.GetSpan())}, true
	}
	return ty, true
}

func (p *parser) parseUpperIdent() lexer.Token {
	return p.expect(lexer.UPPERIDENT, noErr())
}

func (p *parser) tryParseTypeVar() *string {
	if p.iter.peek().Type == lexer.IDENT {
		return p.expect(lexer.IDENT, noErr()).Text
	}
	return nil
}

func (p *parser) parseTypeVar() string {
	return *p.expect(lexer.IDENT, withError(data.TYPE_VAR)).Text
}

func (p *parser) parseRecordTypeRow() data.Entry[ast.SType] {
	label := p.parseLabel()
	p.expect(lexer.COLON, withError(data.RECORD_COLON))
	ty := p.parseType(false)
	return data.Entry[ast.SType]{Label: label.Value.(string), Val: ty}
}

func (p *parser) parseLabel() lexer.Token {
	if p.iter.peek().Type == lexer.IDENT {
		tk := p.expect(lexer.IDENT, noErr())
		return tk
	} else {
		tk := p.expect(lexer.STRING, withError(data.RECORD_LABEL))
		return tk
	}
}

func (p *parser) parseMetadata() *ast.SMetadata {
	if p.iter.peek().Type != lexer.METABRACKET {
		return nil
	}

	//begin := p.expect(lexer.METABRACKET, noErr())
	//rows := p.between(lexer.COMMA)
	panic("")
}

//////////////////////////////////////////////
// helpers
//////////////////////////////////////////////

func withOffsideDef[T any](p *parser, f func() T) T {
	return withOffside(p, p.iter.offside+1, f)
}

func withOffside[T any](p *parser, off int, f func() T) T {
	tmp := p.iter.offside
	p.iter.withOffside(off)
	defer p.iter.withOffside(tmp)
	return f()
}

func withIgnoreOffside[T any](p *parser, should bool, f func() T) T {
	tmp := p.iter.ignoreOffside
	p.iter.ignoreOffside = should
	defer p.iter.withIgnoreOffside(tmp)
	return f()
}

func tryParseListOf[T any](p *parser, incOffside bool, fn func() (T, bool)) []T {
	acc := make([]T, 0)
	if p.iter.peekIsOffside() {
		return acc
	}
	elem, success := fn()
	tmp := p.iter.offside
	if incOffside {
		p.iter.offside = tmp + 1
	}
	for success {
		acc = append(acc, elem)
		if p.iter.peekIsOffside() {
			break
		}
		elem, success = fn()
	}
	p.iter.offside = tmp
	return acc
}

func tryParseListOfDef[T any](p *parser, fn func() (T, bool)) []T {
	return tryParseListOf(p, false, fn)
}

func tryParseListOfNil[T any](p *parser, incOffside bool, fn func() *T) []T {
	acc := make([]T, 0)
	if p.iter.peekIsOffside() {
		return acc
	}
	elem := fn()
	tmp := p.iter.offside
	if incOffside {
		p.iter.offside = tmp + 1
	}
	for elem != nil {
		acc = append(acc, *elem)
		if p.iter.peekIsOffside() {
			break
		}
		elem = fn()
	}
	p.iter.offside = tmp
	return acc
}

func parseListOf[T any](p *parser, parser func() T, keep func(lexer.Token) bool) []T {
	res := make([]T, 0, 1)
	for !p.iter.peekIsOffside() && keep(p.iter.peek()) {
		res = append(res, parser())
	}
	return res
}

func between[T any](p *parser, ttype lexer.TokenType, fun func() T) []T {
	res := make([]T, 0, 1)
	res = append(res, fun())
	for p.iter.peek().Type == ttype {
		p.iter.next()
		res = append(res, fun())
	}
	return res
}

func (p *parser) expect(ttype lexer.TokenType, err func(lexer.Token) data.Tuple[string, data.Span]) lexer.Token {
	tk := p.iter.next()
	if tk.Type == ttype {
		return tk
	}
	e := err(tk)
	return throwError(e).(lexer.Token)
}

func throwMismatchedIdentation(tk lexer.Token) {
	throwError(withError(data.MISMATCHED_INDENTATION)(tk))
}

func throwError(err data.Tuple[string, data.Span]) any {
	panic(ParserError{err.V1, err.V2})
}

func throwError2(msg string, span data.Span) any {
	panic(ParserError{msg, span})
}

func withError(msg string) func(lexer.Token) data.Tuple[string, data.Span] {
	return func(tk lexer.Token) data.Tuple[string, data.Span] {
		return data.Tuple[string, data.Span]{
			V1: msg,
			V2: tk.Span,
		}
	}
}

func noErr() func(lexer.Token) data.Tuple[string, data.Span] {
	return func(tk lexer.Token) data.Tuple[string, data.Span] {
		return data.Tuple[string, data.Span]{
			V1: "Cannot happen, token " + *tk.Text,
			V2: tk.Span,
		}
	}
}

// Fast forward the lexer to the next declaration
func (p *parser) fastForward() {
	peek := p.iter.peek()
	if peek.Type == lexer.EOF {
		return
	}
	ok := true
	for ok {
		p.iter.next()
		peek = p.iter.peek()
		ok = peek.Type != lexer.EOF && peek.Offside() != 1
	}
}

var statementEnding = map[lexer.TokenType]bool{
	lexer.RPAREN:    true,
	lexer.RSBRACKET: true,
	lexer.RBRACKET:  true,
	lexer.ELSE:      true,
	lexer.IN:        true,
	lexer.EOF:       true,
	lexer.COMMA:     true,
}

func span(s1 data.Span, s2 data.Span) data.Span {
	return data.NewSpan(s1, s2)
}

func spanned[T any](x T, span data.Span) ast.Spanned[T] {
	return ast.Spanned[T]{Val: x, Span: span}
}

type ParserError struct {
	msg  string
	span data.Span
}

func (err ParserError) Error() string {
	return err.msg
}

type ModuleDef struct {
	name    ast.Spanned[string]
	span    data.Span
	comment *lexer.Comment
}
