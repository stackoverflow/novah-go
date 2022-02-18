package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Pos struct {
	Line int
	Col  int
}

type Span struct {
	Start Pos
	End   Pos
}

func NewSpan(s1 Span, s2 Span) Span {
	return Span{s1.Start, s2.End}
}

func NewSpan2(ls int, cs int, le int, ce int) Span {
	return Span{Pos{ls, cs}, Pos{le, ce}}
}

func (s Span) String() string {
	return fmt.Sprintf("%d:%d - %d:%d", s.Start.Line, s.Start.Col, s.End.Line, s.End.Col)
}

// Returns true if there's no lines between these spans
func (s Span) adjacent(other Span) bool {
	return s.End.Line+1 == other.Start.Line
}

// Returns true if this ends on the same line as other starts
func (s Span) SameLine(other Span) bool {
	return s.End.Line == other.Start.Line
}

type TokenType int

const (
	LPAREN TokenType = iota
	RPAREN
	LSBRACKET
	RSBRACKET
	LBRACKET
	RBRACKET
	SETBRACKET
	METABRACKET
	HASH
	HASHDASH
	DOT
	DOTBRACKET
	COMMA
	COLON
	SEMICOLON
	EQUALS
	BACKSLASH
	ARROW
	UNDERLINE
	PIPE

	MODULE
	IMPORT
	TYPE
	TYPEALIAS
	AS
	IF
	THEN
	ELSE
	LET
	LETBANG
	BANGBANG
	CASE
	OF
	IN
	DO
	DODOT
	DOBANG
	FOREIGN
	PUBLIC
	PUBLICPLUS
	INSTANCE
	WHILE
	NIL
	RETURN
	YIELD
	FOR

	BOOL
	CHAR
	STRING
	MULTILINESTRING
	PATTERNSTRING
	INT
	FLOAT
	COMPLEX

	IDENT
	UPPERIDENT
	OP

	EOF
)

type Comment struct {
	Text    string
	Span    Span
	IsMulti bool
}

type Token struct {
	Type    TokenType
	Span    Span
	Text    *string
	Value   any
	Comment *Comment
}

func (t Token) Offside() int {
	return t.Span.Start.Col
}

func (t Token) IsDotStart() bool {
	return t.Type == OP && t.Value.(string)[0] == '.'
}

func (t Token) IsDoubleColon() bool {
	return t.Type == OP && t.Value.(string) == "::"
}

type Lexer struct {
	Name   string
	buffer *bufio.Reader
	peeked *rune
	line   int
	col    int
}

func New(name string, reader io.Reader) *Lexer {
	return &Lexer{
		Name:   name,
		buffer: bufio.NewReader(reader),
		line:   1,
		col:    1,
	}
}

func (tk Token) IsEOF() bool {
	return tk.Type == EOF
}

func (lex *Lexer) HasMore() bool {
	_, err := lex.peek()
	return err != io.EOF
}

