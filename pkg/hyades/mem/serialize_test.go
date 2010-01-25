package mem

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func serializeRoundtrip(t *testing.T, obj Serializable, checkObjEquality bool) {
	buf := new(bytes.Buffer)
	obj.Serialize(buf)

	blank := BlankCopy(obj).(Serializable)
	blank.Deserialize(bytes.NewBuffer(buf.Bytes()))

	if checkObjEquality && !reflect.DeepEqual(obj, blank) {
		t.Errorf("Deserialized %v not deep-equal to original %v.", blank, obj)
	}

	buf2 := new(bytes.Buffer)
	blank.Serialize(buf2)

	if !reflect.DeepEqual(buf.Bytes(), buf2.Bytes()) {
		t.Errorf("Roundtrip deserialization doesn't match original deserialization.")
	}
}

type Simple struct {
	x int
}

func (self *Simple) Serialize(out io.Writer) { WriteFixed(out, int32(self.x)) }

func (self *Simple) Deserialize(in io.Reader) { self.x = int(ReadInt32(in)) }

func TestSerialize(t *testing.T) {
	obj := &Simple{123}
	serializeRoundtrip(t, obj, true)
}
