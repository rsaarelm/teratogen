// audio.go
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

package sdl

/*
#include <SDL/SDL.h>
int openAudio(SDL_AudioSpec* spec);
void play(Uint8* data, int nbytes);
void stop();
*/
import "C"

import (
	"io"
	"io/ioutil"
	"unsafe"
)

const SamplingRate = 11025

func initAudio() {
	// Init a simple audio setup.
	var spec C.SDL_AudioSpec
	spec.freq = C.int(SamplingRate)
	spec.format = C.AUDIO_S8
	spec.channels = 1
	spec.samples = 512
	spec.userdata = nil
	C.openAudio(&spec)
}

// BufferAudio adds 8-bit audio to the play buffer.
func BufferAudio(sound []int8) {
	if len(sound) == 0 {
		return
	}
	mutex.Lock()
	C.play((*C.Uint8)(unsafe.Pointer(&sound[0])), C.int(len(sound)))
	mutex.Unlock()
}

func Play(input io.Reader) error {
	buf, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}
	mutex.Lock()
	C.play((*C.Uint8)(unsafe.Pointer(&buf[0])), C.int(len(buf)))
	mutex.Unlock()
	return nil
}

func ClearAudioBuffer() {
	mutex.Lock()
	C.stop()
	mutex.Unlock()
}
