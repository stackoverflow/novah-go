package parser

import (
	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
	"golang.org/x/exp/slices"
)

type Fixity = int

const (
	LEFT Fixity = iota
	RIGHT
)

// Parse a list of expressions and resolve operator precedence
// as well as left/right fixity
func parseApplication(exps []ast.SExpr) ast.SExpr {
	switch len(exps) {
	case 0:
		return nil
	case 1:
		return exps[0]
	}

	// first resolve function application as it has the highest precedence
	resExps := resolveApps(exps)
	if !validateOps(resExps) {
		return nil
	}

	res := resExps
	highest, found := getHighestPrecedence(res)
	for found {
		switch getFixity(highest) {
		case LEFT:
			res = resolveOp(res, func(exps []ast.SExpr) int { return slices.IndexFunc(exps, findOp(highest)) })
		case RIGHT:
			res = resolveOp(res, func(exps []ast.SExpr) int { return data.SliceLastIndexOfFunc(exps, findOp(highest)) })
		}
		highest, found = getHighestPrecedence(res)
	}

	return res[0]
}

// Validates that a list of [Expr]s is well formed.
// This function expects the list to be already resolved
// of function applications and only operators are left.
//
// Ex:
//
// a + 5 * 9 -> good
//
// a + * 9 -> bad
//
// a b * 6 -> bad
func validateOps(exps []ast.SExpr) bool {
	if len(exps)%2 == 0 {
		return false
	}

	var prev ast.SExpr
	for _, e := range exps {
		if prev == nil || isOp(prev) {
			if isOp(e) {
				return false
			}
		} else {
			if !isOp(e) {
				return false
			}
		}
		prev = e
	}
	return true
}

// Get the precedence of some operator
// which depends on the first symbol.
func getPrecedence(op ast.SOperator) int {
	switch op.Name[0] {
	case ';':
		return 0
	case '$':
		return 1
	case '|':
		return 2
	case '&':
		return 3
	case '=', '!':
		return 4
	case '<', '>':
		return 5
	case '+', '-', ':', '?':
		return 6
	case '*', '/', '%':
		return 7
	case '^', '.':
		return 8
	default:
		return 9
	}
}

// Returns the fixity (left/right) of an operator.
// Operators that start with `$` or `:` are right associative.
// Operators that end with `<` are right associative.
// <| is right associative.
// Everything else is left associative.
func getFixity(op ast.SOperator) Fixity {
	switch op.Name[0] {
	case '$', ':':
		return RIGHT
	default:
		{
			if op.Name == "<|" || op.Name[len(op.Name)-1] == '<' {
				return RIGHT
			} else {
				return LEFT
			}
		}
	}
}

// Resolve all the function applications in a list of expressions
func resolveApps(exps []ast.SExpr) []ast.SExpr {
	acc := make([]ast.SExpr, 0, len(exps))

	for _, e := range exps {
		prev := lastOrNil(acc)
		if !isOp(e) && prev != nil && !isOp(prev) {
			acc = acc[:len(acc)-1]
			acc = append(acc, ast.SApp{Fn: prev, Arg: e, Span: span(prev.GetSpan(), e.GetSpan())})
		} else {
			acc = append(acc, e)
		}
	}

	return acc
}

// Resolve some operator from left to right or right to left
// depending on the index function.
func resolveOp(exps []ast.SExpr, index func([]ast.SExpr) int) []ast.SExpr {
	input := make([]ast.SExpr, len(exps))
	copy(input, exps)

	i := index(input)
	for i != -1 {
		left := input[i-1]
		op := input[i]
		right := input[i+1]
		var app ast.SExpr = ast.SBinApp{Op: op, Left: left, Right: right, Span: span(left.GetSpan(), right.GetSpan()), Comment: left.GetComment()}

		input = data.DeleteAtSlice(input, i-1, nil)
		input = data.DeleteAtSlice(input, i-1, nil)
		input = data.DeleteAtSlice(input, i-1, nil)
		input = data.InsertAtSlice(input, i-1, app, nil)
		i = index(input)
	}

	return input
}

// Get the operator with the highest precedence in a list of expressions.
func getHighestPrecedence(exps []ast.SExpr) (ast.SOperator, bool) {
	ops := data.FilterSlice(exps, isOp)
	max := -1
	var op ast.SOperator
	for _, o := range ops {
		oo := o.(ast.SOperator)
		prec := getPrecedence(oo)
		if prec > max {
			max = prec
			op = oo
		}
	}
	return op, max != -1
}

func findOp(op ast.SOperator) func(ast.SExpr) bool {
	return func(exp ast.SExpr) bool {
		switch e := exp.(type) {
		case ast.SOperator:
			return op == e
		default:
			return false
		}
	}
}

func lastOrNil(exps []ast.SExpr) ast.SExpr {
	var exp ast.SExpr
	if len(exps) > 0 {
		exp = exps[len(exps)-1]
	}
	return exp
}

func isOp(exp ast.SExpr) bool {
	switch exp.(type) {
	case ast.SOperator:
		return true
	default:
		return false
	}
}
