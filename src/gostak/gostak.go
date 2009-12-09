/*
 Gostak is a concatenative scripting language on top of Go.
*/
package gostak

import (
	"container/vector";
	"reflect";
)

type gostakMode byte const (
	ImmediateMode = iota;
	CompileMode;
)

type opType byte const (
	LiteralNum = iota;
	LiteralString;
	Word;
)

type gostakOp struct {
	numValue float64;
	stringValue string;
	typ opType;
}

type gostakWord struct {
	// Immediate words are executed even when encountered at compile time.
	immediate bool;
	content []gostakOp;
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
	self.dataStack.Push(val);
}

func (self *GostakState) Pop() interface{} {
	return self.dataStack.Pop();
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
			inputs[i] = reflect.NewValue(self.Pop())
		}

		// TODO: Type checking.
		outputs := val.Call(inputs);

		for i := 0; i < len(outputs); i++ {
			self.Push(outputs[i].Interface());
		}
	} else {
		panic("Tried to apply a non-func value.");
	}
}