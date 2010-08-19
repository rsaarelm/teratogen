package main

import (
	"container/vector"
	"hyades/geom"
	"hyades/gfx"
	"hyades/txt"
	"os"
	"strings"
)

type MsgOut struct {
	lines *vector.StringVector
	input chan string
	// A persistent text input component which remembers old commands.
	commandLine *txt.TextInput
	CursorPos   geom.Pt2I
}

func NewMsgOut() (result *MsgOut) {
	result = new(MsgOut)
	result.lines = new(vector.StringVector)
	result.newLine()
	result.input = make(chan string)
	result.commandLine = txt.NewTextInput()
	result.HideCursor()
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

func (self *MsgOut) HideCursor() {
	self.CursorPos = geom.Pt2I{-1, -1}
}

func (self *MsgOut) InputText(prompt string) (result string) {
	self.WriteString(prompt)

	input := self.commandLine

	running := true
	defer func() { running = false }()

	idx := self.lines.Len() - 1
	prompt = self.lines.At(idx)

	finished := make(chan bool)

	// Display the input area
	go func(anim *gfx.Anim) {
		defer anim.Close()
		for running {
			_, _ = anim.StartDraw()

			self.lines.Set(idx, prompt+input.CurrentInput())
			self.CursorPos = geom.Pt2I{len(prompt) + input.CursorPos(), idx}

			anim.StopDraw()
		}
		finished <- true
	}(ui.AddScreenAnim(gfx.NewAnim(0.0)))

	ui.PushKeyHandler(input)
	result = <-input.Input
	running = false
	_ = <-finished
	self.WriteString(result + "\n")
	ui.PopKeyHandler()
	self.HideCursor()
	return
}
