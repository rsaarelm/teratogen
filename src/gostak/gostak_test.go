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

func TestQuotation(t *testing.T) {
	fg := NewGostakState();
	fg.DefineNativeWord(".", func (x interface{}) { fmt.Println(x); });
	fg.DefineNativeWord("ifte", func (pred bool, t, e interface{}) {
		if pred {
			fg.Eval(t.([]GostakCell));
		} else {
			fg.Eval(e.([]GostakCell));
		}
	});

	progn1 := []GostakCell{
		GostakCell{LiteralBool, true},
		GostakCell{Quotation, []GostakCell{GostakCell{LiteralNum, float64(10.0)}}},
		GostakCell{Quotation, []GostakCell{GostakCell{LiteralNum, float64(20.0)}}},
		GostakCell{Word, "ifte"}};
	fg.Eval(progn1);
	if fg.At(0).(float64) != float64(10.0) { t.Errorf("Bad top value %v", fg.At(0)); }

	progn2 := []GostakCell{
		GostakCell{LiteralBool, false},
		GostakCell{Quotation, []GostakCell{GostakCell{LiteralNum, float64(10.0)}}},
		GostakCell{Quotation, []GostakCell{GostakCell{LiteralNum, float64(20.0)}}},
		GostakCell{Word, "ifte"}};
	fg.Eval(progn2);
	if fg.At(0).(float64) != float64(20.0) { t.Errorf("Bad top value %v", fg.At(0)); }

}