package ast

import (
	"fmt"
	"strings"

	"github.com/huandu/go-clone"
	"github.com/stackoverflow/novah-go/data"
)

type KindType = int

const (
	STAR KindType = iota
	CTOR
)

type Kind struct {
	Type  KindType
	Arity int
}

func (k Kind) String() string {
	if k.Type == STAR {
		return "Type"
	}
	return data.JoinToStringFunc(data.Range(0, k.Arity), " -> ", func(_ int) string { return "Type" })
}

type TypeVarTag = int

const (
	UNBOUND TypeVarTag = iota
	LINK
	GENERIC
)

type Id = int
type Level = int

type TypeVar struct {
	Tag   TypeVarTag
	Id    Id
	Level Level
	Type  Type
}

type Type interface {
	sType()
	Clone() Type
	GetSpan() data.Span
	WithSpan(data.Span) Type
	GetKind() Kind
	Equals(Type) bool
	fmt.Stringer
}

type TConst struct {
	Name string
	Kind Kind
	Span data.Span
}

type TApp struct {
	Type  Type
	Types []Type
	Span  data.Span
}

type TArrow struct {
	Args []Type
	Ret  Type
	Span data.Span
}

type TImplicit struct {
	Type Type
	Span data.Span
}

type TRecord struct {
	Row  Type
	Span data.Span
}

type TRowEmpty struct {
	Span data.Span
}

type TRowExtend struct {
	Labels data.LabelMap[Type]
	Row    Type
	Span   data.Span
}

type TVar struct {
	Tvar *TypeVar
	Span data.Span
}

func (_ TConst) sType()     {}
func (_ TApp) sType()       {}
func (_ TArrow) sType()     {}
func (_ TImplicit) sType()  {}
func (_ TRecord) sType()    {}
func (_ TRowEmpty) sType()  {}
func (_ TRowExtend) sType() {}
func (_ TVar) sType()       {}
func (t TConst) GetSpan() data.Span {
	return t.Span
}
func (t TApp) GetSpan() data.Span {
	return t.Span
}
func (t TArrow) GetSpan() data.Span {
	return t.Span
}
func (t TImplicit) GetSpan() data.Span {
	return t.Span
}
func (t TRecord) GetSpan() data.Span {
	return t.Span
}
func (t TRowEmpty) GetSpan() data.Span {
	return t.Span
}
func (t TRowExtend) GetSpan() data.Span {
	return t.Span
}
func (t TVar) GetSpan() data.Span {
	return t.Span
}
func (t TConst) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}
func (t TApp) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}
func (t TArrow) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}
func (t TImplicit) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}
func (t TRecord) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}
func (t TRowEmpty) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}
func (t TRowExtend) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}
func (t TVar) WithSpan(span data.Span) Type {
	t.Span = span
	return t
}

func (t TConst) Clone() Type {
	return clone.Clone(t).(TConst)
}
func (t TApp) Clone() Type {
	return clone.Clone(t).(TApp)
}
func (t TArrow) Clone() Type {
	return clone.Clone(t).(TArrow)
}
func (t TImplicit) Clone() Type {
	return clone.Clone(t).(TImplicit)
}
func (t TRecord) Clone() Type {
	return clone.Clone(t).(TRecord)
}
func (t TRowEmpty) Clone() Type {
	return clone.Clone(t).(TRowEmpty)
}
func (t TRowExtend) Clone() Type {
	return clone.Clone(t).(TRowExtend)
}
func (t TVar) Clone() Type {
	return clone.Clone(t).(TVar)
}

func (t TConst) String() string {
	return ShowType(t)
}
func (t TApp) String() string {
	return ShowType(t)
}
func (t TArrow) String() string {
	return ShowType(t)
}
func (t TImplicit) String() string {
	return ShowType(t)
}
func (t TRecord) String() string {
	return ShowType(t)
}
func (t TRowEmpty) String() string {
	return ShowType(t)
}
func (t TRowExtend) String() string {
	return ShowType(t)
}
func (t TVar) String() string {
	return ShowType(t)
}

func (t TConst) GetKind() Kind {
	return t.Kind
}
func (t TArrow) GetKind() Kind {
	return Kind{Type: CTOR, Arity: 1}
}
func (t TApp) GetKind() Kind {
	return t.Type.GetKind()
}
func (t TVar) GetKind() Kind {
	tv := t.Tvar
	if tv.Tag == LINK {
		return tv.Type.GetKind()
	}
	return Kind{Type: STAR}
}