func (lex *Lexer) Scan() Token {
	lex.consumeWhiteSpace()

	startLine := lex.line
	startCol := lex.col

	if !lex.HasMore() {
		span := Span{Pos{startLine, startCol}, Pos{startLine, startCol}}
		return Token{Type: EOF, Span: span}
	}

	var comment *Comment

	c := lex.next()

	if c == '/' {
		peeked := lex.peekNoErr()
		if peeked == '/' {
			comm := lex.lineComment()
			span := Span{Pos{startLine, startCol}, Pos{lex.line, lex.col}}
			lex.consumeWhiteSpace()
			nex := lex.Scan()
			nextComm := nex.Comment
			// concat // comments
			if nextComm != nil && !nextComm.IsMulti && span.adjacent(nextComm.Span) {
				newComm := Comment{Text: fmt.Sprintf("%s\n%s", comm, nextComm.Text), Span: NewSpan(span, nextComm.Span)}
				nex.Comment = &newComm
				return nex
			}
			// if there's a blank line between the comment the definition don't add
			if nex.Comment != nil || !span.adjacent(nex.Span) {
				return nex
			}
			comment = &Comment{comm, span, false}
			nex.Comment = comment
			return nex
		}
		if peeked == '*' {
			comm := lex.multilineComment()
			span := Span{Pos{startLine, startCol}, Pos{lex.line, lex.col}}
			lex.consumeWhiteSpace()
			nex := lex.Scan()
			// if there's a blank line between the comment the definition don't add
			if nex.Comment != nil || !span.adjacent(nex.Span) {
				return nex
			}
			comment = &Comment{comm, span, true}
			nex.Comment = comment
			return nex
		}
	}

	var token Token
	switch c {
	case '(':
		token = Token{Type: LPAREN}
	case ')':
		token = Token{Type: RPAREN}
	case '[':
		token = Token{Type: LSBRACKET}
	case ']':
		token = Token{Type: RSBRACKET}
	case '{':
		token = Token{Type: LBRACKET}
	case '}':
		token = Token{Type: RBRACKET}
	case ',':
		token = Token{Type: COMMA}
	case ';':
		token = Token{Type: SEMICOLON}
	case '\\':
		token = Token{Type: BACKSLASH}
	case '#':
		{
			switch lex.peekNoErr() {
			case '}':
				{
					lex.next()
					token = Token{Type: SETBRACKET}
				}
			case '[':
				{
					lex.next()
					token = Token{Type: METABRACKET}
				}
			case '-':
				{
					lex.next()
					token = Token{Type: HASHDASH}
				}
			case '"':
				{
					lex.next()
					token = Token{Type: PATTERNSTRING}
				}
			default:
				token = Token{Type: HASH}
			}
		}
	case '\'':
		token = lex.char()
	case '"':
		token = lex.string()
	case '-':
		{
			tk := lex.peekNoErr()
			if unicode.IsDigit(tk) {
				lex.next()
				token = lex.number(tk, true)
			} else {
				token = lex.operator(c)
			}
		}
	case '`':
		{
			str := lex.backtickOperator()
			token = Token{Type: OP, Value: str, Text: &str}
		}
	case '_':
		if lex.HasMore() && validIdentifier(lex.peekNoErr()) {
			token = lex.ident(&c)
		} else {
			token = Token{Type: UNDERLINE}
		}
	default:
		{
			if unicode.IsDigit(c) {
				token = lex.number(c, false)
			} else if operators[c] {
				token = lex.operator(c)
			} else if validIdentifierStart(c) {
				token = lex.ident(&c)
			} else {
				lex.lexError("Unexpected Identifier:: " + string(c))
			}
		}
	}

	token.Span = Span{Pos{startLine, startCol}, Pos{lex.line, lex.col}}
	token.Comment = comment
	return token
}

func (lex *Lexer) ident(init *rune) Token {
	var sb strings.Builder
	if init != nil {
		sb.WriteRune(*init)
	}
	hasOpEnd := false

	for lex.HasMore() && validIdentifier(lex.peekNoErr()) {
		sb.WriteRune(lex.next())
	}

	if lex.HasMore() && (lex.peekNoErr() == '?' || lex.peekNoErr() == '!') {
		sb.WriteRune(lex.next())
		hasOpEnd = true
	}

	str := sb.String()
	switch str {
	case "":
		lex.lexError("Identifiers cannot be empty")
	case "true":
		return Token{Type: BOOL, Value: true, Text: &str}
	case "false":
		return Token{Type: BOOL, Value: false, Text: &str}
	case "if":
		return Token{Type: IF}
	case "then":
		return Token{Type: THEN}
	case "else":
		return Token{Type: ELSE}
	case "_":
		return Token{Type: UNDERLINE}
	case "module":
		return Token{Type: MODULE}
	case "import":
		return Token{Type: IMPORT}
	case "case":
		return Token{Type: CASE}
	case "of":
		return Token{Type: OF}
	case "type":
		return Token{Type: TYPE}
	case "typealias":
		return Token{Type: TYPEALIAS}
	case "as":
		return Token{Type: AS}
	case "in":
		return Token{Type: IN}
	case "foreign":
		return Token{Type: FOREIGN}
	case "instance":
		return Token{Type: INSTANCE}
	case "while":
		return Token{Type: WHILE}
	case "nil":
		return Token{Type: NIL}
	case "return":
		return Token{Type: RETURN}
	case "yield":
		return Token{Type: YIELD}
	case "for":
		return Token{Type: FOR}
	case "do":
		{
			if lex.HasMore() && lex.peekNoErr() == '.' {
				lex.next()
				return Token{Type: DODOT}
			}
			return Token{Type: DO}
		}
	case "do!":
		return Token{Type: DOBANG}
	case "let":
		return Token{Type: LET}
	case "let!":
		return Token{Type: LETBANG}
	case "pub":
		{
			if lex.HasMore() && lex.peekNoErr() == '+' {
				lex.next()
				return Token{Type: PUBLICPLUS}
			}
			return Token{Type: PUBLIC}
		}
	}

	if len(str) >= 2 && str[0:2] == "__" {
		lex.lexError(fmt.Sprintf("Identifiers cannot start with a double underscore (__)."))
	}

	if IsUpper(str) {
		if hasOpEnd {
			lex.lexError("Upper case identifiers cannot end with `?` or `!`.")
		}
		return Token{Type: UPPERIDENT, Value: str, Text: &str}
	}
	return Token{Type: IDENT, Value: str, Text: &str}
}

