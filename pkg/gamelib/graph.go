package gamelib

type Graph interface {
	AddArc(node1, node2 interface{}, arcObj interface{});
	RemoveArc(node1, node2 interface{});
	// Iterate through the nodes of the graph.
	Iter() <-chan interface{};
	Neighbors(node interface{}) (nodes []interface{}, arcs []interface{});
	GetArc(node1, node2 interface{}) (arc interface{}, found bool);
}

type SparseMatrixGraph struct {
	arcMatrix map[uintptr] (map[uintptr] interface{});
	nodeLookup *ObjLookup;
}

func NewSparseMatrixGraph() (result *SparseMatrixGraph) {
	result = new(SparseMatrixGraph);
	result.arcMatrix = make(map[uintptr] (map[uintptr] interface{}));
	result.nodeLookup = NewObjLookup();

	return;
}

func (self *SparseMatrixGraph)AddArc(node1, node2 interface{}, arcObj interface{}) {
	id1 := self.nodeLookup.IncrObj(node1);
	id2 := self.nodeLookup.IncrObj(node2);

	arcList, ok := self.arcMatrix[id1];
	// There aren't any arcs from node1 yet. Add a map for the arcs.
	if !ok {
		arcList = make(map[uintptr] interface{});
		self.arcMatrix[id1] = arcList;
	}
	arcList[id2] = arcObj;
}

func (self *SparseMatrixGraph)RemoveArc(node1, node2 interface{}) {
	id1, id2 := ObjId(node1), ObjId(node2);
	self.nodeLookup.DecrObj(node1);
	self.nodeLookup.DecrObj(node2);

	if arcList, ok := self.arcMatrix[id1]; ok {
		arcList[id2] = nil, false;
		if len(arcList) == 0 {
			self.arcMatrix[id1] = arcList, false;
		}
	}
}

func (self *SparseMatrixGraph)Iter() <-chan interface{} { return self.nodeLookup.Iter(); }

// Returns the neighbor nodes and the arcs to them from a node.
// XXX: Some kind of wrapper object here to make iterating this a bit less painful.
func (self *SparseMatrixGraph)Neighbors(node interface{}) (nodes []interface{}, arcs []interface{}) {
	if neighbors, ok := self.arcMatrix[ObjId(node)]; ok {
		nodes = make([]interface{}, len(neighbors));
		arcs = make([]interface{}, len(neighbors));
		i := 0;
		for nodeAddr, arc := range neighbors {
			// Cast the stored address back to the pointer.
			neighborNode, ok := self.nodeLookup.GetObj(nodeAddr);
			Assert(ok, "Graph node not found in node lookup.");
			nodes[i] = neighborNode;
			arcs[i] = arc;
			i++;
		}
	} else {
		nodes = make([]interface{}, 0);
		arcs = make([]interface{}, 0);
	}
	return;
}

// Get the arc between two nodes if one exists. Note that arc objects may be
// nil and the graph may still have valid arcs, build logic around the boolean
// secondary return value.
func (self *SparseMatrixGraph)GetArc(node1, node2 interface{}) (arc interface{}, found bool) {

	if neighbors, ok1 := self.arcMatrix[ObjId(node1)]; ok1 {
		if a, ok2 := neighbors[ObjId(node2)]; ok2 {
			arc = a;
			found = ok2;
		}
	}
	return;
}
