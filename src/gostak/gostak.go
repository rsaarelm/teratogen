/*
 Gostak is a concatenative scripting language on top of Go.
*/
package gostak

import (
	"container/vector";
	"fmt";
	"reflect";
)

type CellType byte const (
	LiteralNum = iota;
	LiteralString;
	LiteralBool;
	Word;
	Quotation;
)

type GostakCell struct {
	typ CellType;
	data interface{};
}

func newNumCell(num float64) *GostakCell {
	return &GostakCell{LiteralNum, num};
}

func newStringCell(str string) *GostakCell {
	return &GostakCell{LiteralString, str};
}

func newBoolCell(b bool) *GostakCell {
	return &GostakCell{LiteralBool, b};
}

type GostakState struct {
	dataStack *vector.Vector;
	words map[string] interface{};
}

func NewGostakState() (result *GostakState) {
	result = new(GostakState);
	result.dataStack = new(vector.Vector);
	result.words = make(map[string] interface{});
	return;
}

func (self *GostakState) Push(val interface{}) {
	fmt.Printf("Pushed %v\n", val);
	self.dataStack.Push(val);
}

func (self *GostakState) Pop() interface{} {
	return self.dataStack.Pop();
}

func (self *GostakState) Eval(cells []GostakCell) {
	for _, cell := range cells {
		switch cell.typ {
		case LiteralNum, LiteralString, LiteralBool, Quotation: self.Push(cell.data)
		case Word:
			prog, ok := self.words[cell.data.(string)];
			if !ok {
				panic(fmt.Sprintf("Word %v not defined.", cell.data.(string)));
			}

			typ := reflect.Typeof(prog);
			if _, ok2 := typ.(*reflect.FuncType); ok2 {
				self.ApplyFunc(prog);
			} else {
				seq := prog.([]GostakCell);
				self.Eval(seq);
			}
		}
	}
}

func (self *GostakState) DefineWord(word string, data []GostakCell) {
	self.words[word] = data;
}

func (self *GostakState) DefineNativeWord(word string, fn interface{}) {
	typ := reflect.Typeof(fn);
	if _, ok := typ.(*reflect.FuncType); !ok {
		panic("Native word definition not a func value.");
	}

	self.words[word] = fn;
}

// Use reflection API to convert a function that takes n and returns m values
// into a function that takes the n stack values and pushes the m return
// values to the stack.
func (self *GostakState) ApplyFunc(fn interface{}) {
	val := reflect.NewValue(fn);

	if val, ok := val.(*reflect.FuncValue); ok {
		typ := val.Type().(*reflect.FuncType);

		inputs := make([]reflect.Value, typ.NumIn());
		// Pop stack values to input list, starting from the end of
		// the list.
		for i := len(inputs) - 1; i >= 0; i-- {
			inputs[i] = reflect.NewValue(self.Pop());
			fmt.Printf("Popped %v\n", inputs[i]);
		}
		fmt.Printf("Calling %v\n", val);

		// TODO: Type checking.
		outputs := val.Call(inputs);

		for i := 0; i < len(outputs); i++ {
			self.Push(outputs[i].Interface());
		}
	} else {
		panic("Tried to apply a non-func value.");
	}
}