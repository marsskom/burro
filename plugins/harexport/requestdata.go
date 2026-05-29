package harexport

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"gitlab.com/marsskom/burro/internal/model"
)

func parseRequestBody(ctx *model.RequestContext) (*HARPostData, error) {
	if len(ctx.RequestSnapshot.Body) == 0 {
		return nil, nil
	}

	validMethods := []string{
		string(http.MethodPost),
		string(http.MethodPut),
		string(http.MethodPatch),
	}
	if !slices.Contains(validMethods, ctx.RequestSnapshot.Method) {
		return nil, nil
	}

	contentType, ok := ctx.RequestSnapshot.Headers["Content-Type"]
	if !ok {
		return nil, nil
	}

	contentTypeString := strings.Join(contentType, "")
	if strings.HasPrefix(contentTypeString, "application/json") {
		return &HARPostData{
			MimeType: contentTypeString,
			Text:     string(ctx.RequestSnapshot.Body),
		}, nil
	}

	if strings.HasPrefix(contentTypeString, "application/x-www-form-urlencoded") {
		vals, err := url.ParseQuery(string(ctx.RequestSnapshot.Body))
		if err != nil {
			return nil, fmt.Errorf("error on parse x-www-form-urlencoded request body: %w", err)
		}

		var params []HARParam
		for k, v := range vals {
			for _, vv := range v {
				params = append(params, HARParam{
					Name:  k,
					Value: vv,
				})
			}
		}

		return &HARPostData{
			MimeType: contentTypeString,
			Params:   params,
		}, nil
	}

	if !strings.HasPrefix(contentTypeString, "multipart/form-data") {
		return &HARPostData{
			MimeType: contentTypeString,
			Text:     string(ctx.RequestSnapshot.Body),
		}, nil
	}

	// multipart/form-data
	_, params, err := mime.ParseMediaType(contentTypeString)
	if err != nil {
		return nil, fmt.Errorf("error parse mime of content type of the request: %w", err)
	}

	boundary := params["boundary"]
	mr := multipart.NewReader(bytes.NewReader(ctx.RequestSnapshot.Body), boundary)

	postData := &HARPostData{
		MimeType: contentTypeString,
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error get the next part of multipart request body: %w", err)
		}

		b, _ := io.ReadAll(part)
		if part.FileName() != "" {
			postData.Params = append(postData.Params, HARParam{
				Name:        part.FormName(),
				FileName:    part.FileName(),
				ContentType: part.Header.Get("Content-Type"),
			})
		} else {
			postData.Params = append(postData.Params, HARParam{
				Name:  part.FormName(),
				Value: string(b),
			})
		}
	}

	return postData, nil
}

func isContentPlainText(contentType string) bool {
	return strings.HasPrefix(contentType, "text/") ||
		strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "javascript") ||
		strings.Contains(contentType, "svg") ||
		strings.Contains(contentType, "x-www-form-urlencoded")
}

// TODO: it looks like we should made a custom connection or catch many information at some point of a http connection process, maybe it is good idea make custom http client wrapper on top of existed to get all information we need from the actual source.
func approxHeadersSize(
	method string,
	url string,
	proto string,
	headers map[string][]string,
) int {
	var b strings.Builder

	// Request line.
	b.WriteString(method)
	b.WriteString(" ")
	b.WriteString(url)
	b.WriteString(" ")
	b.WriteString(proto)
	b.WriteString("\r\n")

	// Headers.
	for k, v := range headers {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(strings.Join(v, ", "))
		b.WriteString("\r\n")
	}

	// End of headers.
	b.WriteString("\r\n")

	return len(b.String())
}
