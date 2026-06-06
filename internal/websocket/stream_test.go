package websocket

import (
	"io"
	"testing"
)

func TestWSStream_BasicFlow(t *testing.T) {
	stream := NewWSStream()

	raw1 := buildFrame(true, 0x1, []byte("hello"), true)
	raw2 := buildFrame(true, 0x1, []byte("world"), true)

	stream.Write(raw1)
	stream.Write(raw2)

	f1, err := stream.NextFrame()
	if err != nil {
		t.Fatal(err)
	}

	f2, err := stream.NextFrame()
	if err != nil {
		t.Fatal(err)
	}

	if string(f1.Payload) != "hello" {
		t.Fatalf("expected hello, got %s", f1.Payload)
	}

	if string(f2.Payload) != "world" {
		t.Fatalf("expected world, got %s", f2.Payload)
	}
}

func TestWSStream_PartialFrame(t *testing.T) {
	stream := NewWSStream()

	raw := buildFrame(true, 0x1, []byte("partial-frame"), true)

	// Simulates TCP split.
	stream.Write(raw[:5])

	_, err := stream.NextFrame()
	if err == nil {
		t.Fatal("expected error for incomplete frame")
	}

	stream.Write(raw[5:])

	frame, err := stream.NextFrame()
	if err != nil {
		t.Fatal(err)
	}

	if string(frame.Payload) != "partial-frame" {
		t.Fatalf("bad payload")
	}
}

func TestWSStream_MultipleFramesInOneChunk(t *testing.T) {
	stream := NewWSStream()

	raw := append(
		buildFrame(true, 0x1, []byte("a"), true),
		buildFrame(true, 0x1, []byte("b"), true)...,
	)

	stream.Write(raw)

	f1, _ := stream.NextFrame()
	f2, _ := stream.NextFrame()

	if string(f1.Payload) != "a" || string(f2.Payload) != "b" {
		t.Fatal("frame splitting failed")
	}
}

func TestWSStream_Compact(t *testing.T) {
	stream := NewWSStream()

	raw1 := buildFrame(true, 0x1, []byte("first"), true)
	raw2 := buildFrame(true, 0x1, []byte("second"), true)

	stream.Write(raw1)
	stream.Write(raw2)

	_, _ = stream.NextFrame()
	_, _ = stream.NextFrame()

	// Forces compact.
	stream.Compact()

	if stream.GetPos() != 0 {
		t.Fatal("expected pos reset after compact")
	}

	if len(stream.buf) != 0 {
		t.Fatal("buffer should be empty")
	}
}

func TestWSStream_CompactWithLefovers(t *testing.T) {
	stream := NewWSStream()

	raw1 := buildFrame(true, 0x1, []byte("first"), true)
	raw2 := buildFrame(true, 0x1, []byte("second"), true)
	raw3 := buildFrame(true, 0x1, []byte("third"), true)

	stream.Write(raw1)
	stream.Write(raw2)
	stream.Write(raw3)

	_, _ = stream.NextFrame()
	_, _ = stream.NextFrame()

	// Forces compact.
	stream.Compact()

	if stream.GetPos() != 0 {
		t.Fatal("expected pos reset after compact")
	}

	if len(stream.buf) == 0 {
		t.Fatal("buffer should contains lefovers")
	}
}

func TestWSStream_EmptyBehavior(t *testing.T) {
	stream := NewWSStream()

	_, err := stream.NextFrame()
	if err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}
