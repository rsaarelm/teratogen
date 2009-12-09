/*
 Forgo is a Forth-like scripting language on top of Go.
*/
package forgo

import (
	"container/vector";
	"reflect";
)

type forgoMode byte const (
	ImmediateMode = iota;
	CompileMode;
)

type opType byte const (
	LiteralNum = iota;
	LiteralString;
	Word;
)

type forgoOp struct {
	numValue float64;
	stringValue string;
	typ opType;
}

type forgoWord struct {
	// Immediate words are executed even when encountered at compile time.
	immediate bool;
	content []forgoOp;
}

type ForgoState struct {
	dataStack *vector.Vector;
	words map[string] interface{};
}

func NewForgoState() (result *ForgoState) {
	result = new(ForgoState);
	result.dataStack = new(vector.Vector);
	result.words = make(map[string] interface{});
	return;
}

func (self *ForgoState) Push(val interface{}) {
	self.dataStack.Push(val);
}

func (self *ForgoState) Pop() interface{} {
	return self.dataStack.Pop();
}

// Use reflection API to convert a function that takes n and returns m values
// into a function that takes the n stack values and pushes the m return
// values to the stack.
func (self *ForgoState) ApplyFunc(fn interface{}) {
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