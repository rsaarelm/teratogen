package txt

import (
	"container/vector"
	"fmt"
	"hyades/keyboard"
)

type TextInput struct {
	inputHistory vector.StringVector
	historyPos   int
	cursorPos    int
	Input        chan string
}

func NewTextInput() (result *TextInput) {
	result = &TextInput{Input: make(chan string)}
	result.inputHistory.Push("")
	return
}

func (self *TextInput) CursorPos() int {
	return self.cursorPos
}

func (self *TextInput) CurrentInput() string {
	return self.inputHistory.At(self.historyPos)
}

func (self *TextInput) setCurrent(str string) {
	self.inputHistory.Set(self.historyPos, str)
}

func (self *TextInput) setCursor(pos int) {
	self.cursorPos = pos

	if self.cursorPos < 0 {
		self.cursorPos = 0
	} else if self.cursorPos > len(self.CurrentInput()) {
		self.cursorPos = len(self.CurrentInput())
	}
}

func (self *TextInput) moveCursor(delta int) {
	self.setCursor(self.cursorPos + delta)
}

func (self *TextInput) insert(text string) {
	self.setCurrent(self.CurrentInput()[0:self.cursorPos] + text +
		self.CurrentInput()[self.cursorPos:])
}

func (self *TextInput) delete(n int) {
	if self.cursorPos+n > len(self.CurrentInput()) {
		n = len(self.CurrentInput()) - self.cursorPos
	}
	self.setCurrent(self.CurrentInput()[0:self.cursorPos] +
		self.CurrentInput()[self.cursorPos+n:])
}

func (self *TextInput) eatDuplicate() {
	if self.historyPos > 0 &&
		self.inputHistory.At(self.historyPos) == self.inputHistory.At(self.historyPos-1) {
		self.inputHistory.Delete(self.historyPos)
		self.historyPos--
	}
}

func (self *TextInput) Clear() {
	empty := ""
	self.historyPos = self.inputHistory.Len() - 1
	if self.CurrentInput() != empty {
		self.inputHistory.Push(empty)
		self.historyPos++
	}
	self.setCursor(0)
}

func (self *TextInput) historyMove(delta int) {
	newPos := self.historyPos + delta
	if newPos < 0 {
		// Don't go beyond the oldest input.
		newPos = 0
	}

	if newPos >= self.inputHistory.Len() {
		// If going past the newest input, create a new empty input line.
		self.Clear()
	} else {
		self.historyPos = newPos
	}

	self.setCursor(self.cursorPos)
}

func (self *TextInput) HandleKey(keyCode int) {
	if keyCode > 0 && keyCode&keyboard.Nonprintable == 0 {
		if keyCode >= 128 {
			// XXX: Currently only handling basic ASCII.
			keyCode = '?'
		}
		self.insert(fmt.Sprintf("%c", keyCode))

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
		case keyboard.K_UP, keyboard.K_KP8:
			self.historyMove(-1)
		case keyboard.K_DOWN, keyboard.K_KP2:
			self.historyMove(1)
		case keyboard.K_RETURN, keyboard.K_KP_ENTER:
			self.Input <- self.CurrentInput()
			self.eatDuplicate()
			self.Clear()
		}
	}
}
