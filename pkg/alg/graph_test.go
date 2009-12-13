package alg

import (
	"testing"
)

type Typ struct {
	x int
}

func TestGraph(t *testing.T) {
	var a, b *Typ
	a = &Typ{1}
	b = &Typ{2}
	a.x = 11
	b.x = 22

	graph := NewSparseMatrixGraph()

	graph.AddArc(a, b, nil)

	{
		nodes, _ := graph.Neighbors(a)
		if len(nodes) != 1 {
			t.Errorf("Arc not showing in graph.")
		}
		if nodes[0].(*Typ).x != 22 {
			t.Errorf("Wrong value at the end of arc.")
		}
	}

	{
		nodes, _ := graph.Neighbors(b)
		if len(nodes) != 0 {
			t.Errorf("Nonexistent reverse arc in graph.")
		}
	}

	if _, ok := graph.GetArc(a, b); !ok {
		t.Errorf("Expected arc not found.")
	}
}
