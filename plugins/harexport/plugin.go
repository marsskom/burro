package harexport

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"gitlab.com/marsskom/burro/internal/export"
	"gitlab.com/marsskom/burro/internal/model"
	"gitlab.com/marsskom/burro/internal/plugin"
)

func init() {
	plugin.Register("harexport", func() plugin.Plugin {
		return New()
	})
}

type HARExporConfig struct {
	Enabled  *bool  `yaml:"enabled"`
	Priority int    `yaml:"priority"`
	File     string `yaml:"file"`
	Override bool   `yaml:"override"`
}

type HARExportPlugin struct {
	enabled  *bool
	priority int

	outputFile string
	override   bool

	entries map[string]*HAREntry

	mu sync.Mutex
}

func New() *HARExportPlugin {
	return &HARExportPlugin{
		entries: make(map[string]*HAREntry),
	}
}

func (p *HARExportPlugin) Enabled() *bool {
	return p.enabled
}

func (p *HARExportPlugin) Priority() int {
	return p.priority
}

func (p *HARExportPlugin) Name() string {
	return "harexport"
}

func (p *HARExportPlugin) Init(cfg any) error {
	slog.Debug("HAR exporter plugin is going to init with config", "config", cfg)

	var config HARExporConfig
	if err := plugin.DecodeYAML(cfg, &config); err != nil {
		return fmt.Errorf("HAR exporter plugin init: cannot read pluigin config: %w", err)
	}

	p.enabled = config.Enabled
	p.priority = config.Priority
	p.outputFile = config.File
	p.override = config.Override

	return nil
}

func (p *HARExportPlugin) Flush(opts *export.FileNameVars) error {
	if strings.TrimSpace(opts.Session) == "" && strings.Contains(p.outputFile, "%session%") {
		return fmt.Errorf("HAR cannot save data into file since provided session is empty but filename expects it: '%s'", p.outputFile)
	}

	filename := strings.ReplaceAll(p.outputFile, "%session%", opts.Session)
	filename = strings.ReplaceAll(filename, "%datetime%", time.Now().Format("2026_12_31_12_00_00"))

	isExists := false
	if _, err := os.Stat(filename); err == nil {
		if !p.override {
			return fmt.Errorf("HAR cannot override existed file: %s", filename)
		}

		isExists = true
	}

	tmp := filename + ".tmp"
	os.Remove(tmp)

	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("HAR cannot create output file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	entries := make([]HAREntry, 0, len(p.entries))
	for _, e := range p.entries {
		entries = append(entries, *e)
	}

	err = enc.Encode(HAR{
		Log: HARLog{
			Version: "1.2",
			Creator: HARCreator{
				Name:    "Burro",
				Version: "1",
			},
			Pages:   make([]HARPage, 0),
			Entries: entries,
		},
	})
	if err != nil {
		return fmt.Errorf("HAR cannot write into file: %w", err)
	}

	if isExists {
		err := os.Remove(filename)
		if err != nil {
			return fmt.Errorf("HAR cannot remove existed file: %s", filename)
		}
	}

	os.Rename(tmp, filename)

	return nil
}

func (p *HARExportPlugin) OnConnect(ctx *model.RequestContext) error {
	return nil
}

func (p *HARExportPlugin) OnRequest(ctx *model.RequestContext) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	headers := make([]HARHeader, 0, len(ctx.RequestSnapshot.Headers))
	for k, v := range ctx.RequestSnapshot.Headers {
		headers = append(headers, HARHeader{
			Name:  k,
			Value: strings.Join(v, ", "),
		})
	}

	queryString := make([]HARQueryParam, 0, len(ctx.RequestSnapshot.QueryParams))
	for k, v := range ctx.RequestSnapshot.QueryParams {
		queryString = append(queryString, HARQueryParam{
			Name:  k,
			Value: strings.Join(v, ", "),
		})
	}

	cookies := make([]HARCookie, 0, len(ctx.RequestSnapshot.Cookies))
	for _, c := range ctx.RequestSnapshot.Cookies {
		cookies = append(cookies, HARCookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  c.Expires.Format("2006-01-02T15:04:05.000Z"),
			HTTPOnly: c.HTTPOnly,
			Secure:   c.Secure,
		})
	}

	postData, err := parseRequestBody(ctx)
	if err != nil {
		return fmt.Errorf("error on parsing request body: %w", err)
	}

	entry := &HAREntry{
		StartedDateTime: ctx.StartTime.Format(time.RFC3339Nano),
		Request: HARRequest{
			Method:      ctx.RequestSnapshot.Method,
			URL:         ctx.RequestSnapshot.URL,
			HTTPVersion: ctx.RequestSnapshot.Proto,
			Headers:     headers,
			QueryString: queryString,
			Cookies:     cookies,
			PostData:    postData,
			HeaderSize: approxHeadersSize(
				ctx.RequestSnapshot.Method,
				ctx.RequestSnapshot.URL,
				ctx.RequestSnapshot.Proto,
				ctx.RequestSnapshot.Headers,
			),
			BodySize: len(ctx.RequestSnapshot.Body),
		},
		Cache: HARCache{},
		Timings: HARTimings{
			Blocked: -1,
			DNS:     -1,
			Connect: -1,
			Send:    0,
			Wait:    0,
			Receive: 0,
			SSL:     -1,
		},
	}

	slog.Debug("On request HAR adds entry for request context", "ctx", ctx.ID, "entry", entry)

	p.entries[ctx.ID] = entry

	return nil
}

