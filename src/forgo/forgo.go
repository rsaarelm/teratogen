/*
 Forgo is a Forth-like scripting language on top of Go.
*/
package forgo

import (
	"container/vector";
)

type ForgoState struct {
	dataStack vector.Vector;
}

func (self *ForgoState) Push(val interface{}) {
	(&self.dataStack).Push(val);
}

func (self *ForgoState) Pop() interface{} {
	return (&self.dataStack).Pop();
}