// this is not right, but I'll postpone adding real row kinds for now
func (t TRowEmpty) GetKind() Kind {
	return Kind{Type: STAR}
}
func (t TRecord) GetKind() Kind {
	return t.Row.GetKind()
}
func (t TRowExtend) GetKind() Kind {
	return t.Row.GetKind()
}
func (t TImplicit) GetKind() Kind {
	return t.Type.GetKind()
}

func (t TConst) Equals(other Type) bool {
	tc, isTc := other.(TConst)
	if isTc {
		return t.Name == tc.Name
	}
	return false
}
func (t TApp) Equals(other Type) bool {
	tapp, isTapp := other.(TApp)
	if !isTapp {
		return false
	}
	if !t.Type.Equals(tapp.Type) || len(t.Types) != len(tapp.Types) {
		return false
	}
	for i := 0; i < len(t.Types); i++ {
		if !t.Types[i].Equals(tapp.Types[i]) {
			return false
		}
	}
	return true
}
func (t TArrow) Equals(other Type) bool {
	tarr, isTaarr := other.(TArrow)
	if !isTaarr {
		return false
	}
	if !t.Ret.Equals(tarr.Ret) || len(t.Args) != len(tarr.Args) {
		return false
	}
	for i := 0; i < len(t.Args); i++ {
		if !t.Args[i].Equals(tarr.Args[i]) {
			return false
		}
	}
	return true
}
func (t TVar) Equals(other Type) bool {
	tv, isTvar := other.(TVar)
	if !isTvar {
		return false
	}
	tvar := t.Tvar
	if tvar.Tag == LINK {
		return tv.Tvar.Tag == LINK && tvar.Type.Equals(tv.Tvar.Type)
	}
	if tvar.Tag == UNBOUND {
		return tv.Tvar.Tag == UNBOUND && tvar.Id == tvar.Id && tvar.Level == tv.Tvar.Level
	}
	return tv.Tvar.Tag == GENERIC && tvar.Id == tv.Tvar.Id
}
func (t TRowEmpty) Equals(other Type) bool {
	_, isTre := other.(TRowEmpty)
	return isTre
}
func (t TRecord) Equals(other Type) bool {
	rec, isTre := other.(TRecord)
	if !isTre {
		return false
	}
	return t.Row.Equals(rec.Row)
}
func (t TRowExtend) Equals(other Type) bool {
	rec, isTre := other.(TRowExtend)
	if !isTre {
		return false
	}
	if !t.Row.Equals(rec.Row) {
		return false
	}
	ents1, ents2 := t.Labels.Entries(), rec.Labels.Entries()
	if len(ents1) != len(ents2) {
		return false
	}
	for i := 0; i < len(ents1); i++ {
		e1, e2 := ents1[i], ents2[i]
		if e1.Label != e2.Label || !e1.Val.Equals(e2.Val) {
			return false
		}
	}
	return true
}
func (t TImplicit) Equals(other Type) bool {
	tim, isTim := other.(TImplicit)
	if !isTim {
		return false
	}
	return t.Type.Equals(tim.Type)
}

func RealType(ty Type) Type {
	if tv, isTvar := ty.(TVar); isTvar && tv.Tvar.Tag == LINK {
		return RealType(tv.Tvar.Type)
	}
	return ty
}

// pretty print this type
func ShowType(typ Type) string {
	return ShowTypeInner(typ, true, make(map[int]string))
}

