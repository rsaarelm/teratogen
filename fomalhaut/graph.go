package fomalhaut

type SparseMatrixGraph struct {
	arcMatrix map[uintptr] (map[uintptr] interface{});
}

func NewGraph() (result *SparseMatrixGraph) {
	const initialNodeCapacity = 32;

	result = new(SparseMatrixGraph);
	result.arcMatrix = make(map[uintptr] (map[uintptr] interface{}));

	return;
}

func (self *SparseMatrixGraph)AddArc(node1, node2 interface{}, arcObj interface{}) {
	id1, id2 := Obj2Id(node1), Obj2Id(node2);
	arcList, ok := self.arcMatrix[id1];
	// There aren't any arcs from node1 yet. Add a map for the arcs.
	if !ok {
		arcList = make(map[uintptr] interface{});
		self.arcMatrix[id1] = arcList;
	}
	arcList[id2] = arcObj;
}

func (self *SparseMatrixGraph)RemoveArc(node1, node2 interface{}) {
	id1, id2 := Obj2Id(node1), Obj2Id(node2);
	if arcList, ok := self.arcMatrix[id1]; ok {
		arcList[id2] = nil, false;
		if len(arcList) == 0 {
			// FIXME: Issue 288, deleting from map of maps won't work.
//			self.arcMatrix[id1] = arcList, false;
		}
	}
}

// Returns the neighbor nodes and the arcs to them from a node.
func (self *SparseMatrixGraph)Neighbors(node interface{}) (nodes []interface{}, arcs []interface{}) {
	if neighbors, ok := self.arcMatrix[Obj2Id(node)]; ok {
		nodes = make([]interface{}, len(neighbors));
		arcs = make([]interface{}, len(neighbors));
		i := 0;
		for nodeAddr, arc := range neighbors {
			// Cast the stored address back to the pointer.
			neighborNode := Id2Obj(nodeAddr);
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
func (self *SparseMatrixGraph)GetArc(node1, node2 interface{})
	(arc interface{}, ok bool) {

	if neighbors, ok1 := self.arcMatrix[Obj2Id(node1)]; ok1 {
		if a, ok2 := neighbors[Obj2Id(node2)]; ok2 {
			arc = a;
			ok = ok2;
		}
	}
	return;
}