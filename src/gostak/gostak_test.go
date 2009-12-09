package gostak

import (
	"fmt";
	"testing";
)

func TestArith(t *testing.T) {
	// TODO dynamic typing with graceful error handling...
	fg := NewGostakState();
	fg.DefineNativeWord(".", func (x interface{}) { fmt.Println(x); });
	fg.DefineNativeWord("+", func (x, y float64) float64 { return x + y; });

	progn := []GostakCell{
		GostakCell{LiteralNum, float64(1.0)},
		GostakCell{LiteralNum, float64(2.0)},
		GostakCell{Word, "+"},
		GostakCell{Word, "."}};
	fg.Eval(progn);
}