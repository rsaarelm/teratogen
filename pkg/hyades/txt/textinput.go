package txt

import (
	"container/vector"
	"hyades/keyboard"
	"utf8"
)

type TextInput struct {
	CurrentInput string
	inputHistory vector.StringVector
	cursorPos    int
	Input        chan string
}

func NewTextInput() (result *TextInput) {
	result = &TextInput{Input: make(chan string)}
	return
}

func (self *TextInput) CursorPos() int {
	return self.cursorPos
}

func (self *TextInput) setCursor(pos int) {
	self.cursorPos = pos

	if self.cursorPos < 0 {
		self.cursorPos = 0
	} else if self.cursorPos > len(self.CurrentInput)+1 {
		self.cursorPos = len(self.CurrentInput) + 1
	}
}

func (self *TextInput) moveCursor(delta int) {
	self.setCursor(self.cursorPos + delta)
}

func (self *TextInput) insert(text string) {
	self.CurrentInput = self.CurrentInput[0:self.cursorPos] + text +
		self.CurrentInput[self.cursorPos:]
}

func (self *TextInput) delete(n int) {
	if self.cursorPos+n > len(self.CurrentInput) {
		n = len(self.CurrentInput) - self.cursorPos
	}
	self.CurrentInput = self.CurrentInput[0:self.cursorPos] +
		self.CurrentInput[self.cursorPos+n:]
}

func (self *TextInput) Clear() {
	self.CurrentInput = ""
	self.setCursor(0)
}

func (self *TextInput) HandleKey(keyCode int) {
	if keyCode > 0 && keyCode&keyboard.Nonprintable == 0 {
		buf := make([]byte, 4)
		utf8.EncodeRune(keyCode, buf)
		self.insert(string(buf))

		self.moveCursor(1)
	} else {
		switch keyCode {
		case keyboard.K_BACKSPACE:
			if self.cursorPos > 0 {
				self.moveCursor(-1)
				self.delete(1)
			}
		case keyboard.K_DELETE:
			self.delete(1)
		case keyboard.K_LEFT, keyboard.K_KP4:
			self.moveCursor(-1)
		case keyboard.K_RIGHT, keyboard.K_KP6:
			self.moveCursor(1)
			// TODO: Moving in input history with up & down
		case keyboard.K_RETURN, keyboard.K_KP_ENTER:
			self.Input <- self.CurrentInput
			// TODO: Store CurrentInput in input history
			self.Clear()
		}
	}
}
