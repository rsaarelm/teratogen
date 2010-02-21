package mem

import (
	"encoding/binary"
	"gob"
	"hyades/dbg"
	"io"
	"strings"
)

type Serializable interface {
	Serialize(out io.Writer)
	Deserialize(in io.Reader)
}

var endianness = binary.LittleEndian

// XXX: The functions don't propagate errors, since that would clutter up the
// serialization routines too much. Should make a more sophisticated error
// handling system here, since this stuff deals with data from outside the
// program and therefore shouldn't use assertions for error handling.

// WriteFixed writes a binary representation of data into out. The data must be
// a fixed size value, either a numeric primitive type or an array or a struct
// containing only fixed-size values.
func WriteFixed(out io.Writer, data interface{}) {
	// Simplify the one in package binary by fixing endianness and crashing on
	// error.
	err := binary.Write(out, endianness, data)
	dbg.AssertNoError(err)
}

// ReadFixed reads a binary representation of a fixed data from in into a
// pointer to the data value, dataPtr.
func ReadFixed(in io.Reader, dataPtr interface{}) {
	err := binary.Read(in, endianness, dataPtr)
	dbg.AssertNoError(err)
}

func ReadByte(in io.Reader) byte {
	var result byte
	ReadFixed(in, &result)
	return result
}

func ReadInt32(in io.Reader) int32 {
	var result int32
	ReadFixed(in, &result)
	return result
}

func ReadInt64(in io.Reader) int64 {
	var result int64
	ReadFixed(in, &result)
	return result
}

func ReadFloat64(in io.Reader) float64 {
	var result float64
	ReadFixed(in, &result)
	return result
}

func WriteString(out io.Writer, str string) {
	WriteFixed(out, int32(len(str)))
	if len(str) == 0 {
		return
	}
	_, err := out.Write(strings.Bytes(str))
	dbg.AssertNoError(err)
}

func ReadString(in io.Reader) string {
	buf := make([]byte, ReadInt32(in))
	if len(buf) == 0 {
		return string(buf)
	}
	_, err := in.Read(buf)
	dbg.AssertNoError(err)
	return string(buf)
}

func WriteNTimes(out io.Writer, count int, write func(int, io.Writer)) {
	WriteFixed(out, int32(count))
	for i := 0; i < count; i++ {
		write(i, out)
	}
}

// Read a count of items from instream. First call func init with the number
// of items, then sequentially call the read function with the current index
// item count times.
func ReadNTimes(in io.Reader, init func(int), read func(int, io.Reader)) {
	count := int(ReadInt32(in))
	init(count)
	for i := 0; i < count; i++ {
		read(i, in)
	}
}


// GobSerialize writes val into the output stream using the gob package to
// serialize it.
func GobSerialize(out io.Writer, val interface{}) {
	enc := gob.NewEncoder(out)
	err := enc.Encode(val)
	dbg.AssertNoError(err)
}

// GobDeserialize decodes a value serialized with the gob package into the
// given struct value.
func GobDeserialize(in io.Reader, val interface{}) {
	dec := gob.NewDecoder(in)
	err := dec.Decode(val)
	dbg.AssertNoError(err)
}

// GobOrMethodSerialize checks if val implements the Serializable interface.
// If it does, it serializes val using val's Serialize method. Otherwise it
// uses the gob package.
func GobOrMethodSerialize(out io.Writer, val interface{}) {
	if ser, ok := val.(Serializable); ok {
		ser.Serialize(out)
	} else {
		GobSerialize(out, val)
	}
}

// GobOrMethodDeserialize checks if val implements the Serializable interface.
// If it does, it deserializes to val using val's Deserialize method.
// Otherwise it uses the gob package.
func GobOrMethodDeserialize(in io.Reader, val interface{}) {
	if ser, ok := val.(Serializable); ok {
		ser.Deserialize(in)
	} else {
		GobDeserialize(in, val)
	}
}
