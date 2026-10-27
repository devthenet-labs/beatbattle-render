package synth

import "encoding/binary"

// HeaderSize is the size of the RIFF header EncodeWAV writes.
const HeaderSize = 44

// EncodeWAV wraps mono 16-bit PCM samples at SampleRate in a standard
// 44-byte RIFF/WAVE header.
func EncodeWAV(samples []int16) []byte {
	dataLen := len(samples) * 2
	b := make([]byte, HeaderSize+dataLen)
	le := binary.LittleEndian
	copy(b[0:], "RIFF")
	le.PutUint32(b[4:], uint32(36+dataLen))
	copy(b[8:], "WAVE")
	copy(b[12:], "fmt ")
	le.PutUint32(b[16:], 16)           // fmt chunk size
	le.PutUint16(b[20:], 1)            // PCM
	le.PutUint16(b[22:], 1)            // mono
	le.PutUint32(b[24:], SampleRate)   // sample rate
	le.PutUint32(b[28:], SampleRate*2) // byte rate
	le.PutUint16(b[32:], 2)            // block align
	le.PutUint16(b[34:], 16)           // bits per sample
	copy(b[36:], "data")
	le.PutUint32(b[40:], uint32(dataLen))
	for i, s := range samples {
		le.PutUint16(b[HeaderSize+2*i:], uint16(s))
	}
	return b
}
