if ctx.req == nil then
	return
end

local host = ctx.req.host

log.debug("before_request to host", { host = host })

local value, stop = shared.load_kv_json("before_request")
if stop or value == nil then
	return
end

if value[host] == nil then
	value[host] = 0
end

value[host] = tonumber(value[host]) + 1

stop = shared.save_kv_json(value, "before_request")
if stop then
	return
end

-- Adds a header.
mut.req.add_header("X-Metric", "1")
mut.req.set_cookie({ name = "X-Metric", value = "1" })
