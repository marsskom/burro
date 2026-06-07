package websocket

import (
	"bytes"
	"encoding/binary"
	"io"
)

type WSFrame struct {
	Fin     bool
	OpCode  byte
	Masked  bool
	Mask    [4]byte
	Length  int64
	Payload []byte
}

// https://www.rfc-editor.org/info/rfc6455/#section-5.2
func readWSFrame(r io.Reader) (*WSFrame, error) {
	var header [2]byte

	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}

	fin := header[0]&0x80 != 0
	opcode := header[0] & 0x0F

	masked := header[1]&0x80 != 0
	length := int64(header[1] & 0x7F)

	// Extended length.
	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return nil, err
		}
		length = int64(binary.BigEndian.Uint16(ext[:]))

	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return nil, err
		}
		length = int64(binary.BigEndian.Uint64(ext[:]))
	}

	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(r, mask[:]); err != nil {
			return nil, err
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	// Unmasks (client to server).
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}

	return &WSFrame{
		Fin:     fin,
		OpCode:  opcode,
		Masked:  masked,
		Mask:    mask,
		Length:  length,
		Payload: payload,
	}, nil
}

func buildFrame(fin bool, opcode byte, payload []byte, masked bool) []byte {
	var buf bytes.Buffer

	b0 := opcode
	if fin {
		b0 |= 0x80
	}
	buf.WriteByte(b0)

	length := len(payload)

	switch {
	case length < 126:
		if masked {
			buf.WriteByte(byte(length) | 0x80)
		} else {
			buf.WriteByte(byte(length))
		}
	case length < 65536:
		if masked {
			buf.WriteByte(126 | 0x80)
		} else {
			buf.WriteByte(126)
		}
		binary.Write(&buf, binary.BigEndian, uint16(length))
	default:
		if masked {
			buf.WriteByte(127 | 0x80)
		} else {
			buf.WriteByte(127)
		}
		binary.Write(&buf, binary.BigEndian, uint64(length))
	}

	var mask [4]byte
	if masked {
		mask = [4]byte{1, 2, 3, 4}
		buf.Write(mask[:])

		maskedPayload := make([]byte, len(payload))
		for i := range payload {
			maskedPayload[i] = payload[i] ^ mask[i%4]
		}
		buf.Write(maskedPayload)
	} else {
		buf.Write(payload)
	}

	return buf.Bytes()
}
