package lexer

import (
	"os"
	"strings"
	"testing"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/test"
)

func TestSpans(t *testing.T) {
	tks := lexResource("../../test_data/span.novah")

	test.Equals(t, tks[0].Span, data.NewSpan2(1, 1, 1, 5))
	test.Equals(t, tks[1].Span, data.NewSpan2(1, 6, 1, 19))
	test.Equals(t, tks[2].Span, data.NewSpan2(1, 20, 1, 21))
	test.Equals(t, tks[3].Span, data.NewSpan2(3, 1, 3, 21))
	test.Equals(t, tks[4].Type, EOF)
}

func TestComments(t *testing.T) {
	tks := lexResource("../../test_data/comments.novah")

	typ, _ := data.FindSlice(tks, func(tk Token) bool { return tk.Type == TYPE })
	myFun, _ := data.FindSlice(tks, func(tk Token) bool { return tk.Type == IDENT && *tk.Text == "myFun" })
	foo, _ := data.FindSlice(tks, func(tk Token) bool { return tk.Type == IDENT && *tk.Text == "foo" })
	other, _ := data.FindSlice(tks, func(tk Token) bool { return tk.Type == IDENT && *tk.Text == "other" })

	test.Equals(t, typ.Span, data.NewSpan2(4, 1, 4, 5))
	test.Equals(t, Comment{Text: " comments on type definitions work", Span: data.NewSpan2(3, 1, 3, 37)}, *typ.Comment)

	test.Equals(t, myFun.Span, data.NewSpan2(10, 1, 10, 6))
	test.Equals(t, *myFun.Comment, Comment{Text: "\n comments on var\n types work\n", Span: data.NewSpan2(6, 1, 9, 3), IsMulti: true})

	test.Equals(t, *other.Comment, Comment{Text: " comments on var declaration work\n and are concatenated", Span: data.NewSpan2(13, 1, 14, 24)})

	test.Equals(t, foo.Comment, nil)
}

func TestUTFEscapes(t *testing.T) {
	tk := lexString(` "bla bla \u0062 a" `)[0]

	v, isStr := tk.Value.(string)
	test.Equals(t, isStr, true)
	test.Equals(t, v, "bla bla b a")
	test.Equals(t, *tk.Text, "bla bla \\u0062 a")
}

func lexResource(input string) []Token {
	tokens := make([]Token, 0, 10)

	reader, _ := os.Open(input)
	defer reader.Close()

	lexer := New(input, reader)
	for {
		tk := lexer.Scan()
		tokens = append(tokens, tk)
		if tk.IsEOF() {
			break
		}
	}
	return tokens
}

func lexString(str string) []Token {
	tokens := make([]Token, 0, 10)

	reader := strings.NewReader(str)

	lexer := New("string", reader)
	for {
		tk := lexer.Scan()
		tokens = append(tokens, tk)
		if tk.IsEOF() {
			break
		}
	}
	return tokens
}
