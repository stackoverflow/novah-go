package typechecker

import (
	"testing"

	"github.com/stackoverflow/novah-go/data"
	"golang.org/x/exp/slices"
)

func TestConcatLabels(t *testing.T) {
	lm := data.LabelMapFrom([]data.Entry[int]{{Label: "c", Val: 1}, {Label: "b", Val: 2}, {Label: "c", Val: 3}, {Label: "a", Val: 5}, {Label: "a", Val: 4}}...)
	res := concatLabelMap(lm)

	if len(res) != 3 {
		t.Error("expected 3 labels")
	}

	a := res[0]
	if a.Label != "a" {
		t.Error("first label should be `a`")
	}
	if !slices.Equal(a.Val, []int{5, 4}) {
		t.Error("a values should be 5, 4")
	}

	b := res[1]
	if b.Label != "b" {
		t.Error("second label should be `b`")
	}
	if !slices.Equal(b.Val, []int{2}) {
		t.Error("b values should be 2")
	}

	c := res[2]
	if c.Label != "c" {
		t.Error("third label should be `c`")
	}
	if !slices.Equal(c.Val, []int{1, 3}) {
		t.Error("c values should be 1, 3")
	}
}