func (p *HARExportPlugin) OnResponse(ctx *model.RequestContext) error {
	if ctx.ResponseSnapshot == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.entries[ctx.ID]
	if !ok {
		slog.Warn("On response HAR entry for context doesn't exist", "ctx", ctx.ID)

		return nil
	}

	headers := make([]HARHeader, 0, len(ctx.ResponseSnapshot.Headers))
	for k, v := range ctx.ResponseSnapshot.Headers {
		headers = append(headers, HARHeader{
			Name:  k,
			Value: strings.Join(v, ", "),
		})
	}

	contentType, ok := ctx.ResponseSnapshot.Headers["Content-Type"]
	if !ok {
		contentType = []string{"text/plain"}
	}
	contentTypeString := strings.Join(contentType, "")

	cookies := make([]HARCookie, 0)

	var body string
	var encoding string
	if isContentPlainText(contentTypeString) {
		body = string(ctx.ResponseSnapshot.Body)
		encoding = ""
	} else {
		body = base64.StdEncoding.EncodeToString(ctx.ResponseSnapshot.Body)
		encoding = "base64"
	}

	// TODO: request must contains execution time in itself.
	entry.Time = time.Since(ctx.StartTime).Milliseconds()
	entry.Response = HARResponse{
		Status:      ctx.ResponseSnapshot.StatusCode,
		StatusText:  ctx.ResponseSnapshot.Status,
		HTTPVersion: ctx.ResponseSnapshot.Proto,
		Headers:     headers,
		Cookies:     cookies,
		Content: HARResponseContent{
			Size:     len(body),
			MimeType: contentTypeString,
			Text:     body,
			Encoding: encoding,
		},
	}
	entry.Timings = HARTimings{
		Blocked: -1,
		DNS:     int(ctx.ResponseSnapshot.TimeDNS.Milliseconds()),
		Connect: int(ctx.ResponseSnapshot.TimeConnect.Milliseconds()),
		Send:    0,
		Wait:    int(ctx.ResponseSnapshot.TimeWait.Milliseconds()),
		Receive: 0,
		SSL:     int(ctx.ResponseSnapshot.TimeSSL.Milliseconds()),
	}

	slog.Debug("On response HAR add to entry response data", "ctx", ctx.ID, "response", entry.Response)

	return nil
}

func (p *HARExportPlugin) OnError(ctx *model.RequestContext, err error) error {
	return nil
}

func (p *HARExportPlugin) OnClose(ctx *model.RequestContext) error {
	return nil
}
