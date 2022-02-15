package parser

import (
	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
	"github.com/stackoverflow/novah-go/frontend/lexer"
)

type parser struct {
	lex        *lexer.Lexer
	sourceName string

	iter       *PeekableIterator
	moduleName *string
	errors     []ast.CompilerProblem
}

func NewParser(tokens *lexer.Lexer, sourceName string) *parser {
	return &parser{
		lex:        tokens,
		sourceName: sourceName,
		iter:       newPeekableIterator(tokens, throwMismatchedIdentation),
	}
}

func (p *parser) ParseFullModule() (res data.Result[ast.SModule]) {
	defer func() {
		if r := recover(); r != nil {
			var msg string
			var span lexer.Span
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
			err := ast.CompilerProblem{Msg: msg, Span: span, Filename: p.sourceName, Module: p.moduleName, Severity: ast.FATAL}
			res = data.Err[ast.SModule](err)
		}
	}()

	return data.Ok(p.parseFullModule())
}

func (p *parser) parseFullModule() ast.SModule {
	mdef := p.parseModule()
	p.moduleName = &mdef.name.Val

	var imports []ast.SImport
	next := p.iter.peek().Type
	// TODO: parse foreign imports
	if next == lexer.IMPORT {
		imports = p.parseImports()
	}
	// TODO: add default imports

	return ast.SModule{
		Name:       mdef.name,
		SourceName: p.sourceName,
		Imports:    imports,
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
			throwError(data.Tuple[string, lexer.Span]{V1: data.MODULE_NAME, V2: it.Span})
		}
	}
	span := span(idents[0].Span, idents[len(idents)-1].Span)
	name := data.JoinToString(idents, ".", func(t lexer.Token) string {
		return *t.Text
	})
	return ast.Spanned[string]{Val: name, Span: span}
}

func (p *parser) parseImports() []ast.SImport {
	imps := make([]ast.SImport, 0, 2)
	for {
		if p.iter.peek().Type == lexer.IMPORT {
			imps = append(imps, p.parseImport())
		} else {
			break
		}
	}
	return imps
}

func (p *parser) parseImport() ast.SImport {
	impTk := p.expect(lexer.IMPORT, noErr())
	mod := p.parseModuleName()
	var impor ast.SImport
	switch p.iter.peek().Type {
	case lexer.LPAREN:
		{
			imp := p.parseDeclarationRefs()
			if p.iter.peek().Type == lexer.AS {
				p.iter.next()
				alias := p.expect(lexer.UPPERIDENT, withError(data.IMPORT_ALIAS))
				impor = ast.SImport{Module: mod, Defs: &imp, Alias: alias.Text, Span: span(impTk.Span, p.iter.current.Span)}
			} else {
				impor = ast.SImport{Module: mod, Defs: &imp, Span: span(impTk.Span, p.iter.current.Span)}
			}
		}
	case lexer.AS:
		{
			p.iter.next()
			alias := p.expect(lexer.UPPERIDENT, withError(data.IMPORT_ALIAS))
			impor = ast.SImport{Module: mod, Alias: alias.Text, Span: span(impTk.Span, p.iter.current.Span)}
		}
	}
	impor.Comment = impTk.Comment
	return impor
}

func (p *parser) parseDeclarationRefs() []ast.SDeclarationRef {
	p.expect(lexer.LPAREN, withError(data.LParensExpected("import")))
	if p.iter.peek().Type == lexer.RPAREN {
		throwError(withError(data.EmptyImport("Import"))(p.iter.peek()))
	}

	exps := between(p, lexer.COMMA, func() ast.SDeclarationRef {
		return p.parseDeclarationRef()
	})

	p.expect(lexer.RPAREN, withError(data.RParensExpected("import")))
	return exps
}

func (p *parser) parseDeclarationRef() ast.SDeclarationRef {
	sp := p.iter.next()
	switch sp.Type {
	case lexer.IDENT:
		return ast.SDeclarationRef{Tag: ast.VAR, Name: spanned(*sp.Text, sp.Span)}
	case lexer.OP:
		return ast.SDeclarationRef{Tag: ast.VAR, Name: spanned(*sp.Text, sp.Span)}
	case lexer.UPPERIDENT:
		{
			binder := spanned(*sp.Text, sp.Span)
			if p.iter.peek().Type == lexer.LPAREN {
				p.expect(lexer.LPAREN, noErr())
				var ctors []ast.Spanned[string]
				if p.iter.peek().Type == lexer.OP {
					op := p.expect(lexer.OP, withError(data.DECLARATION_REF_ALL))
					if *op.Text == ".." {
						throwError(withError(data.DECLARATION_REF_ALL)(op))
					}
				} else {
					ctors = between(p, lexer.COMMA, func() ast.Spanned[string] {
						ident := p.expect(lexer.UPPERIDENT, withError(data.CTOR_NAME))
						return spanned(*ident.Text, ident.Span)
					})
				}
				end := p.expect(lexer.RPAREN, withError(data.DECLARATION_REF_ALL))
				return ast.SDeclarationRef{Tag: ast.TYPE, Name: binder, Span: span(sp.Span, end.Span), Ctors: &ctors}
			} else {
				return ast.SDeclarationRef{Tag: ast.TYPE, Name: binder, Span: sp.Span}
			}
		}
	default:
		return throwError(withError(data.IMPORT_REFER)(sp)).(ast.SDeclarationRef)
	}
}

