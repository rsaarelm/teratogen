package fomalhaut

import "unsafe"

// Create an identifier for an object from its memory address.
func Obj2Id(obj interface{}) uintptr {
	return uintptr(unsafe.Pointer(&obj));
}

// Turn the identifier back to the object. XXX: Crashes and burns if the
// runtime does anything clever like moving live objects around in memory.
func Id2Obj(id uintptr) (result interface{}) {
	// XXX: unsafe.Typeof(result) is kinda ugly, any more straightforward
	// way to write the interface{} type value as a literal?
	result = unsafe.Unreflect(unsafe.Typeof(result), unsafe.Pointer(id));
	return;
}