func ShowTypeInner(typ Type, qualified bool, tvarsMap map[int]string) string {
	showId := func(id Id) string {
		if id, has := tvarsMap[id]; has {
			return id
		}
		if id >= 0 {
			return fmt.Sprintf("t%d", id)
		} else {
			return fmt.Sprintf("u%d", -id)
		}
	}

	var run func(Type, bool, bool) string
	run = func(ty Type, nested bool, topLevel bool) string {
		switch t := ty.(type) {
		case TConst:
			if qualified {
				return t.Name
			} else {
				name := strings.Split(t.Name, ".")
				return name[len(name)-1]
			}
		case TApp:
			{
				sname := run(t.Type, nested, false)
				if len(t.Types) == 0 {
					return sname
				}
				sname = fmt.Sprintf("%s %s", sname, data.JoinToStringFunc(t.Types, " ", func(t Type) string { return run(t, true, false) }))
				if nested {
					return fmt.Sprintf("(%s)", sname)
				}
				return sname
			}
		case TArrow:
			{
				arg := t.Args[0]
				_, isArr := arg.(TArrow)
				args := run(arg, isArr, false)
				if nested {
					return fmt.Sprintf("(%s -> %s)", args, run(t.Ret, false, false))
				}
				return fmt.Sprintf("%s -> %s", args, run(t.Ret, nested, false))
			}
		case TVar:
			{
				tv := t.Tvar
				if tv.Tag == LINK {
					return run(tv.Type, nested, topLevel)
				}
				return showId(tv.Id)
			}
		case TRowEmpty:
			return "[]"
		case TRecord:
			{
				switch r := RealType(t.Row).(type) {
				case TRowEmpty:
					return "{}"
				case TRowExtend:
					{
						rows := run(r, false, true)
						return fmt.Sprintf("{%s}", rows[1:len(rows)-1])
					}
				default:
					return fmt.Sprintf("{ | %s }", run(r, false, true))
				}
			}
		case TRowExtend:
			{
				labels := t.Labels.Show(func(k string, v Type) string { return fmt.Sprintf("%s : %s", k, run(v, false, true)) })
				var str string
				switch r := RealType(t.Row).(type) {
				case TRowEmpty:
					str = labels
				case TRowExtend:
					{
						rows := run(r, false, true)
						if labels == "" {
							str = rows[2 : len(rows)-2]
						} else {
							str = fmt.Sprintf("%s, %s", labels, rows[2:len(rows)-2])
						}
					}
				default:
					if labels == "" {
						str = fmt.Sprintf("| %s", run(r, false, true))
					} else {
						str = fmt.Sprintf("%s | %s", labels, run(r, false, true))
					}
				}
				return fmt.Sprintf("[ %s ]", str)
			}
		case TImplicit:
			return fmt.Sprintf("{{ %s }}", run(t.Type, false, false))
		default:
			panic("got unknow type in ShowType")
		}
	}
	return run(typ, false, true)
}

func SubstConst(typ Type, m map[string]Type) Type {
	switch t := typ.(type) {
	case TConst:
		if ty, has := m[t.Name]; has {
			return ty
		} else {
			return typ
		}
	case TApp:
		return TApp{
			Type:  SubstConst(t.Type, m),
			Types: data.MapSlice(t.Types, func(t Type) Type { return SubstConst(t, m) }),
			Span:  t.Span,
		}
	case TArrow:
		return TArrow{
			Args: data.MapSlice(t.Args, func(t Type) Type { return SubstConst(t, m) }),
			Ret:  SubstConst(t.Ret, m),
			Span: t.Span,
		}
	case TVar:
		{
			if t.Tvar.Tag == LINK {
				return TVar{Tvar: &TypeVar{Tag: LINK, Type: SubstConst(t.Tvar.Type, m)}, Span: t.Span}
			}
			return typ
		}
	case TRowEmpty:
		return typ
	case TRecord:
		return TRecord{Row: SubstConst(t.Row, m), Span: t.Span}
	case TRowExtend:
		return TRowExtend{
			Labels: data.LabelMapValues(t.Labels, func(t Type) Type { return SubstConst(t, m) }),
			Row:    SubstConst(t.Row, m),
			Span:   t.Span,
		}
	case TImplicit:
		return TImplicit{Type: SubstConst(t.Type, m), Span: t.Span}
	default:
		panic("Got unknow type in SubstConst")
	}
}

// Recursively walks this type up->bottom
func EverywhereTypeUnit(this Type, f func(Type)) {
	var run func(Type)
	run = func(typ Type) {
		f(typ)
		switch t := typ.(type) {
		case TApp:
			{
				run(t.Type)
				for _, ty := range t.Types {
					run(ty)
				}
			}
		case TArrow:
			{
				for _, ty := range t.Args {
					run(ty)
				}
				run(t.Ret)
			}
		case TVar:
			if t.Tvar.Tag == LINK {
				run(t.Tvar.Type)
			}
		case TRecord:
			run(t.Row)
		case TRowExtend:
			{
				run(t.Row)
				for _, ty := range t.Labels.Values() {
					run(ty)
				}
			}
		case TImplicit:
			run(t.Type)
		}
	}
	run(this)
}

func NestArrows(args []Type, ret Type) Type {
	if len(args) == 0 {
		return ret
	}
	return TArrow{Args: []Type{args[0]}, Ret: NestArrows(args[1:], ret)}
}
