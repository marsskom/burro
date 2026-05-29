package harexport

type HAR struct {
	Log HARLog `json:"log"`
}

type HARLog struct {
	Version string     `json:"version"`
	Comment string     `json:"comment,omitempty"`
	Creator HARCreator `json:"creator"`
	Pages   []HARPage  `json:"pages"`
	Entries []HAREntry `json:"entries"`
}

type HARCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

type HARPage struct {
}

type HAREntry struct {
	StartedDateTime string      `json:"startedDateTime"`
	Time            int64       `json:"time"`
	Request         HARRequest  `json:"request"`
	Response        HARResponse `json:"response"`
	Cache           HARCache    `json:"cache"`
	Timings         HARTimings  `json:"timings"`
}

type HARCache struct{}

type HARTimings struct {
	Blocked int `json:"blocked"`
	DNS     int `json:"dns"`
	Connect int `json:"connect"`
	Send    int `json:"send"`
	Wait    int `json:"wait"`
	Receive int `json:"receive"`
	SSL     int `json:"ssl"`
}

type HARRequest struct {
	Method      string          `json:"method"`
	URL         string          `json:"url"`
	HTTPVersion string          `json:"httpVersion"`
	Headers     []HARHeader     `json:"headers"`
	QueryString []HARQueryParam `json:"queryString"`
	Cookies     []HARCookie     `json:"cookies"`
	PostData    *HARPostData    `json:"postData,omitempty"`
	HeaderSize  int             `json:"headerSize"`
	BodySize    int             `json:"bodySize"`
}

type HARQueryParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARPostData struct {
	MimeType string     `json:"mimeType"`
	Params   []HARParam `json:"params,omitempty"`
	Text     string     `json:"text,omitempty"`
}

type HARParam struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

type HARResponse struct {
	Status      int                `json:"status"`
	StatusText  string             `json:"statusText"`
	HTTPVersion string             `json:"httpVersion"`
	Headers     []HARHeader        `json:"headers"`
	Cookies     []HARCookie        `json:"cookies"`
	Content     HARResponseContent `json:"content"`
}

type HARHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	Expires  string `json:"expires"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type HARResponseContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
	Encoding string `json:"encoding"`
}
