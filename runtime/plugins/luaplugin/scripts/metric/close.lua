local value, stop = shared.load_kv_json("close")
if stop or value == nil then
	return
end

if value["closed"] == nil then
	value["closed"] = 0
end

value["closed"] = tonumber(value["closed"]) + 1

stop = shared.save_kv_json(value, "close")
if stop then
	return
end
