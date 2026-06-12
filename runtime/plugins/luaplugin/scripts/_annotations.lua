---@meta

---@class Log
---@field trace  fun(msg: string, attrs?: table)
---@field debug  fun(msg: string, attrs?: table)
---@field info   fun(msg: string, attrs?: table)
---@field warn   fun(msg: string, attrs?: table)
---@field error  fun(msg: string, attrs?: table)
---@field audit  fun(msg: string, attrs?: table)

---@type Log
log = {}

---@class KV
---@field get fun(key: string): (string?, string?) @ returns value, error
---@field get_base64 fun(key: string): (string?, string?) @ returns value, error
---@field set fun(key: string, value: string): (boolean, string?) @ returns ok, error
---@field delete fun(key: string): (boolean, string?) @ returns ok, error
---@field list fun(prefix: string): (table<string, string>?, string?) @returns result, error

---@type KV
kv = {}

---@class Bus
---@field emit fun(name: string, data: table?): string? @returns error

---@type Bus
bus = {}

---@class Artifacts
---@field write fun(name: string, content: string): (boolean, string?) @ returns ok, error
---@field read fun(name: string): (string?, string?) @ returns content, error
---@field exists fun(name: string): boolean
---@field delete fun(name: string): (boolean, string?) @ returns ok, error
---@field rename fun(oldpath: string, newpath: string): (boolean, string?) @ returns ok, error
---@field list fun(): (string[]?, string?) @ returns result, error

---@type Artifacts
artifacts = {}

---@class DataStore
---@field exists fun(name: string): boolean
---@field read fun(name: string): (string?, string?) @ returns content, error
---@field list fun(path: string, exts?: string[]): (string[]?, string?) @ returns result, error

---@type DataStore
data_store = {}

---@class Time
---@field unix fun(): integer @ returns integer Unix timestamp in seconds
---@field rfc3339 fun(): string @ returns RFC3339 timestamp
---@field date fun(format?: string): string @ returns Formatted date/time, supports: %Y %y %m %d %H %M %S %F %T

---@type Time
time = {}

---@class Base64
---@field encode fun(data: string): string
---@field decode fun(data: string): (string?, string?) @ returns result, error

---@type Base
base64 = {}

---@class JSON
---@field encode fun(data: any): (string?, string?) @ returns result, error
---@field decode fun(data: string): (any?, string?) @ returns result, error

---@type JSON
json = {}

---@class Cookie
---@field name string
---@field value string
---@field quoted boolean
---@field path string
---@field domain string
---@field expires integer
---@field max_age integer
---@field secure boolean
---@field http_only boolean
---@field same_site integer
---@field partitioned boolean

---@class Request
---@field proto string
---@field host string
---@field scheme string
---@field method string
---@field path string
---@field query table<string, string[]>?
---@field url string
---@field remote_addr string
---@field headers table<string, string[]>?
---@field cookies Cookie[]?
---@field body string?
---@field content_length integer

---@class Response
---@field status string
---@field status_code integer
---@field proto string
---@field headers table<string, string[]>?
---@field body string?
---@field content_length integer

---@class ContextInfo
---@field id string
---@field session_id string
---@field is_finished boolean
---@field req Request?
---@field resp Response?

---@type ContextInfo
ctx = {}

---@class Mut
---@field ctx MutCtx
---@field req MutReq
---@field resp MutResp

---@class MutCtx
---@field set_finish fun(): nil @ marks request as finished

---@class MutReq
---@field set_host fun(host: string): nil
---@field set_scheme fun(scheme: string): nil
---@field set_method fun(method: string): nil
---@field set_path fun(path: string): nil
---@field set_url fun(url: string): nil
---@field set_body fun(body: string): nil
---@field set_header fun(key: string, value: string): nil
---@field add_header fun(key: string, value: string): nil
---@field del_header fun(key: string): nil
---@field set_cookie fun(cookie: { name: string, value: string, path?: string, domain?: string, secure?: boolean, http_only?: boolean, max_age?: integer }): nil
---@field del_cookie fun(name: string): nil

---@class MutResp
---@field set_status fun(status: number): nil
---@field set_body fun(body: string): nil
---@field set_header fun(key: string, value: string): nil
---@field add_header fun(key: string, value: string): nil
---@field del_header fun(key: string): nil

---@type Mut
mut = mut
