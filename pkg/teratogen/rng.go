package teratogen

import (
	"bytes"
	"encoding/binary"
	"hyades/babble"
	"hyades/num"
	"os"
	"unsafe"
)

func RandStateToBabble(state num.RandState) string {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, state)
	return babble.EncodeToString(buf.Bytes())
}

func BabbleToRandState(bab string) (result num.RandState, err os.Error) {
	data, err := babble.DecodeString(bab)
	if err != nil {
		return
	}
	if len(data) != unsafe.Sizeof(num.RandState(0)) {
		err = os.NewError("Bad babble data length.")
		return
	}
	var state num.RandState
	err = binary.Read(bytes.NewBuffer(data), binary.BigEndian, &state)
	result = state
	return
}
