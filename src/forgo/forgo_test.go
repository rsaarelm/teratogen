package forgo

import (
	"testing";
)

func TestArith(t *testing.T) {
	// TODO dynamic typing with graceful error handling...
	fg := new(ForgoState);
	fg.Push(2);
	fg.Push(4);
	fg.Push(fg.Pop().(int) + fg.Pop().(int));
	ret := fg.Pop();
	if ret.(int) != 6 {
		t.Errorf("Bad return value %v", ret);
	}
}