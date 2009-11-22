package fomalhaut

type Field2 interface {
	Get(x, y int) (interface{}, bool);
	Set(x, y int, item interface{});
}

type MapField2 struct {
	data map[int] interface{};
}

func NewMapField2() (result *MapField2) {
	result = new(MapField2);
	result.data = make(map[int] interface{}, 32);
	return;
}

const mapField2Dim = 8192;

// XXX: Hacky mapping of points into a fixed-size rectangle, since we don't
// have tuple keys.
func (f2 *MapField2) encodePoint(x, y int) int {
	return (x - mapField2Dim / 2) + (y - mapField2Dim / 2) * mapField2Dim;
}

func (f2 *MapField2) Contains(x, y int) bool {
	return x >= -mapField2Dim / 2 && y >= -mapField2Dim / 2 &&
		x < mapField2Dim / 2 && y < mapField2Dim / 2;
}

func (f2 *MapField2) Get(x, y int) (ret interface{}, found bool) {
	if f2.Contains(x, y) {
		ret, found = f2.data[f2.encodePoint(x, y)];
	} else {
		found = false;
	}
	return;
}

func (f2 *MapField2) Set(x, y int, item interface{}) {
	if f2.Contains(x, y) {
		f2.data[f2.encodePoint(x, y)] = item;
	} else {
		Die("Point outside region this type of field can handle.");
	}
}
