package model

type WSDirection string

const (
	WSClientToServer WSDirection = "client_to_server"
	WSServerToCLient WSDirection = "server_to_client"
)

type WSOpCode int

const (
	WSContinuation WSOpCode = 0x0
	WSText         WSOpCode = 0x1
	WSBinary       WSOpCode = 0x2
	WSClose        WSOpCode = 0x8
	WSPing         WSOpCode = 0x9
	WSPong         WSOpCode = 0xA
)

type WSMessage struct {
	Direction WSDirection
	OpCode    WSOpCode
	Data      []byte
	Text      string
	Timestamp int64
}

func ToWSOpCode(op byte) WSOpCode {
	switch op {
	case 0x0:
		return WSContinuation
	case 0x1:
		return WSText
	case 0x2:
		return WSBinary
	case 0x8:
		return WSClose
	case 0x9:
		return WSPing
	case 0xA:
		return WSPong
	default:
		return WSOpCode(op) // unknown/custom
	}
}
