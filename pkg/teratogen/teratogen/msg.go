package teratogen

import (
	"container/vector"
	"os"
	"strings"
)

type MsgOut struct {
	lines *vector.StringVector
	input chan string
}

func NewMsgOut() (result *MsgOut) {
	result = new(MsgOut)
	result.lines = new(vector.StringVector)
	result.newLine()
	result.input = make(chan string)
	return
}

func (self *MsgOut) newLine() { self.lines.Push("") }

// Can use negative indices to get the last lines.
func (self *MsgOut) GetLine(idx int) string {
	if idx < 0 {
		idx = self.lines.Len() + idx
	}
	if idx < 0 || idx >= self.lines.Len() {
		return ""
	}
	return self.lines.At(idx)
}

func (self *MsgOut) NumLines() int { return self.lines.Len() }

func (self *MsgOut) WriteString(str string) {
	newLineIdx := strings.Index(str, "\n")
	if newLineIdx != -1 {
		// If newline found, make a new line.
		self.WriteString(str[0:newLineIdx])
		self.newLine()
		self.WriteString(str[newLineIdx+1:])
	} else {
		// Append text to last line.
		idx := self.lines.Len() - 1
		self.lines.Set(idx, self.lines.At(idx)+str)
	}
}

func (self *MsgOut) Write(p []byte) (n int, err os.Error) {
	self.WriteString(string(p))
	n = len(p)
	return
}

func Msg(format string, a ...) { /* XXX Dummy function */ }
