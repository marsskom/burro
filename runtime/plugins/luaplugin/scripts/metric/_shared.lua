shared = shared or {}

--- Loads JSON value from KV storage.
--- Returns:
---   1. decoded table (or empty table on missing key)
---   2. stop flag (true if execution should stop)
---
---@param key string
---@return table|nil value
---@return boolean stop
function shared.load_kv_json(key)
	local data, get_err = kv.get(key)

	if get_err ~= nil or data == nil then
		return {}, false
	end

	local v, decode_err = json.decode(data)
	if decode_err ~= nil then
		log.debug("cannot decode kv data", {
			key = key,
			err = decode_err,
		})

		return nil, true -- stop execution
	end

	return v, false
end

--- Save JSON value to KV storage.
--- Returns:
---   1. stop flag (true if execution should stop)
---
---@param value any
---@param key string
---@return boolean stop
function shared.save_kv_json(value, key)
	local v_str, encode_err = json.encode(value)
	if encode_err ~= nil then
		log.error("cannot encode to json", { value = value, err = encode_err })

		return true
	end

	if v_str == nil then
		log.error("encoded value to json is nil", { value = v_str })

		return true
	end

	local _, set_err = kv.set(key, v_str)
	if set_err ~= nil then
		log.error("cannot save before_request request data", { err = set_err })

		return true
	end

	return false
end