func (lex *Lexer) operator(init rune) Token {
	isOp := func(r rune) bool {
		return operators[r]
	}
	var sb strings.Builder
	sb.WriteRune(init)
	nex := lex.acceptFunc(isOp)
	for nex != nil {
		sb.WriteRune(*nex)
		if sb.String() == "!!" && lex.HasMore() && lex.peekNoErr() == '.' {
			return Token{Type: BANGBANG}
		}
		nex = lex.acceptFunc(isOp)
	}

	str := sb.String()
	switch str {
	case "=":
		return Token{Type: EQUALS}
	case "->":
		return Token{Type: ARROW}
	case "|":
		return Token{Type: PIPE}
	case ":":
		return Token{Type: COLON}
	case "!!":
		return Token{Type: BANGBANG}
	case ".":
		{
			if lex.HasMore() && lex.peekNoErr() == '[' {
				lex.next()
				return Token{Type: DOTBRACKET}
			}
			return Token{Type: DOT}
		}
	default:
		return Token{Type: OP, Value: str, Text: &str}
	}
}

func (lex *Lexer) backtickOperator() string {
	var sb strings.Builder
	c := lex.next()
	for c != '`' {
		if escapes[c] {
			lex.lexError("Invalid character in backtick operator.")
		}
		sb.WriteRune(c)
		c = lex.next()
	}
	return sb.String()
}

var escapes map[rune]bool = map[rune]bool{'t': true, 'n': true, 'r': true, 'f': true, 'b': true}

func (lex *Lexer) number(init rune, neg bool) Token {
	var sb strings.Builder

	if neg {
		sb.WriteString("-")
	}
	sb.WriteRune(init)

	rest := lex.acceptManyStr("0123456789abcdefABCDEFoOxXbBeEpPi-+.")
	sb.WriteString(rest)

	str := sb.String()
	var tk TokenType
	var v any
	var err error
	if strings.ContainsRune(str, 'i') {
		v, err = strconv.ParseComplex(str, 128)
		tk = COMPLEX
	} else if strings.ContainsAny(str, ".eE") {
		v, err = strconv.ParseFloat(str, 64)
		tk = FLOAT
	} else {
		v, err = strconv.ParseInt(str, 0, 64)
		tk = INT
	}

	if err != nil {
		lex.lexError("Invalid number " + str)
	}
	return Token{Type: tk, Value: v, Text: &str}
}

func (lex *Lexer) string() Token {
	var sb strings.Builder
	var raw strings.Builder
	c := lex.next()
	if c == '"' {
		if lex.HasMore() && lex.peekNoErr() != '"' {
			str := ""
			return Token{Type: STRING, Value: str, Text: &str}
		}
		lex.next()
		return lex.multilineString()
	}

	for c != '"' {
		if c == '\n' {
			lex.lexError("Newline is not allowed inside strings.")
		}
		if c == '\\' {
			esc, str := lex.readEscapes()
			c = esc
			raw.WriteString(str)
		} else {
			raw.WriteRune(c)
		}
		sb.WriteRune(c)
		c = lex.next()
	}
	str := sb.String()
	rawStr := raw.String()
	return Token{Type: STRING, Value: str, Text: &rawStr}
}

func (lex *Lexer) multilineString() Token {
	var sb strings.Builder
	last1 := ' '
	last0 := ' '
	c := lex.next()
	for c != '"' || last1 != '"' || last0 != '"' {
		sb.WriteRune(c)
		last1 = last0
		last0 = c
		c = lex.next()
	}
	rawStr := sb.String()
	str := rawStr[:len(rawStr)-2]
	return Token{Type: MULTILINESTRING, Value: str, Text: &str}
}

func (lex *Lexer) char() Token {
	c := lex.next()
	var token Token
	if c == '\\' {
		esc, str := lex.readEscapes()
		token = Token{Type: CHAR, Value: esc, Text: &str}
	} else {
		str := string(c)
		token = Token{Type: CHAR, Value: c, Text: &str}
	}
	n := lex.next()
	if n != '\'' {
		lex.lexError("Expected ' after char literal")
	}
	return token
}

func (lex *Lexer) lineComment() string {
	var sb strings.Builder
	lex.acceptMany('/')

	for lex.HasMore() && lex.peekNoErr() != '\n' {
		sb.WriteRune(lex.next())
	}
	return sb.String()
}

