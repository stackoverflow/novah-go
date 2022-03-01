package typechecker

import (
	"fmt"

	"github.com/stackoverflow/novah-go/data"
	"github.com/stackoverflow/novah-go/frontend/ast"
	"golang.org/x/exp/slices"
)

type Unification struct {
	tc *Typechecker
}

func NewUnification(tc *Typechecker) *Unification {
	return &Unification{tc: tc}
}

func (u *Unification) Unify(t1, t2 ast.Type, span data.Span) *data.CompilerProblem {
	err := u.unify(t1, t2, span)
	if err == nil {
		return nil
	}
	var reason string
	switch e := err.(type) {
	case noMatch:
		if ast.RealType(t1).Equals(ast.RealType(e.t1)) && ast.RealType(t2).Equals(ast.RealType(e.t2)) {
			reason = ""
		} else {
			reason = data.IncompatibleTypes(e.t1.String(), e.t2.String())
		}
	case missingLabels:
		reason = data.RecordMissingLabels(e.labels.Show(func(s string, t ast.Type) string { return fmt.Sprintf("%s : %s", s, t.String()) }))
	case infiniteType:
		reason = data.InfiniteType(e.ty.String())
	case notRow:
		reason = data.NotARow(e.ty.String())
	case recursiveRows:
		reason = data.RECURSIVE_ROWS
	}
	return u.tc.makeErrorRef(data.TypesDontMatch(t1.String(), t2.String(), reason), span)
}

func (u *Unification) unify(t1, t2 ast.Type, span data.Span) error {
	var err error
	tc1, isTconst1 := t1.(ast.TConst)
	tc2, isTconst2 := t2.(ast.TConst)
	if isTconst1 && isTconst2 && tc1.Name == tc2.Name {
		return nil
	}

	ta1, isTapp1 := t1.(ast.TApp)
	ta2, isTapp2 := t2.(ast.TApp)
	if isTapp1 && isTapp2 {
		err = u.unify(ta1.Type, ta2.Type, span)
		if err != nil {
			return err
		}
		if len(ta1.Types) != len(ta2.Types) {
			return noMatch{t1, t2}
		}
		for i := 0; i < len(ta1.Types); i++ {
			err = u.unify(ta1.Types[i], ta2.Types[i], span)
			if err != nil {
				return err
			}
		}
		return nil
	}

	tf1, isTarr1 := t1.(ast.TArrow)
	tf2, isTarr2 := t2.(ast.TArrow)
	if isTarr1 && isTarr2 {
		if len(tf1.Args) != len(tf2.Args) {
			return noMatch{t1, t2}
		}
		for i := 0; i < len(tf1.Args); i++ {
			err = u.unify(tf1.Args[i], tf2.Args[i], span)
			if err != nil {
				return err
			}
		}
		err = u.unify(tf1.Ret, tf2.Ret, span)
		if err != nil {
			return err
		}
		return nil
	}

	tv1, isTvar1 := t1.(ast.TVar)
	if isTvar1 && tv1.Tvar.Tag == ast.LINK {
		return u.unify(tv1.Tvar.Type, t2, span)
	}
	tv2, isTvar2 := t2.(ast.TVar)
	if isTvar2 && tv2.Tvar.Tag == ast.LINK {
		return u.unify(t1, tv2.Tvar.Type, span)
	}
	if isTvar1 && isTvar2 && tv1.Tvar.Tag == ast.UNBOUND && tv2.Tvar.Tag == ast.UNBOUND && tv1.Tvar.Id == tv2.Tvar.Id {
		panic(fmt.Sprintf("error in unification: %s with %s", t1.String(), t2.String()))
	}
	if isTvar1 && tv1.Tvar.Tag == ast.UNBOUND {
		err = u.occursCheckAndAdjustLevels(tv1.Tvar.Id, tv1.Tvar.Level, t2)
		if err != nil {
			return err
		}
		tv1.Tvar.Tag = ast.LINK
		tv1.Tvar.Type = t2
		return nil
	}
	if isTvar2 && tv2.Tvar.Tag == ast.UNBOUND {
		err = u.occursCheckAndAdjustLevels(tv2.Tvar.Id, tv2.Tvar.Level, t1)
		if err != nil {
			return err
		}
		tv2.Tvar.Tag = ast.LINK
		tv2.Tvar.Type = t1
		return nil
	}

	_, isTRempty1 := t1.(ast.TRowEmpty)
	_, isTRempty2 := t2.(ast.TRowEmpty)
	if isTRempty1 && isTRempty2 {
		return nil
	}

	tr1, isTRec1 := t1.(ast.TRecord)
	tr2, isTRec2 := t2.(ast.TRecord)
	if isTRec1 && isTRec2 {
		return u.unify(tr1.Row, tr2.Row, span)
	}

	_, isTRex1 := t1.(ast.TRowExtend)
	_, isTRex2 := t2.(ast.TRowExtend)
	if isTRex1 && isTRex2 {
		return u.unifyRows(t1, t2, span)
	}

	if isTRempty1 && isTRex2 {
		labels, _, err := u.matchRowType(t2)
		if err != nil {
			return err
		}
		return missingLabels{labels: labels}
	}
	if isTRex2 && isTRempty2 {
		labels, _, err := u.matchRowType(t1)
		if err != nil {
			return err
		}
		return missingLabels{labels: labels}
	}

	ti1, isTimp1 := t1.(ast.TImplicit)
	ti2, isTimp2 := t2.(ast.TImplicit)
	if isTimp1 && isTimp2 {
		return u.unify(ti1.Type, ti2.Type, span)
	}
	if isTimp1 {
		return u.unify(ti1.Type, t2, span)
	}
	if isTimp2 {
		return u.unify(t1, ti2.Type, span)
	}
	return noMatch{t1: t1, t2: t2}
}

