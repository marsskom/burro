package actions

import (
	"io"
	"strings"
	"testing"

	"gitlab.com/marsskom/burro/internal/testutils"
)

func TestExecute_EmptyRules(t *testing.T) {
	r := req("GET", "/", "", "")

	resp := Execute(testutils.NewMemoryLogger(), nil, r)

	if resp != nil {
		t.Fatal("expected nil response")
	}
}

func TestExecute_Deny(t *testing.T) {
	r := req("GET", "/", "", "")

	rules := []ActionRule{
		{
			Match: Match{Method: "GET"},
			Action: []Action{
				{Operation: OpDeny},
			},
		},
	}

	resp := Execute(testutils.NewMemoryLogger(), rules, r)

	if resp == nil {
		t.Fatal("expected forbidden response")
	}
}

func TestExecute_AllowStopsChain(t *testing.T) {
	r := req("GET", "/", "", "")

	rules := []ActionRule{
		{
			Match: Match{Method: "GET"},
			Action: []Action{
				{Operation: OpAllow},
			},
		},
	}

	resp := Execute(testutils.NewMemoryLogger(), rules, r)

	if resp != nil {
		t.Fatal("expected nil response for allow")
	}
}

func TestExecute_SetHeader(t *testing.T) {
	r := req("GET", "/", "", "")

	rules := []ActionRule{
		{
			Match: Match{Method: "GET"},
			Action: []Action{
				{
					Operation: OpSetHeader,
					Args: SetHeaderArgs{
						Name:  "X-Test",
						Value: "123",
					},
				},
			},
		},
	}

	Execute(testutils.NewMemoryLogger(), rules, r)

	if r.Header.Get("X-Test") != "123" {
		t.Fatal("header not set")
	}
}

func TestExecute_RemoveHeader(t *testing.T) {
	r := req("GET", "/", "", "")
	r.Header.Set("X-Test", "1")

	rules := []ActionRule{
		{
			Match: Match{Method: "GET"},
			Action: []Action{
				{
					Operation: OpRemoveHeader,
					Args: RemoveHeaderArgs{
						Names: []string{"X-Test"},
					},
				},
			},
		},
	}

	Execute(testutils.NewMemoryLogger(), rules, r)

	if r.Header.Get("X-Test") != "" {
		t.Fatal("header not removed")
	}
}

func TestExecute_RedactBody(t *testing.T) {
	body := `{"token":"secret","user":"a"}`

	r := req("POST", "/", "", "")
	r.Body = io.NopCloser(strings.NewReader(body))

	rules := []ActionRule{
		{
			Match: Match{Method: "POST"},
			Action: []Action{
				{
					Operation: OpRedactBody,
					Args: RedactBodyArgs{
						Fields: []string{"token"},
					},
				},
			},
		},
	}

	Execute(testutils.NewMemoryLogger(), rules, r)

	b, _ := io.ReadAll(r.Body)

	if strings.Contains(string(b), "secret") {
		t.Fatal("token not redacted")
	}
}

func TestExecute_LogRouting(t *testing.T) {
	r := req("GET", "/", "", "")

	l := testutils.NewMemoryLogger()

	rules := []ActionRule{
		{
			Match: Match{Method: "GET"},
			Action: []Action{
				{
					Operation: OpLog,
					Args: LogArgs{
						Level:   "audit",
						Message: "hello audit",
					},
				},
			},
		},
	}

	Execute(l, rules, r)

	if len(l.Messages["audit"]) != 1 {
		t.Fatal("expected audit log")
	}

	if !strings.HasPrefix(l.Messages["audit"][0], "hello audit") {
		t.Fatal("wrong audit message")
	}
}

func TestExecute_UnknownOperation(t *testing.T) {
	r := req("GET", "/", "", "")

	rules := []ActionRule{
		{
			Match: Match{Method: "GET"},
			Action: []Action{
				{Operation: "unknown"},
			},
		},
	}

	resp := Execute(testutils.NewMemoryLogger(), rules, r)

	if resp == nil {
		t.Fatal("expected internal error response")
	}
}

func TestExecute_OnMatchStop(t *testing.T) {
	r := req("GET", "/", "", "")

	rules := []ActionRule{
		{
			Match:   Match{Method: "GET"},
			OnMatch: OnMatchStop,
			Action: []Action{
				{Operation: OpAllow},
			},
		},
		{
			Match: Match{Method: "GET"},
			Action: []Action{
				{Operation: OpDeny},
			},
		},
	}

	resp := Execute(testutils.NewMemoryLogger(), rules, r)

	if resp != nil {
		t.Fatal("expected allow to stop chain before deny")
	}
}
