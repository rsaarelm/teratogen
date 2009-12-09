package gostak

import (
	"testing";
)

func TestArith(t *testing.T) {
	// TODO dynamic typing with graceful error handling...
	fg := NewGostakState();
	fg.Push(2);
	fg.Push(4);
	fg.ApplyFunc(func (x, y int) int { return x + y; });
	ret := fg.Pop();
	if ret.(int) != 6 {
		t.Errorf("Bad return value %v", ret);
	}
}