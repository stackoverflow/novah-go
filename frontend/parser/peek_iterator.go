package parser

import "github.com/stackoverflow/novah-go/frontend/lexer"

type PeekableIterator struct {
	lexer   lexer.Lexer
	onError func(*lexer.Token)

	lookahead     *lexer.Token
	current       *lexer.Token
	offside       int // keep track of offside rules
	ignoreOffside bool
}

func newPeekableIterator(lex lexer.Lexer, onError func(*lexer.Token)) PeekableIterator {
	return PeekableIterator{
		lexer:         lex,
		onError:       onError,
		lookahead:     nil,
		current:       nil,
		offside:       1,
		ignoreOffside: false,
	}
}

func (it PeekableIterator) hasNext() bool {
	return it.lookahead != nil || it.lexer.HasMore()
}

func (it PeekableIterator) next() lexer.Token {
	var t lexer.Token
	if it.lookahead != nil {
		tmp := it.lookahead
		it.lookahead = nil
		t = *tmp
	} else {
		t = it.lexer.Scan()
	}
	if !it.ignoreOffside && t.Offside() < it.offside {
		it.onError(&t)
	}
	it.current = &t
	return t
}

func (it PeekableIterator) peek() lexer.Token {
	if it.lookahead == nil {
		n := it.next()
		it.lookahead = &n
	}
	return *it.lookahead
}

func (it PeekableIterator) peekIsOffside() bool {
	return !it.ignoreOffside && it.peek().Offside() < it.offside
}

func (it PeekableIterator) Current() lexer.Token {
	if it.current == nil {
		panic("called current element before the iterator started")
	}
	return *it.current
}

func (it PeekableIterator) withIgnoreOffside(shouldIgnore bool) {
	it.ignoreOffside = shouldIgnore
}

func (it PeekableIterator) withOffside(col int) {
	it.offside = col
}
