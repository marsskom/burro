if ctx.req == nil or ctx.resp == nil then
	return
end

local host = ctx.req.host

log.debug("after_response to host", { host = host })

local value, stop = shared.load_kv_json("after_response")
if stop or value == nil then
	return
end

if value[host] == nil then
	value[host] = 0
end

value[host] = tonumber(value[host]) + 1

value.statuses = value.statuses or {}
value.statuses[host] = value.statuses[host] or {}
value.statuses[host][ctx.resp.status] = value.statuses[host][ctx.resp.status] or 0

value.statuses[host][ctx.resp.status] = tonumber(value.statuses[host][ctx.resp.status]) + 1

value.statuses[host]["X-Metric"] = {}

--- Checks the header.
local flag = false
if ctx.req.headers["X-Metric"] then
	for _, v in ipairs(ctx.req.headers["X-Metric"]) do
		if v == "1" then
			table.insert(value.statuses[host]["X-Metric"], "1")
			flag = true

			break
		end
	end
end

if not flag then
	table.insert(value.statuses[host]["X-Metric"], "0")
end

--- Removes the header.
mut.req.del_header("X-Metric")

stop = shared.save_kv_json(value, "after_response")
if stop then
	return
end
