package websocket

import (
	"bytes"
	"testing"
)

func TestReadWSFrame_Text(t *testing.T) {
	payload := []byte("hello world")

	raw := buildFrame(true, 0x1, payload, true)

	frame, err := readWSFrame(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if frame.OpCode != 0x1 {
		t.Fatalf("expected opcode 1, got %d", frame.OpCode)
	}

	if string(frame.Payload) != "hello world" {
		t.Fatalf("expected payload hello world, got %s", frame.Payload)
	}

	if !frame.Fin {
		t.Fatal("expected FIN=true")
	}
}

func TestReadWSFrame_Binary(t *testing.T) {
	payload := []byte{1, 2, 3, 4, 5}

	raw := buildFrame(true, 0x2, payload, true)

	frame, err := readWSFrame(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if frame.OpCode != 0x2 {
		t.Fatalf("expected binary opcode, got %d", frame.OpCode)
	}

	if !bytes.Equal(frame.Payload, payload) {
		t.Fatalf("payload mismatch")
	}
}

func TestReadWSFrame_Extended16Bit(t *testing.T) {
	payload := make([]byte, 300)
	for i := range payload {
		payload[i] = byte(i % 255)
	}

	raw := buildFrame(true, 0x1, payload, true)

	frame, err := readWSFrame(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if len(frame.Payload) != 300 {
		t.Fatalf("expected 300, got %d", len(frame.Payload))
	}
}

func TestReadWSFrame_NoMask(t *testing.T) {
	payload := []byte("server message")

	raw := buildFrame(true, 0x1, payload, false)

	frame, err := readWSFrame(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if string(frame.Payload) != "server message" {
		t.Fatalf("expected server message")
	}
}

func TestReadWSFrame_FragmentedInput(t *testing.T) {
	payload := []byte("fragmented hello")

	raw := buildFrame(true, 0x1, payload, true)

	// split into chunks like TCP would do
	part1 := raw[:5]
	part2 := raw[5:]

	r := bytes.NewReader(append(part1, part2...))

	frame, err := readWSFrame(r)
	if err != nil {
		t.Fatal(err)
	}

	if string(frame.Payload) != "fragmented hello" {
		t.Fatalf("bad reconstruction")
	}
}
