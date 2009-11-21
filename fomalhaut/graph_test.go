package fomalhaut

import "testing";
import . "fmt"

type Typ struct {
	a int;
}

func TestGraph(t *testing.T) {
	var a, b Typ;
	a = Typ{1};
	b = Typ{2};

	graph := NewGraph();
	Printf("&a: %v\n", Obj2Id(a));
	graph.AddArc(a, b, nil);
	Printf("Graph test: %v %v %v\n", a, b, graph);
	if arc, ok := graph.GetArc(a, b); ok {
		Printf("%v %v\n", arc, ok);
	} else {
		t.Errorf("Expected arc not found.");
	}
}
