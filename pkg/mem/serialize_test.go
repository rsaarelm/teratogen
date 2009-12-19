package mem

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func serializeRoundtrip(t *testing.T, obj Serializable, checkObjEquality bool) {
	buf := new(bytes.Buffer)
	err := obj.Serialize(buf)
	if err != nil {
		t.Errorf("Serialization error: %v", err)
	}

	blank := BlankCopy(obj).(Serializable)
	err = blank.Deserialize(bytes.NewBuffer(buf.Bytes()))
	if err != nil {
		t.Errorf("Deserialization error: %v", err)
	}

	if checkObjEquality && !reflect.DeepEqual(obj, blank) {
		t.Errorf("Deserialized %v not deep-equal to original %v.", blank, obj)
	}

	buf2 := new(bytes.Buffer)
	err = blank.Serialize(buf2)
	if err != nil {
		t.Errorf("Second object serialization error: %v", err)
	}

	if !reflect.DeepEqual(buf.Bytes(), buf2.Bytes()) {
		t.Errorf("Roundtrip deserialization doesn't match original deserialization.")
	}
}

type Simple struct {
	x int
}

func (self *Simple) Serialize(out io.Writer) os.Error {
	return WriteInt32(out, int32(self.x))
}

func (self *Simple) Deserialize(in io.Reader) os.Error {
	// REALLY too verbose...
	x32, err := ReadInt32(in)
	self.x = int(x32)
	return err
}

func TestSerialize(t *testing.T) {
	obj := &Simple{123}
	serializeRoundtrip(t, obj, true)
}
