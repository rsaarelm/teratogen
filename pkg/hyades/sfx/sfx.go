// Sound effects package.

package sfx

import (
	"bytes"
	"hyades/dbg"
	"io"
	"unsafe"
)

type WaveFunc func(t float64) float64

// WAV format reference used:
// http://technology.niagarac.on.ca/courses/ctec1631/WavFileFormat.html

// TODO: Move byte output stuff to its own library
// Little-endian byte output
func writeUint32(out io.Writer, val uint32) {
	out.Write([]byte{
		byte(val % 0x100),
		byte((val >> 8) % 0x100),
		byte((val >> 16) % 0x100),
		byte((val >> 24) % 0x100),
	})
}

func writeUint16(out io.Writer, val uint16) {
	out.Write([]byte{
		byte(val % 0x100),
		byte((val >> 8) % 0x100),
	})
}

func writeInt16(out io.Writer, val int16) {
	writeUint16(out, *(*uint16)(unsafe.Pointer(&val)))
}

func writeInt8(out io.Writer, val int8) { out.Write([]byte{*(*byte)(unsafe.Pointer(&val))}) }

type NumChannels uint16

const (
	Mono   NumChannels = 1
	Stereo NumChannels = 2
)

type SampleBytes uint16

const (
	Bit8  SampleBytes = 1
	Bit16 SampleBytes = 2
)

type SampleRate uint32

const (
	Rate8k   SampleRate = 8000
	Rate11k  SampleRate = 11025
	Rate16k  SampleRate = 16000
	Rate22k  SampleRate = 22050
	Rate32k  SampleRate = 3200
	Rate44k  SampleRate = 44100
	Rate48k  SampleRate = 48000
	Rate96k  SampleRate = 96000
	Rate192k SampleRate = 192000
)

func sampleMonoSound(out io.Writer, wave WaveFunc, durationSec float64, rate SampleRate, sampleBytes SampleBytes) {
	timeStep := 1.0 / float64(rate)
	for t := float64(0.0); t < durationSec; t += timeStep {
		switch sampleBytes {
		case Bit8:
			writeInt8(out, int8(wave(t)*0x7f))
		case Bit16:
			writeInt16(out, int16(wave(t)*0x7fff))
		default:
			dbg.Die("Bad sample bytes %v", sampleBytes)
		}
	}
}

func sampleStereoSound(out io.Writer, wave1, wave2 WaveFunc, durationSec float64, rate SampleRate, sampleBytes SampleBytes) {
	timeStep := 1.0 / float64(rate)
	for t := float64(0.0); t < durationSec; t += timeStep {
		switch sampleBytes {
		case Bit8:
			writeInt8(out, int8(wave1(t)*0x7f))
			writeInt8(out, int8(wave2(t)*0x7f))
		case Bit16:
			writeInt16(out, int16(wave1(t)*0x7fff))
			writeInt16(out, int16(wave2(t)*0x7fff))
		default:
			dbg.Die("Bad sample bytes %v", sampleBytes)
		}
	}
}

func writeWavRiff(out io.Writer, dataLen int) {
	const headerLen = 44
	const extraLength = headerLen - 8

	io.WriteString(out, "RIFF")
	// Write number of bytes to come. Length of data + headerLen - the 8
	// bytes already written.
	writeUint32(out, uint32(dataLen+extraLength))
	io.WriteString(out, "WAVE")
}

func writeWavFormat(out io.Writer, channels NumChannels, rate SampleRate, sampleBytes SampleBytes) {
	io.WriteString(out, "fmt ")
	// FORMAT chunk length
	writeUint32(out, 16)
	// unknown
	writeUint16(out, 1)
	writeUint16(out, uint16(channels))
	writeUint32(out, uint32(rate))
	writeUint32(out, uint32(rate)*uint32(channels)*uint32(sampleBytes))
	writeUint16(out, uint16(sampleBytes))
	writeUint16(out, 8*uint16(sampleBytes))
}

func writeWavData(out io.Writer, dataLen int) {
	io.WriteString(out, "data")
	writeUint32(out, uint32(dataLen))
}

func WriteWav(out io.Writer, channels NumChannels, rate SampleRate, sampleBytes SampleBytes, data []byte) {
	writeWavRiff(out, len(data))
	writeWavFormat(out, channels, rate, sampleBytes)
	writeWavData(out, len(data))
	out.Write(data)
}

func WriteWav16(out io.Writer, channels NumChannels, rate SampleRate, data []uint16) {
	writeWavRiff(out, len(data)*2)
	writeWavFormat(out, channels, rate, Bit16)
	writeWavData(out, len(data)*2)
	for _, i := range data {
		writeUint16(out, i)
	}
}

func SampleMonoWav(out io.Writer, wave WaveFunc, durationSec float64, rate SampleRate, sampleBytes SampleBytes) {
	buf := new(bytes.Buffer)
	sampleMonoSound(buf, wave, durationSec, rate, sampleBytes)
	WriteWav(out, Mono, rate, sampleBytes, buf.Bytes())
}

func SampleStereoWav(out io.Writer, wave1, wave2 WaveFunc, durationSec float64, rate SampleRate, sampleBytes SampleBytes) {
	buf := new(bytes.Buffer)
	sampleStereoSound(buf, wave1, wave2, durationSec, rate, sampleBytes)
	WriteWav(out, Stereo, rate, sampleBytes, buf.Bytes())
}
