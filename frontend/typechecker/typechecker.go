package typechecker

type Typechecker struct {
	TypeVarMap map[int]string
	currentId  int
}

func NewTypechecker() *Typechecker {
	return &Typechecker{TypeVarMap: make(map[int]string)}
}

func (tc *Typechecker) NewVar(level int) Type {
	tc.currentId++
	return TVar{Tvar: TypeVar{Tag: UNBOUND, Id: tc.currentId, Level: level}}
}

func (tc *Typechecker) NewGenVarName(name string) TVar {
	tc.currentId++
	id := tc.currentId
	tc.TypeVarMap[id] = name
	return TVar{Tvar: TypeVar{Tag: GENERIC, Id: id}}
}
