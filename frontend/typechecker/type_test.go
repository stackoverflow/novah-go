package typechecker

import "testing"

func TestClone(t *testing.T) {
	ty := TApp{Type: TConst{Name: "bla"}, Types: []Type{TVar{Tvar: TypeVar{Id: 3, Level: 5}}}}
	ty2 := ty.Clone()
	ty.Type = TConst{Name: "err"}
	ty.Types[0] = TConst{Name: "oops"}

	t.Logf("%v", ty)
	t.Logf("%v", ty2)
}
