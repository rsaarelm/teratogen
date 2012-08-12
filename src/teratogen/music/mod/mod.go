// mod.go
//
// Copyright (C) 2012 Risto Saarelma
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package mod

import (
	"encoding/binary"
	"io"
)

const internalFreq = 16574 // http://www.eblong.com/zarf/blorb/mod-spec.txt

type sampleHeader struct {
	Name         [22]byte
	Length       int16
	Tune         uint8
	Volume       uint8
	RepeatStart  int16
	RepeatLength int16
}

type Pattern [64]Division

type Division [4]Channel

type Channel [4]byte

func (c Channel) Sample() byte {
	return c[0]&0xF0 | (c[2] & 0xF0 >> 4)
}

func (c Channel) Period() int {
	return int(c[0]&0x0F)<<8 | int(c[1])
}

func (c Channel) Effect() int {
	return int(c[2] & 0x0F)
}

func (c Channel) EffectParams() int {
	return int(c[3])
}

type Sample struct {
	Name         string
	RepeatStart  int
	RepeatLength int
	Wave         []int8
}

func (s Sample) Sample8(offset int) int8 {
	if len(s.Wave) == 0 {
		return 0
	}
	if offset >= len(s.Wave) {
		if s.RepeatLength > 0 {
			offset -= s.RepeatStart
			offset %= len(s.Wave) - s.RepeatStart
			offset += s.RepeatStart
		} else {
			return 0
		}
	}
	return s.Wave[offset]
}

func (s Sample) ResampleLen(freq int) int {
	if freq == 0 {
		return 0
	}
	return len(s.Wave) * internalFreq / freq
}

func (s Sample) Resample(freq int, out []int8) {
	resLen := s.ResampleLen(freq)
	for i := 0; i < resLen; i++ {
		out[i] = s.Wave[i*len(s.Wave)/resLen]
	}
}

type Mod struct {
	Length       int
	Samples      []Sample
	PatternTable [128]byte
	Patterns     []Pattern
}

func (m *Mod) Sample(t float64) (x float64) {
	// TODO
	return 0
}

func Decode(r io.Reader) (result *Mod, err error) {
	var title [20]byte
	binary.Read(r, binary.BigEndian, &title)
	var sampleHeaders [31]sampleHeader
	for i, _ := range sampleHeaders {
		binary.Read(r, binary.BigEndian, &sampleHeaders[i])
	}

	result = new(Mod)
	var length, dummy byte
	binary.Read(r, binary.BigEndian, &length)
	binary.Read(r, binary.BigEndian, &dummy)
	result.Length = int(length)

	binary.Read(r, binary.BigEndian, &result.PatternTable)

	numPatterns := 0
	for _, n := range result.PatternTable {
		if int(n)+1 > numPatterns {
			numPatterns = int(n) + 1
		}
	}

	// Can have various mod file identifying values. Can tell a 15-sample file
	// by the absence of this.
	var id [4]byte
	binary.Read(r, binary.BigEndian, &id)

	result.Patterns = make([]Pattern, numPatterns)
	for i, _ := range result.Patterns {
		binary.Read(r, binary.BigEndian, &result.Patterns[i])
	}

	result.Samples = make([]Sample, len(sampleHeaders))
	for i, _ := range result.Samples {
		result.Samples[i] = Sample{
			Name:         string(sampleHeaders[i].Name[:]),
			RepeatStart:  int(sampleHeaders[i].RepeatStart) * 2,
			RepeatLength: int(sampleHeaders[i].RepeatLength) * 2}
		result.Samples[i].Wave = make([]int8, sampleHeaders[i].Length*2)
		binary.Read(r, binary.BigEndian, &result.Samples[i].Wave)
	}

	return
}
