package mem

import (
	"hyades/dbg"
	"io"
	"math"
	"strings"
)

type Serializable interface {
	Serialize(out io.Writer)
	Deserialize(in io.Reader)
}

// XXX: The functions don't propagate errors, since that would clutter up the
// serialization routines too much. Should make a more sophisticated error
// handling system here, since this stuff deals with data from outside the
// program and therefore shouldn't use assertions for error handling.

func WriteByte(out io.Writer, b byte) {
	_, err := out.Write([]byte{b})
	dbg.AssertNoError(err)
}

func ReadByte(in io.Reader) byte {
	var buf = make([]byte, 1)
	_, err := in.Read(buf)
	dbg.AssertNoError(err)
	return buf[0]
}

func WriteInt32(out io.Writer, num int32) {
	buf := make([]byte, 4)
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(num % 0x100)
		num >>= 8
	}
	_, err := out.Write(buf)
	dbg.AssertNoError(err)
}

func ReadInt32(in io.Reader) (result int32) {
	buf := make([]byte, 4)
	_, err := in.Read(buf)
	dbg.AssertNoError(err)
	for i := len(buf) - 1; i >= 0; i-- {
		result <<= 8
		result += int32(buf[i])
	}
	return
}

func WriteInt64(out io.Writer, num int64) {
	buf := make([]byte, 8)
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(num % 0x100)
		num >>= 8
	}
	_, err := out.Write(buf)
	dbg.AssertNoError(err)
}

func ReadInt64(in io.Reader) (result int64) {
	buf := make([]byte, 8)
	_, err := in.Read(buf)
	dbg.AssertNoError(err)
	for i := len(buf) - 1; i >= 0; i-- {
		result <<= 8
		result += int64(buf[i])
	}
	return
}

func WriteFloat64(out io.Writer, num float64) { WriteInt64(out, int64(math.Float64bits(num))) }

func ReadFloat64(in io.Reader) float64 { return math.Float64frombits(uint64(ReadInt64(in))) }

func WriteString(out io.Writer, str string) {
	WriteInt32(out, int32(len(str)))
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
	WriteInt32(out, int32(count))
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
