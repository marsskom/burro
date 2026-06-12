if ctx.req == nil then
	return
end

local host = ctx.req.host

log.debug("before_response to host", { host = host })

local value, stop = shared.load_kv_json("before_response")
if stop or value == nil then
	return
end

if value[host] == nil then
	value[host] = 0
end

value[host] = tonumber(value[host]) + 1

stop = shared.save_kv_json(value, "before_response")
if stop then
	return
end