func (lex *Lexer) multilineComment() string {
	stars := lex.acceptMany('*')
	if len(stars) > 1 && lex.peekNoErr() == '/' {
		lex.next()
		return ""
	}

	var sb strings.Builder
	last := ' '
	for {
		c := lex.next()
		if last == '*' && c == '/' {
			break
		}
		sb.WriteRune(c)
		last = c
	}
	str := sb.String()
	return str[0 : len(str)-1]
}

var validEscapes map[rune]bool = map[rune]bool{'t': true, '\\': true, 'n': true, 'r': true, 'f': true, 'b': true, 'u': true}

func (lex *Lexer) readEscapes() (rune, string) {
	c := lex.next()
	if !validEscapes[c] {
		lex.lexError("Unexpected UTF-8 escape character")
	}
	switch c {
	case 'n':
		return '\n', "\\n"
	case 't':
		return '\t', "\\t"
	case '\\':
		return '\\', "\\\\"
	case 'r':
		return '\r', "\\r"
	case 'f':
		return '\f', "\\f"
	case 'b':
		return '\b', "\\b"
	case 'u':
		{
			chars := "0123456789abcdefABCDEF"
			u1 := lex.accept(chars)
			u2 := lex.accept(chars)
			u3 := lex.accept(chars)
			u4 := lex.accept(chars)
			if u1 == nil || u2 == nil || u3 == nil || u4 == nil {
				lex.lexError("Unexpected UTF-8 escape character ")
			}
			str := fmt.Sprintf("%c%c%c%c", *u1, *u2, *u3, *u4)
			u, _ := strconv.ParseInt(str, 16, 32)
			return rune(u), "\\u" + str
		}
	default:
		return c, string(c)
	}
}

func (lex *Lexer) acceptMany(r rune) string {
	var sb strings.Builder
	for lex.HasMore() && lex.peekNoErr() == r {
		sb.WriteRune(lex.next())
	}
	return sb.String()
}

func (lex *Lexer) acceptManyStr(str string) string {
	var sb strings.Builder
	for lex.HasMore() && strings.ContainsRune(str, lex.peekNoErr()) {
		sb.WriteRune(lex.next())
	}
	return sb.String()
}

func (lex *Lexer) accept(str string) *rune {
	if !lex.HasMore() {
		return nil
	}
	c := lex.next()
	if strings.ContainsRune(str, c) {
		return &c
	}
	return nil
}

func (lex *Lexer) acceptFunc(f func(rune) bool) *rune {
	if !lex.HasMore() {
		return nil
	}
	c := lex.next()
	if f(c) {
		return &c
	}
	return nil
}

func (lex *Lexer) consumeWhiteSpace() {
	for lex.HasMore() {
		p, _ := lex.peek()
		if unicode.IsSpace(p) {
			lex.next()
		} else {
			break
		}
	}
}

func (lex *Lexer) lexError(msg string) {
	span := Span{Pos{lex.line, lex.col}, Pos{lex.line, lex.col}}
	panic(LexerError{msg, span})
}

func (lex *Lexer) peek() (rune, error) {
	if lex.peeked != nil {
		return *lex.peeked, nil
	}
	rune, _, err := lex.buffer.ReadRune()
	if err == nil {
		lex.peeked = &rune
	}
	return rune, err
}

func (lex *Lexer) peekNoErr() rune {
	r, _ := lex.peek()
	return r
}

func (lex *Lexer) next() rune {
	if lex.peeked != nil {
		rune := *lex.peeked
		lex.peeked = nil
		return lex.advancePos(rune)
	}

	rune, _, err := lex.buffer.ReadRune()
	if err != nil {
		panic(fmt.Sprintf("Could not read from file %s", lex.Name))
	}
	return lex.advancePos(rune)
}

func (lex *Lexer) advancePos(r rune) rune {
	if r == '\n' {
		lex.line++
		lex.col = 1
	} else {
		lex.col++
	}
	return r
}

func IsUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func validIdentifierStart(char rune) bool {
	return unicode.IsLetter(char) || char == '_'
}

func validIdentifier(char rune) bool {
	return validIdentifierStart(char) || unicode.IsDigit(char)
}

var operators map[rune]bool = map[rune]bool{
	'$': true,
	'=': true,
	'<': true,
	'>': true,
	'|': true,
	'&': true,
	'+': true,
	'-': true,
	':': true,
	'*': true,
	'/': true,
	'%': true,
	'^': true,
	'.': true,
	'?': true,
	'!': true,
}

type LexerError struct {
	Msg  string
	Span Span
}

func (err LexerError) Error() string {
	return err.Msg
}
