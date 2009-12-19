package mem

import (
	"io"
	"math"
	"os"
	"strings"
)

func WriteInt32(out io.Writer, num int32) os.Error {
	buf := make([]byte, 4)
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(num % 0x100)
		num >>= 8
	}
	_, err := out.Write(buf)
	return err
}

func ReadInt32(in io.Reader) (result int32, err os.Error) {
	buf := make([]byte, 4)
	_, err = in.Read(buf)
	if err != nil {
		return
	}
	for i := len(buf) - 1; i >= 0; i-- {
		result <<= 8
		result += int32(buf[i])
	}
	return
}

func WriteInt64(out io.Writer, num int64) os.Error {
	buf := make([]byte, 8)
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(num % 0x100)
		num >>= 8
	}
	_, err := out.Write(buf)
	return err
}

func ReadInt64(in io.Reader) (result int64, err os.Error) {
	buf := make([]byte, 8)
	_, err = in.Read(buf)
	if err != nil {
		return
	}
	for i := len(buf) - 1; i >= 0; i-- {
		result <<= 8
		result += int64(buf[i])
	}
	return
}

func WriteFloat64(out io.Writer, num float64) os.Error {
	return WriteInt64(out, int64(math.Float64bits(num)))
}

func ReadFloat64(in io.Reader) (result float64, err os.Error) {
	b, err := ReadInt64(in)
	if err != nil {
		return
	}
	result = math.Float64frombits(uint64(b))
	return
}

func WriteString(out io.Writer, str string) (err os.Error) {
	err = WriteInt32(out, int32(len(str)))
	if err != nil {
		return
	}
	_, err = out.Write(strings.Bytes(str))
	return
}

func ReadString(in io.Reader) (result string, err os.Error) {
	length, err := ReadInt32(in)
	if err != nil {
		return
	}
	buf := make([]byte, length)
	_, err = in.Read(buf)
	if err != nil {
		return
	}
	result = string(buf)
	return
}
