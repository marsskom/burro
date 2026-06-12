local value, stop = shared.load_kv_json("ws_open")
if stop or value == nil then
	return
end

if value["attempts"] == nil then
	value["attempts"] = 0
end

value["attempts"] = tonumber(value["attempts"]) + 1

stop = shared.save_kv_json(value, "ws_open")
if stop then
	return
end
