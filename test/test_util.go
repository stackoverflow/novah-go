package test

import (
	"os"

	"github.com/stackoverflow/novah-go/frontend/lexer"
)

func LexString(input string) []lexer.Token {
	var tokens []lexer.Token

	reader, _ := os.Open(input)
	defer reader.Close()

	lexer := lexer.New("span.novah", reader)
	for {
		tk := lexer.Scan()
		tokens = append(tokens, tk)
		if tk.IsEOF() {
			break
		}
	}
	return tokens
}
