package typechecker

import (
	"testing"

	"github.com/huandu/go-clone"
)

func TestClone(t *testing.T) {
	ty := TApp{Type: TConst{Name: "bla"}, Types: []Type{TVar{Tvar: TypeVar{Id: 3, Level: 5}}}}
	ty2 := ty.Clone()
	ty.Type = TConst{Name: "err"}
	ty.Types[0] = TConst{Name: "oops"}

	m := make(map[string]int)
	m["a"] = 3
	m["f"] = 9
	m2 := clone.Clone(m).(map[string]int)
	m2["z"] = 2
	m2["r"] = 0

	t.Log(m)
	t.Log(m2)

	t.Logf("%v", ty)
	t.Logf("%v", ty2)
}