type TypeLabel = data.LabelMap[ast.Type]

// unify two row types
func (u *Unification) unifyRows(row1, row2 ast.Type, span data.Span) error {
	labels1, restTy1, err := u.matchRowType(row1)
	if err != nil {
		return err
	}
	labels2, restTy2, err2 := u.matchRowType(row2)
	if err2 != nil {
		return err2
	}

	var unifyTypes func([]ast.Type, []ast.Type) ([]ast.Type, []ast.Type, error)
	unifyTypes = func(t1s, t2s []ast.Type) ([]ast.Type, []ast.Type, error) {
		if len(t1s) == 0 || len(t2s) == 0 {
			return t1s, t2s, nil
		}
		err := u.unify(t1s[0], t2s[0], span)
		if err != nil {
			return nil, nil, err
		}
		return unifyTypes(t1s[1:], t2s[1:])
	}

	var unifylabels func(TypeLabel, TypeLabel, []data.Entry[[]ast.Type], []data.Entry[[]ast.Type]) (TypeLabel, TypeLabel, error)
	unifylabels = func(missing1, missing2 TypeLabel, labels1, labels2 []data.Entry[[]ast.Type]) (TypeLabel, TypeLabel, error) {
		if len(labels1) == 0 && len(labels2) == 0 {
			return missing1, missing2, nil
		}
		if len(labels1) == 0 {
			return u.addDistinctLabels(missing1, labels2), missing2, nil
		}
		if len(labels2) == 0 {
			return missing1, u.addDistinctLabels(missing2, labels1), nil
		}
		label1, tys1, rest1 := headTail(labels1)
		label2, tys2, rest2 := headTail(labels2)
		if label1 == label2 {
			m1s, m2s, err := unifyTypes(tys1, tys2)
			if err != nil {
				return missing1, missing2, err
			}
			var missing11, missing22 TypeLabel
			if len(m1s) == 0 && len(m2s) == 0 {
				missing11, missing22 = missing1, missing2
			} else if len(m2s) == 0 {
				missing11, missing22 = missing1, missing2.Put(label1, m1s)
			} else if len(m1s) == 0 {
				missing11, missing22 = missing1.Put(label2, m2s), missing2
			} else {
				panic("unifyLabels: impossible")
			}
			return unifylabels(missing11, missing22, rest1, rest2)
		}
		if label1 < label2 {
			return unifylabels(missing1, missing2.Put(label1, tys1), rest1, labels2)
		}
		return unifylabels(missing1.Put(label2, tys2), missing2, labels1, rest2)
	}

	missing1, missing2, err := unifylabels(data.EmptyLabelMap[ast.Type](), data.EmptyLabelMap[ast.Type](), concatLabelMap(labels1), concatLabelMap(labels2))
	if err != nil {
		return err
	}
	empty1, empty2 := missing1.IsEmpty(), missing2.IsEmpty()
	if empty1 && empty2 {
		return u.unify(restTy1, restTy2, span)
	}
	if empty1 && !empty2 {
		return u.unify(restTy2, ast.TRowExtend{Labels: missing2, Row: restTy1}, span)
	}
	if !empty1 && empty2 {
		return u.unify(restTy1, ast.TRowExtend{Labels: missing1, Row: restTy2}, span)
	}
	// both not empty
	if _, isREmpty := restTy1.(ast.TRowEmpty); isREmpty {
		return u.unify(restTy1, ast.TRowExtend{Labels: missing1, Row: u.tc.NewVar(0)}, span)
	}
	if tv, isTvar := restTy1.(ast.TVar); isTvar && tv.Tvar.Tag == ast.UNBOUND {
		restRow := u.tc.NewVar(tv.Tvar.Level)
		err = u.unify(restTy2, ast.TRowExtend{Labels: missing2, Row: restRow}, span)
		if err != nil {
			return err
		}
		if tv.Tvar.Tag == ast.LINK {
			return recursiveRows{}
		}
		return u.unify(restTy1, ast.TRowExtend{Labels: missing1, Row: restRow}, span)
	}
	return noMatch{t1: row1, t2: row2}
}

