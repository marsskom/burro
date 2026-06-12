package mapper

import (
	lua "github.com/yuin/gopher-lua"
	"gitlab.com/marsskom/burro/internal/model"
)

func WSMsgToLua(L *lua.LState, msg *model.WSMessage) (*lua.LTable, error) {
	tbl := L.NewTable()

	L.SetField(tbl, "direction", lua.LString(msg.Direction))
	L.SetField(tbl, "opcode", lua.LNumber(msg.OpCode))
	L.SetField(tbl, "opcode_name", lua.LString(opcodeName(msg.OpCode)))
	L.SetField(tbl, "timestamp", lua.LNumber(msg.Timestamp))

	if len(msg.Data) == 0 {
		L.SetField(tbl, "data", lua.LNil)
	} else {
		L.SetField(tbl, "data", lua.LString(msg.Data))
	}

	if msg.Text == "" {
		L.SetField(tbl, "text", lua.LNil)
	} else {
		L.SetField(tbl, "text", lua.LString(msg.Text))
	}

	return tbl, nil
}

func opcodeName(op model.WSOpCode) string {
	switch op {
	case model.WSText:
		return "text"
	case model.WSBinary:
		return "binary"
	case model.WSClose:
		return "close"
	case model.WSPing:
		return "ping"
	case model.WSPong:
		return "pong"
	case model.WSContinuation:
		return "continuation"
	default:
		return "unknown"
	}
}
