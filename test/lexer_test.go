package test

import (
	"os"
	"testing"

	"github.com/stackoverflow/novah-go/frontend/lexer"
)

func TestSpans(t *testing.T) {

	wd, _ := os.Getwd()
	t.Log(wd)

	//all, _ := os.ReadFile("data/span.novah")
	//t.Log(string(all))

	tks := LexString("data/span.novah")
	if len(tks) != 5 {
		t.Error("total tokens should be 3 but was", len(tks))
	}

	if tks[0].Span != lexer.NewSpan2(1, 1, 1, 5) {
		t.Error("Expected first token to be at 1:1, 1:5")
	}
}
