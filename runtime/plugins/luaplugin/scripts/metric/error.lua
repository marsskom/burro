---Exists only in `error.lua`
---@class Err
---@field msg string

---@type Err
err = err or {}

local value, stop = shared.load_kv_json("error")
if stop or value == nil then
	return
end

table.insert(value, err.msg)

stop = shared.save_kv_json(value, "error")
if stop then
	return
end
