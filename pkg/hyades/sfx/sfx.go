// Sound effects package.

package sfx

import (
	"bytes"
	"encoding/binary"
	"hyades/dbg"
	"io"
	"os"
)

type Sound interface {
	Play()
}

type AudioContext interface {
	MakeSound(wavData []byte) (result Sound, err os.Error)
}

type WaveFunc func(t float64) float64

// WAV format reference used:
// http://technology.niagarac.on.ca/courses/ctec1631/WavFileFormat.html

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

const (
	DefaultSampleRate  = Rate44k
	DefaultSampleBytes = Bit16
	DefaultNumChannels = Stereo
)

func sampleMonoSound(out io.Writer, wave WaveFunc, durationSec float64, rate SampleRate, sampleBytes SampleBytes) {
	timeStep := 1.0 / float64(rate)
	for t := float64(0.0); t < durationSec; t += timeStep {
		switch sampleBytes {
		case Bit8:
			binary.Write(out, binary.LittleEndian, int8(wave(t)*0x7f))
		case Bit16:
			binary.Write(out, binary.LittleEndian, int16(wave(t)*0x7fff))
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
			binary.Write(out, binary.LittleEndian, int8(wave1(t)*0x7f))
			binary.Write(out, binary.LittleEndian, int8(wave2(t)*0x7f))
		case Bit16:
			binary.Write(out, binary.LittleEndian, int16(wave1(t)*0x7fff))
			binary.Write(out, binary.LittleEndian, int16(wave2(t)*0x7fff))
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
	binary.Write(out, binary.LittleEndian, uint32(dataLen+extraLength))
	io.WriteString(out, "WAVE")
}

func writeWavFormat(out io.Writer, channels NumChannels, rate SampleRate, sampleBytes SampleBytes) {
	io.WriteString(out, "fmt ")
	// FORMAT chunk length
	binary.Write(out, binary.LittleEndian, uint32(16))
	// unknown
	binary.Write(out, binary.LittleEndian, uint16(1))
	binary.Write(out, binary.LittleEndian, uint16(channels))
	binary.Write(out, binary.LittleEndian, uint32(rate))
	binary.Write(out, binary.LittleEndian, uint32(rate)*uint32(channels)*uint32(sampleBytes))
	binary.Write(out, binary.LittleEndian, uint16(sampleBytes))
	binary.Write(out, binary.LittleEndian, uint16(8*sampleBytes))
}

func writeWavData(out io.Writer, dataLen int) {
	io.WriteString(out, "data")
	binary.Write(out, binary.LittleEndian, uint32(dataLen))
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
		binary.Write(out, binary.LittleEndian, i)
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

func MonoWaveToSound(context AudioContext, wave WaveFunc, durationSec float64) (result Sound, err os.Error) {
	buf := new(bytes.Buffer)
	SampleMonoWav(buf, wave, durationSec, DefaultSampleRate, DefaultSampleBytes)
	return context.MakeSound(buf.Bytes())
}

func StereoWaveToSound(context AudioContext, wave1, wave2 WaveFunc, durationSec float64) (result Sound, err os.Error) {
	buf := new(bytes.Buffer)
	SampleStereoWav(buf, wave1, wave2, durationSec, DefaultSampleRate, DefaultSampleBytes)
	return context.MakeSound(buf.Bytes())
}