func (p *parser) parseExpression(inDo bool) ast.SExpr {

	panic("")
}

func (p *parser) tryParseAtom() *ast.SExpr {
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
	}

	return &exp
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
	v := p.expect(lexer.IDENT, withError(data.OPERATOR))
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

	vars := tryParseListOf(p, false, func() *ast.SPattern {
		return p.tryParsePattern(true)
	})
	if len(vars) == 0 {
		throwError(withError(data.LAMBDA_VAR)(*p.iter.current))
	}

	arr := p.expect(lexer.ARROW, withError(data.LAMBDA_ARROW))

	var exp ast.SExpr
	if arr.Span.SameLine(p.iter.peek().Span) {
		exp = p.parseExpression(false)
	} else {
		exp = withOffsideDef(p, func() ast.SExpr {
			return p.parseDo()
		})
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

	elses := data.None[ast.SExpr]()
	if hasElse {
		if els.Span.SameLine(p.iter.peek().Span) {
			elses = data.Some(p.parseExpression(false))
		} else {
			elses = data.Some(p.parseDo())
		}
	}

	var end lexer.Span
	if !elses.IsEmpty() {
		end = elses.Value().GetSpan()
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
		throwError(data.Tuple[string, lexer.Span]{V1: data.EXP_SIMPLE, V2: cond.GetSpan()})
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
	pat := p.tryParsePattern(isDestructuring)
	if pat == nil {
		throwError(withError(data.PATTERN)(p.iter.peek()))
	}
	return *pat
}

func (p *parser) tryParsePattern(isDestructuring bool) *ast.SPattern {
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
				fields := tryParseListOf(p, false, func() *ast.SPattern {
					return p.tryParsePattern(false)
				})
				var end lexer.Span
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
				rows := between(p, lexer.COMMA, func() data.Tuple[string, ast.SPattern] {
					return p.parsePatternRow()
				})
				end := p.expect(lexer.RBRACKET, withError(data.RBracketExpected("record pattern"))).Span
				pat = ast.SRecordP{Labels: rows, Span: span(tk.Span, end)}
			}
		}
	case lexer.LSBRACKET:
		{
			p.iter.next()
			if p.iter.next().Type == lexer.RSBRACKET {
				end := p.iter.next().Span
				pat = ast.SListP{Elems: []ast.SPattern{}, Span: span(tk.Span, end)}
			} else {
				elems := between(p, lexer.COMMA, func() ast.SPattern {
					return p.parsePattern(false)
				})
				if p.iter.peek().IsDoubleColon() {
					p.iter.next()
					tail := p.parsePattern(false)
					end := p.expect(lexer.RSBRACKET, withError(data.RSBracketExpected("list pattern"))).Span
					pat = ast.SListP{Elems: elems, Tail: &tail, Span: span(tk.Span, end)}
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
				return nil
			}
		}
	default:
		return nil
	}

	// named, tuple and annotation patterns
	token := p.iter.peek().Type
	if token == lexer.AS {
		p.iter.next()
		name := p.expect(lexer.IDENT, withError(data.VARIABLE))
		var res ast.SPattern = ast.SNamed{Pat: pat, Name: ast.Spanned[string]{Val: *name.Text, Span: name.Span}, Span: span(pat.GetSpan(), name.Span)}
		return &res
	}
	if token == lexer.COLON {
		varr, isVar := pat.(ast.SVarP)
		if isVar {
			p.iter.next()
			ty := p.parseType(false)
			var res ast.SPattern = ast.STypeAnnotationP{Par: varr.V, Type: ty, Span: span(pat.GetSpan(), ty.GetSpan())}
			return &res
		} else {
			return &pat
		}
	}
	if token == lexer.SEMICOLON {
		p.iter.next()
		p2 := p.parsePattern(isDestructuring)
		var res ast.SPattern = ast.STupleP{P1: pat, P2: p2, Span: span(pat.GetSpan(), p2.GetSpan())}
		return &res
	}
	return &pat
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

func (p *parser) parsePatternRow() data.Tuple[string, ast.SPattern] {
	label := p.parseLabel()
	if p.iter.peek().Type != lexer.COLON && label.Type == lexer.IDENT {
		v := ast.SVar{Name: label.Value.(string), Span: label.Span}
		return data.Tuple[string, ast.SPattern]{V1: label.Value.(string), V2: ast.SVarP{V: v}}
	}
	p.expect(lexer.COLON, withError(data.RECORD_COLON))
	pat := p.parsePattern(false)
	return data.Tuple[string, ast.SPattern]{V1: label.Value.(string), V2: pat}
}

