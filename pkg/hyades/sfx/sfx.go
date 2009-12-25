// Sound effects package.

package sfx

import (
	"bytes"
	"io"
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

func MakeMono8Wav(wave WaveFunc, rateHz uint32, durationSec float64) []byte {
	return MakeMono8WavFrom(MakeSound8Bit(wave, rateHz, durationSec),
		rateHz)
}

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

func writeWavRiff(out io.Writer, dataLen int) {
	const headerLen = 44
	const extraLength = headerLen - 8

	io.WriteString(out, "RIFF")
	// Write number of bytes to come. Length of data + headerLen - the 8
	// bytes already written.
	writeUint32(out, uint32(dataLen+extraLength))
	io.WriteString(out, "WAVE")
}

func writeWavFormat(out io.Writer, channels NumChannels, rate SampleRate, bytes SampleBytes) {
	io.WriteString(out, "fmt ")
	// FORMAT chunk length
	writeUint32(out, 16)
	// unknown
	writeUint16(out, 1)
	writeUint16(out, uint16(channels))
	writeUint32(out, uint32(rate))
	writeUint32(out, uint32(rate)*uint32(channels)*uint32(bytes))
	writeUint16(out, uint16(bytes))
	writeUint16(out, 8*uint16(bytes))
}

func writeWavData(out io.Writer, dataLen int) {
	io.WriteString(out, "data")
	writeUint32(out, uint32(dataLen))
}

func WriteWav8(out io.Writer, channels NumChannels, rate SampleRate, data []byte) {
	writeWavRiff(out, len(data))
	writeWavFormat(out, channels, rate, Bit8)
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

func MakeMono8WavFrom(data []byte, rateHz uint32) []byte {
	buf := new(bytes.Buffer)

	writeWavRiff(buf, len(data))

	writeWavFormat(buf, Mono, SampleRate(rateHz), Bit8)

	// DATA Chunk
	buf.WriteString("data")
	writeUint32(buf, uint32(len(data)))

	buf.Write(data)

	return buf.Bytes()
}

// Convert a wave function that maps time in seconds into an amplitude between
// 1 and -1 into a sample of the given rate and duration.
func MakeSound8Bit(wave WaveFunc, rateHz uint32, durationSec float64) []byte {
	buf := new(bytes.Buffer)

	timeStep := 1.0 / float64(rateHz)
	for t := float64(0.0); t < durationSec; t += timeStep {
		sample := byte((wave(t) + 1.0) / 2.0 * 255.0)
		buf.WriteByte(sample)
	}
	return buf.Bytes()
}
