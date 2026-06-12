---Exists only in `flush.lua`
---@class Opts
---@field session string

---@type Opts
opts = opts or {}

local json_data = {}
local data_keys = {
	"connect",
	"before_request",
	"after_request",
	"before_response",
	"after_response",
	"error",
	"close",
	"ws_open",
	"ws_msg",
	"ws_close",
}

for _, v in ipairs(data_keys) do
	local kv_data, get_err = kv.get(v)
	if get_err ~= nil or kv_data == nil then
		log.warn("wrong data in kv storage", { data = kv_data, err = get_err })
	else
		local v_dec, dec_err = json.decode(kv_data)

		if dec_err ~= nil then
			log.error("cannot decode data from json", { data = kv_data, err = dec_err })
		else
			json_data[v] = v_dec
		end
	end
end

if next(json_data) == nil then
	log.info("metric data is empty, nothing to write into a file", { data = json_data })

	return
end

local js_enc, enc_err = json.encode(json_data)
if enc_err ~= nil or js_enc == nil then
	log.error("cannot encode metric data into json", { data = json_data, err = enc_err })

	return
end

local now = time.date()
local _, write_err = artifacts.write("metric/" .. opts.session .. "-" .. now .. ".json", tostring(js_enc))
if write_err ~= nil then
	log.error("cannot write to file", { data = js_enc, err = write_err })
end
