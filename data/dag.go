package data

// A Direct Acyclic Graph.
// It doesn't actually check for cycles while adding nodes and links.
type Dag[T comparable, D any] struct {
	nodes []*DagNode[T, D]
}

func NewDag[T comparable, D any](expectedSize int) *Dag[T, D] {
	return &Dag[T, D]{nodes: make([]*DagNode[T, D], 0, expectedSize)}
}

func (d *Dag[T, D]) AddNodes(ns ...*DagNode[T, D]) {
	d.nodes = append(d.nodes, ns...)
}

func (d *Dag[T, D]) Size() int {
	return len(d.nodes)
}

// Finds the first cycle in the DAG.
// The second return tells if there's a cycle
func (d *Dag[T, D]) FindCycle() ([]DagNode[T, D], bool) {
	whiteSet := make(map[T]*DagNode[T, D])
	graySet := make(map[T]DagNode[T, D])
	blackSet := make(map[T]DagNode[T, D])
	parentage := make(map[T]*DagNode[T, D])

	for _, node := range d.nodes {
		whiteSet[node.Val] = node
	}

	// depth first search
	var dfs func(DagNode[T, D], *DagNode[T, D]) *DagNode[T, D]
	dfs = func(current DagNode[T, D], parent *DagNode[T, D]) *DagNode[T, D] {
		delete(whiteSet, current.Val)
		graySet[current.Val] = current
		parentage[current.Val] = parent

		for _, neighbor := range current.Neighbors {
			if _, has := blackSet[neighbor.Val]; has {
				continue
			}

			// found cycle
			if _, has := graySet[neighbor.Val]; has {
				return &current
			}
			res := dfs(*neighbor, &current)
			if res != nil {
				return res
			}
		}

		delete(graySet, current.Val)
		blackSet[current.Val] = current
		return nil
	}

	for len(whiteSet) > 0 {
		current := FirstInMap(whiteSet)
		cycled := dfs(*current, nil)
		if cycled != nil {
			return d.reportCycle(*cycled, parentage), true
		}
	}
	return []DagNode[T, D]{}, false
}

// Returns a topological sorted representation of this graph.
func (d *Dag[T, D]) Toposort() *Stack[DagNode[T, D]] {
	visited := NewSet[T]()
	stack := NewStack[DagNode[T, D]]()

	var helper func(*DagNode[T, D])
	helper = func(node *DagNode[T, D]) {
		visited.Add(node.Val)

		for _, neighbor := range node.Neighbors {
			if visited.Contains(neighbor.Val) {
				continue
			}
			helper(neighbor)
		}
		stack.Push(*node)
	}

	for _, node := range ReverseSlice(d.nodes) {
		if visited.Contains(node.Val) {
			continue
		}
		helper(node)
	}
	return stack
}

func (d *Dag[T, D]) reportCycle(node DagNode[T, D], parentage map[T]*DagNode[T, D]) []DagNode[T, D] {
	cycle := []DagNode[T, D]{node}

	parent := parentage[node.Val]
	for parent != nil {
		cycle = append(cycle, *parent)
		parent = parentage[parent.Val]
	}
	return cycle
}

// A node in the DAG
// `value` has to be unique for every node.
type DagNode[T comparable, D any] struct {
	Val       T
	Data      D
	Neighbors []*DagNode[T, D]
}

func NewDagNode[T comparable, D any](val T, data D) *DagNode[T, D] {
	return &DagNode[T, D]{Val: val, Data: data}
}

func (n *DagNode[T, R]) Link(other *DagNode[T, R]) {
	n.Neighbors = append(n.Neighbors, other)
}

func FirstInMap[K comparable, V any](m map[K]V) V {
	for _, v := range m {
		return v
	}
	panic("map empty")
}
