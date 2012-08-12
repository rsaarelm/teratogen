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
	"unsafe"
)

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
	if s.RepeatLength > 2 {
		if offset >= s.RepeatStart+s.RepeatLength {
			offset -= s.RepeatStart
			offset %= s.RepeatLength
			offset += s.RepeatStart
		}
	} else if offset >= len(s.Wave) {
		return 0
	}

	return s.Wave[offset]
}

type Mod struct {
	Length       int
	Samples      []Sample
	PatternTable [128]byte
	Patterns     []Pattern
}

func pitchOffset(samplingRate int, pitch int, offset int) int {
	const magicModNum = 7093789. / 2
	return int(float64(offset) * (magicModNum / float64(samplingRate)) / float64(pitch))
}

// Player returns a reader that contains the bytes of the mod's wave data.
func (m *Mod) Player(samplingRate int) io.Reader {
	return &modPlayer{Mod: m, Rate: samplingRate}
}

const numChans = 4

type modPlayer struct {
	Mod  *Mod
	Rate int

	playPos    int
	instrument [numChans]struct {
		sample byte
		offset int
		pitch  int
	}
}

func (mp *modPlayer) Read(p []byte) (n int, err error) {
	for i, _ := range p {
		next, nerr := mp.NextSample()
		if nerr != nil {
			err = nerr
			return
		}
		bitEquivalentByte := *(*byte)(unsafe.Pointer(&next))
		p[i] = bitEquivalentByte
		n++
	}
	return
}

func (mp *modPlayer) NextSample() (s int8, err error) {
	if mp.playPos >= mp.length() {
		err = io.EOF
		return
	}

	if mp.playPos%mp.bytesPerRow() == 0 {
		for i, ch := range mp.currentRow() {
			if ch.Sample() != 0 {
				mp.instrument[i].sample = ch.Sample()
				mp.instrument[i].offset = 0
				mp.instrument[i].pitch = ch.Period()
			}
		}
	}

	var mix int

	for i, inst := range mp.instrument {
		if inst.sample != 0 {
			sample := mp.Mod.Samples[inst.sample-1].Sample8(pitchOffset(mp.Rate, inst.pitch, inst.offset))
			mix += int(sample)
			mp.instrument[i].offset++
		}
	}

	s = int8(mix / numChans)
	mp.playPos++

	return
}

func (mp *modPlayer) bytesPerRow() int {
	// Mods play at 125 BPM, with four rows per beat, so this thing operates
	// at 8.3 Hz.

	return int(float64(mp.Rate) / 8.3)
}

func (mp *modPlayer) length() int {
	return mp.Mod.Length * 64 * mp.bytesPerRow()
}

func (mp *modPlayer) currentRow() Division {
	idx := mp.playPos / mp.bytesPerRow()
	return mp.Mod.Patterns[mp.Mod.PatternTable[idx/64]][idx%64]
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
		result.Samples[i].Wave = make([]int8, int(sampleHeaders[i].Length)*2)
		binary.Read(r, binary.BigEndian, &result.Samples[i].Wave)
	}

	return
}
