package gostak

import (
	"fmt";
	"testing";
)

func TestArith(t *testing.T) {
	fg := NewGostakState();
	fg.DefineNativeWord(".", func (x interface{}) { fmt.Println(x); });
	fg.DefineNativeWord("+", func (x, y float64) float64 { return x + y; });

	progn := []GostakCell{
		GostakCell{LiteralNum, float64(1.0)},
		GostakCell{LiteralNum, float64(2.0)},
		GostakCell{Word, "+"}};
	fg.Eval(progn);
	if fg.Len() != 1 { t.Errorf("Bad stack len %v", fg.Len()); }
	if fg.At(0).(float64) != float64(3.0) { t.Errorf("Bad top value %v", fg.At(0)); }
}