func (u *Unification) occursCheckAndAdjustLevels(id ast.Id, level ast.Level, typ ast.Type) error {
	var run func(ast.Type) error
	run = func(ty ast.Type) error {
		switch t := ty.(type) {
		case ast.TVar:
			{
				tv := t.Tvar
				if tv.Tag == ast.LINK {
					run(tv.Type)
				} else if tv.Tag == ast.UNBOUND {
					if id == tv.Id {
						return infiniteType{ty: typ}
					} else if tv.Level > level {
						t.Tvar.Id = tv.Id
						t.Tvar.Level = level
					}
				}
			}
		case ast.TApp:
			{
				run(t.Type)
				for _, ty := range t.Types {
					run(ty)
				}
			}
		case ast.TArrow:
			{
				for _, ty := range t.Args {
					run(ty)
				}
				run(t.Ret)
			}
		case ast.TRecord:
			run(t.Row)
		case ast.TRowExtend:
			{
				for _, v := range t.Labels.Values() {
					run(v)
				}
				run(t.Row)
			}
		case ast.TImplicit:
			run(t.Type)
		}
		return nil
	}
	return run(typ)
}

func (u *Unification) MatchRowType(typ ast.Type, span data.Span) (data.LabelMap[ast.Type], ast.Type, *data.CompilerProblem) {
	labels, ty, err := u.matchRowType(typ)
	if err != nil {
		e := err.(notRow)
		return labels, ty, u.tc.makeErrorRef(data.NotARow(e.ty.String()), span)
	}
	return labels, ty, nil
}

func (u *Unification) matchRowType(typ ast.Type) (data.LabelMap[ast.Type], ast.Type, error) {
	switch t := typ.(type) {
	case ast.TRowEmpty:
		return data.EmptyLabelMap[ast.Type](), ast.TRowEmpty{Span: t.Span}, nil
	case ast.TVar:
		{
			tv := t.Tvar
			if tv.Tag == ast.LINK {
				return u.matchRowType(tv.Type)
			}
			return data.EmptyLabelMap[ast.Type](), typ, nil
		}
	case ast.TRowExtend:
		{
			restLabels, restTy, err := u.matchRowType(t.Row)
			if err != nil {
				return restLabels, restTy, err
			}
			if restLabels.IsEmpty() {
				return t.Labels, restTy, nil
			}
			return t.Labels.Merge(restLabels), restTy, nil
		}
	default:
		return data.EmptyLabelMap[ast.Type](), typ, notRow{ty: typ}
	}
}

func (u *Unification) addDistinctLabels(lm TypeLabel, other []data.Entry[[]ast.Type]) data.LabelMap[ast.Type] {
	ents := lm.Entries()
	m := make([]data.Entry[ast.Type], 0, len(ents)+len(other))
	keys := data.NewSet[string]()
	for _, e := range ents {
		keys.Add(e.Label)
		m = append(m, e)
	}
	for _, es := range other {
		if keys.Contains(es.Label) {
			panic("Label map already contains label " + es.Label)
		}
		for _, ty := range es.Val {
			m = append(m, data.Entry[ast.Type]{Label: es.Label, Val: ty})
		}
	}
	return data.LabelMapFrom(m...)
}

// Returns the first entry and the rest.
// Panics if it's empty
func headTail[T any](labels []data.Entry[[]T]) (string, []T, []data.Entry[[]T]) {
	head := labels[0]
	return head.Label, head.Val, labels[1:]
}

// Concatenate and sort all values for a label
func concatLabelMap[T any](lm data.LabelMap[T]) []data.Entry[[]T] {
	res := make([]data.Entry[[]T], 0, lm.Size())
	entries := lm.Copy().Entries()
	slices.SortStableFunc(entries, func(a, b data.Entry[T]) bool { return a.Label < b.Label })
	tmp := entries
	key := tmp[0].Label
	vals := make([]T, 0, 1)
	for len(tmp) > 0 {
		ent := tmp[0]
		if key == ent.Label {
			vals = append(vals, ent.Val)
		} else {
			res = append(res, data.Entry[[]T]{Label: key, Val: vals})
			vals = make([]T, 0, 1)
			vals = append(vals, ent.Val)
		}
		key = ent.Label
		tmp = tmp[1:]
	}
	res = append(res, data.Entry[[]T]{Label: key, Val: vals})
	return res
}

// inner errors

type noMatch struct {
	t1 ast.Type
	t2 ast.Type
}

type missingLabels struct {
	labels TypeLabel
}

type infiniteType struct {
	ty ast.Type
}

type notRow struct {
	ty ast.Type
}

type recursiveRows struct{}

func (e noMatch) Error() string {
	return ""
}

func (e missingLabels) Error() string {
	return ""
}

func (e infiniteType) Error() string {
	return ""
}

func (e notRow) Error() string {
	return ""
}

func (e recursiveRows) Error() string {
	return data.RECURSIVE_ROWS
}
