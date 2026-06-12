---Exists only in `ws_msg.lua`
---@alias WSDirection
---| "client_to_server"
---| "server_to_client"

---@alias WSOpCode
---| 0   -- continuation
---| 1   -- text
---| 2   -- binary
---| 8   -- close
---| 9   -- ping
---| 10  -- pong

---@class WSMessage
---@field direction WSDirection
---@field opcode WSOpCode
---@field opcode_name string
---@field data string?
---@field text string?
---@field timestamp integer

---@type WSMessage
ws = ws or {}

local value, stop = shared.load_kv_json("ws_msg")
if stop or value == nil then
	return
end

if value[ws.opcode_name] == nil then
	value[ws.opcode_name] = 0
end

value[ws.opcode_name] = tonumber(value[ws.opcode_name]) + 1

stop = shared.save_kv_json(value, "ws_msg")
if stop then
	return
end