func (p *parser) parseTypeSignature() ast.SType {
	p.expect(lexer.COLON, withError(data.TYPE_COLON))
	return p.parseType(false)
}

func (p *parser) parseType(inCtor bool) ast.SType {
	ty := p.parseTypeAtom(inCtor)
	if ty == nil {
		throwError(withError(data.TYPE_DEF)(p.iter.peek()))
	}
	return *ty
}

func (p *parser) parseTypeAtom(inCtor bool) *ast.SType {
	if p.iter.peekIsOffside() {
		return nil
	}

	parseRowExtend := func() ast.STRowExtend {
		tk := p.iter.current.Span
		labels := between(p, lexer.COMMA, func() data.Tuple[string, ast.SType] {
			return p.parseRecordTypeRow()
		})
		var rowInner ast.SType
		if p.iter.peek().Type == lexer.PIPE {
			p.iter.next()
			rowInner = p.parseType(false)
		} else {
			rowInner = ast.STRowEmpty{Span: span(tk, p.iter.current.Span)}
		}
		return ast.STRowExtend{Labels: labels, Row: rowInner, Span: span(tk, p.iter.current.Span)}
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
			alias := data.None[string]()
			if p.iter.peek().Type == lexer.DOT {
				p.iter.next()
				alias = data.Some(typ)
				typ = *p.expect(lexer.UPPERIDENT, withError(data.TYPEALIAS_DOT)).Text
			}
			if !inCtor {
				ty = ast.STConst{Name: typ, Alias: alias.ValueOrNil(), Span: span(tk.Span, p.iter.current.Span)}
			} else {
				tconst := ast.STConst{Name: typ, Alias: alias.ValueOrNil(), Span: span(tk.Span, p.iter.current.Span)}
				pars := tryParseListOf(p, true, func() *ast.SType {
					return p.parseTypeAtom(true)
				})

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
						return ast.STRecord{Row: ast.STRowExtend{Labels: ast.EmptyLabelMap[ast.SType](), Row: typ, Span: sp}, Span: sp}
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
		return nil
	}

	if inCtor {
		return &ty
	}

	if p.iter.peek().Type == lexer.ARROW {
		p.iter.next()
		ret := p.parseType(false)
		var fun ast.SType = ast.STFun{Arg: ty, Ret: ret, Span: span(ty.GetSpan(), ret.GetSpan())}
		return &fun
	}
	return &ty
}

func (p *parser) parseUpperIdent() lexer.Token {
	return p.expect(lexer.UPPERIDENT, noErr())
}

func (p *parser) parseTypeVar() string {
	return *p.expect(lexer.IDENT, withError(data.TYPE_VAR)).Text
}

func (p *parser) parseRecordTypeRow() data.Tuple[string, ast.SType] {
	label := p.parseLabel()
	p.expect(lexer.COLON, withError(data.RECORD_COLON))
	ty := p.parseType(false)
	return data.Tuple[string, ast.SType]{V1: label.Value.(string), V2: ty}
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

	begin := p.expect(lexer.METABRACKET, noErr())
	//rows := p.between(lexer.COMMA)
	println(begin)
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

func tryParseListOf[T any](p *parser, incOffside bool, fn func() *T) []T {
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

func between[T any](p *parser, ttype lexer.TokenType, fun func() T) []T {
	res := make([]T, 0, 1)
	res = append(res, fun())
	for p.iter.peek().Type == ttype {
		p.iter.next()
		res = append(res, fun())
	}
	return res
}

func (p *parser) expect(ttype lexer.TokenType, err func(lexer.Token) data.Tuple[string, lexer.Span]) lexer.Token {
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

func throwError(err data.Tuple[string, lexer.Span]) any {
	panic(ParserError{err.V1, err.V2})
}

func withError(msg string) func(lexer.Token) data.Tuple[string, lexer.Span] {
	return func(tk lexer.Token) data.Tuple[string, lexer.Span] {
		return data.Tuple[string, lexer.Span]{
			V1: msg,
			V2: tk.Span,
		}
	}
}

func noErr() func(lexer.Token) data.Tuple[string, lexer.Span] {
	return func(tk lexer.Token) data.Tuple[string, lexer.Span] {
		return data.Tuple[string, lexer.Span]{
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

func span(s1 lexer.Span, s2 lexer.Span) lexer.Span {
	return lexer.NewSpan(s1, s2)
}

func spanned[T any](x T, span lexer.Span) ast.Spanned[T] {
	return ast.Spanned[T]{Val: x, Span: span}
}

type ParserError struct {
	msg  string
	span lexer.Span
}

func (err ParserError) Error() string {
	return err.msg
}

type ModuleDef struct {
	name    ast.Spanned[string]
	span    lexer.Span
	comment *lexer.Comment
}
