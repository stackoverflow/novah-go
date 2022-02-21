package data

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestNoCycles(t *testing.T) {
	dag := NewDag[int, int](6)

	n1 := NewDagNode(1, 0)
	n2 := NewDagNode(2, 0)
	n3 := NewDagNode(3, 0)
	n4 := NewDagNode(4, 0)
	n5 := NewDagNode(5, 0)
	n6 := NewDagNode(6, 0)

	n1.Link(n2)
	n1.Link(n3)
	n2.Link(n3)
	n4.Link(n1)
	n4.Link(n5)
	n5.Link(n6)

	dag.AddNodes(n1, n2, n3, n4, n5, n6)

	cycle, found := dag.FindCycle()
	if found {
		t.Logf("%v", cycle)
		t.Error("Found cycle in graph ")
	}
}

func TestOneCycle(t *testing.T) {
	dag := NewDag[int, int](6)

	n1 := NewDagNode(1, 0)
	n2 := NewDagNode(2, 0)
	n3 := NewDagNode(3, 0)
	n4 := NewDagNode(4, 0)
	n5 := NewDagNode(5, 0)
	n6 := NewDagNode(6, 0)

	n1.Link(n2)
	n1.Link(n3)
	n2.Link(n3)
	n4.Link(n1)
	n4.Link(n5)
	n5.Link(n6)
	n6.Link(n4)

	dag.AddNodes(n1, n2, n3, n4, n5, n6)

	cycle, found := dag.FindCycle()
	if !found {
		t.Error("Found no cycle in graph")
	}

	cycleVals := MapSlice(cycle, func(t DagNode[int, int]) int { return t.Val })
	if !slices.Equal(cycleVals, []int{6, 5, 4}) {
		t.Error("Expected cycle to be [6, 5, 4]")
	}
}

func TestAnotherCycle(t *testing.T) {
	dag := NewDag[int, int](6)

	n1 := NewDagNode(1, 0)
	n2 := NewDagNode(2, 0)
	n4 := NewDagNode(4, 0)
	n5 := NewDagNode(5, 0)
	n6 := NewDagNode(6, 0)

	n1.Link(n2)
	n2.Link(n6)
	n4.Link(n1)
	n4.Link(n5)
	n6.Link(n4)

	dag.AddNodes(n1, n2, n4, n5, n6)

	cycle, found := dag.FindCycle()
	if !found {
		t.Error("Found no cycle in graph")
	}

	cycleVals := NewSet(MapSlice(cycle, func(t DagNode[int, int]) int { return t.Val })...)
	if !cycleVals.Equals(NewSet(4, 6, 2, 1)) {
		t.Error("Expected cycle to contain 4, 6, 2 and 1")
	}
}

func TestToposort(t *testing.T) {
	dag := NewDag[int, int](6)

	n1 := NewDagNode(1, 0)
	n2 := NewDagNode(2, 0)
	n3 := NewDagNode(3, 0)
	n4 := NewDagNode(4, 0)
	n5 := NewDagNode(5, 0)
	n6 := NewDagNode(6, 0)

	n1.Link(n2)
	n1.Link(n3)
	n2.Link(n3)
	n4.Link(n1)
	n4.Link(n5)
	n5.Link(n6)

	dag.AddNodes(n1, n2, n3, n4, n5, n6)

	sortedStack := dag.Toposort()
	sorted := make([]int, 0, sortedStack.Size())
	var v DagNode[int, int]
	for !sortedStack.IsEmpty() {
		v = sortedStack.Pop()
		sorted = append(sorted, v.Val)
	}
	expected := []int{4, 1, 2, 3, 5, 6}
	if !slices.Equal(sorted, expected) {
		t.Error("Sorted graph should be {4, 1, 2, 3, 5, 6}")
	}
}
