package websocket

import (
	"bytes"
	"io"
)

type WSStream struct {
	buf []byte
	pos int
}

func NewWSStream() *WSStream {
	return &WSStream{}
}

func (s *WSStream) Write(p []byte) {
	s.buf = append(s.buf, p...)
}

func (s *WSStream) NextFrame() (*WSFrame, error) {
	if s.pos >= len(s.buf) {
		return nil, io.EOF
	}

	r := bytes.NewReader(s.buf[s.pos:])

	frame, err := readWSFrame(r)
	if err != nil {
		return nil, err
	}

	consumed := len(s.buf[s.pos:]) - r.Len()
	s.pos += consumed

	return frame, nil
}

func (s *WSStream) GetPos() int {
	return s.pos
}

func (s *WSStream) Compact() {
	if s.pos == 0 {
		return
	}

	s.buf = append([]byte(nil), s.buf[s.pos:]...)
	s.pos = 0
}
