// Sound effects package.

package sfx

import (
	"bytes"
	"io"
)

// WAV format reference used:
// http://technology.niagarac.on.ca/courses/ctec1631/WavFileFormat.html

// Little-endian byte output
func writeUint32(out io.Writer, val uint32) {
	out.Write([]byte{
		byte(val % 0x100),
		byte((val >> 8) % 0x100),
		byte((val >> 16) % 0x100),
		byte((val >> 24) % 0x100)})
}

func writeUint16(out io.Writer, val uint16) {
	out.Write([]byte{
		byte(val % 0x100),
		byte((val >> 8) % 0x100)})
}

func MakeMono8Wav(wave func(float) float, rateHz uint32, durationSec float) []byte {
	return MakeMono8WavFrom(MakeSound8Bit(wave, rateHz, durationSec),
		rateHz)
}

func MakeMono8WavFrom(data []byte, rateHz uint32) []byte {
	const headerLen = 44
	const extraLength = headerLen - 8

	numChannels := uint16(1)
	bytesPerSecond := rateHz
	bytesPerSample := uint16(1)
	bitsPerSample := uint16(bytesPerSample * 8)

	buf := new(bytes.Buffer)

	// RIFF Chunk
	buf.WriteString("RIFF")
	// Write number of bytes to come. Length of data + headerLen - the 8
	// bytes already written.
	writeUint32(buf, uint32(len(data) + extraLength))
		buf.WriteString("WAVE")

	// FORMAT Chunk
	buf.WriteString("fmt ")
	// FORMAT chunk length
	writeUint32(buf, 16)
	// unknown
	writeUint16(buf, 1)
	writeUint16(buf, numChannels)
	writeUint32(buf, rateHz)
	writeUint32(buf, bytesPerSecond)
	writeUint16(buf, bytesPerSample)
	writeUint16(buf, bitsPerSample)

	// DATA Chunk
	buf.WriteString("data")
	writeUint32(buf, uint32(len(data)))

	buf.Write(data)

	return buf.Bytes()
}

// Convert a wave function that maps time in seconds into an amplitude between
// 1 and -1 into a sample of the given rate and duration.
func MakeSound8Bit(wave func(float) float, rateHz uint32, durationSec float) []byte {
	buf := new(bytes.Buffer)

	timeStep := 1.0 / float(rateHz);
	for t := 0.0; t < durationSec; t += timeStep {
		// FIXME: Compiler error workaround. Remove the int cast when 8g is fixed.
//		sample := byte((wave(t) + 1.0) * 128.0)
		sample := byte(int((wave(t) + 1.0) * 128.0))
		buf.WriteByte(sample)
	}
	return buf.Bytes()
}
