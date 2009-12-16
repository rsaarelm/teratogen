// Sound effects package.

package sfx

import (
	"bytes"
	"io"
)

// Little-endian byte output
func writeUint32(out io.Writer, val uint32) {
	out.Write([]byte{
		byte(val % 0xff),
		byte((val >> 8) % 0xff),
		byte((val >> 16) % 0xff),
		byte((val >> 24) % 0xff)})
}

func writeUint16(out io.Writer, val uint16) {
	out.Write([]byte{
		byte(val % 0xff),
		byte((val >> 8) % 0xff)})
}

func MakeMono8Wav(data []byte, rateHz uint32) []byte {
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
	writeUint32(buf, 10)
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
