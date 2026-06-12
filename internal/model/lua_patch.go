package model

type Patch struct {
	Ctx  *CtxPatch
	Req  *RequestPatch
	Resp *ResponsePatch
}

func NewPatch() *Patch {
	return &Patch{
		Ctx:  &CtxPatch{},
		Req:  &RequestPatch{},
		Resp: &ResponsePatch{},
	}
}

type CtxPatch struct {
	IsFinished *bool
}

type RequestPatch struct {
	Host    *string
	Scheme  *string
	Method  *string
	Path    *string
	URL     *string
	Headers map[string][]string // nil = no change, key with nil value = delete header
	Cookies []CookiePatch
	Body    *[]byte
}

type CookiePatch struct {
	Name  string
	Value string

	Path   string
	Domain string

	Secure   bool
	HTTPOnly bool
	MaxAge   int

	Delete bool
}

type ResponsePatch struct {
	StatusCode *int
	Headers    map[string][]string // nil = no change, key with nil value = delete header
	Body       *[]byte
}
