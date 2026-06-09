package actions

import (
	"strings"
	"testing"
)

func validRule() ActionRule {
	return ActionRule{
		ID:       "rule-1",
		Priority: 10,
		OnMatch:  OnMatchContinue,
		Match: Match{
			Method: "GET",
		},
		Action: []Action{
			{
				Operation: OpAllow,
			},
		},
	}
}

func TestActionRuleValidate(t *testing.T) {

	tests := []struct {
		name      string
		modify    func(*ActionRule)
		wantError bool
		errPart   string
	}{
		{
			name:      "valid rule",
			modify:    func(r *ActionRule) {},
			wantError: false,
		},
		{
			name: "missing ID",
			modify: func(r *ActionRule) {
				r.ID = ""
			},
			wantError: true,
			errPart:   "ID required",
		},
		{
			name: "negative priority",
			modify: func(r *ActionRule) {
				r.Priority = -1
			},
			wantError: true,
			errPart:   ">= 0",
		},
		{
			name: "no actions",
			modify: func(r *ActionRule) {
				r.Action = nil
			},
			wantError: true,
			errPart:   "at least one action",
		},
		{
			name: "invalid on_match",
			modify: func(r *ActionRule) {
				r.OnMatch = "invalid"
			},
			wantError: true,
			errPart:   "invalid value",
		},
		{
			name: "match invalid combination",
			modify: func(r *ActionRule) {
				r.Match.Method = "GET"
				r.Match.All = []Match{{Method: "POST"}}
			},
			wantError: true,
			errPart:   "cannot mix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := validRule()
			tt.modify(&r)

			err := r.Validate()

			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if !strings.Contains(err.Error(), tt.errPart) {
					t.Fatalf("expected error to contain %q, got %v", tt.errPart, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestActionValidate(t *testing.T) {

	tests := []struct {
		name      string
		action    Action
		wantError bool
	}{
		{
			name: "valid deny",
			action: Action{
				Operation: OpDeny,
			},
		},
		{
			name: "deny with args invalid",
			action: Action{
				Operation: OpDeny,
				Args: map[string]any{
					"foo": "bar",
				},
			},
			wantError: true,
		},
		{
			name: "set header valid",
			action: Action{
				Operation: OpSetHeader,
				Args: SetHeaderArgs{
					Name:  "X-Test",
					Value: "1",
				},
			},
		},
		{
			name: "set header missing name",
			action: Action{
				Operation: OpSetHeader,
				Args: SetHeaderArgs{
					Value: "1",
				},
			},
			wantError: true,
		},
		{
			name: "remove header empty",
			action: Action{
				Operation: OpRemoveHeader,
				Args: RemoveHeaderArgs{
					Names: nil,
				},
			},
			wantError: true,
		},
		{
			name: "redact body empty fields",
			action: Action{
				Operation: OpRedactBody,
				Args:      RedactBodyArgs{},
			},
			wantError: true,
		},
		{
			name: "log valid",
			action: Action{
				Operation: OpLog,
				Args: LogArgs{
					Level:   "info",
					Message: "test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.action.validateAction()

			if tt.wantError && err == nil {
				t.Fatalf("expected error")
			}

			if !tt.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMatchValidate(t *testing.T) {

	tests := []struct {
		name      string
		m         Match
		wantError bool
	}{
		{
			name: "valid leaf",
			m: Match{
				Method: "GET",
			},
		},
		{
			name: "valid all",
			m: Match{
				All: []Match{{Method: "GET"}},
			},
		},
		{
			name: "mix all and any",
			m: Match{
				All: []Match{{Method: "GET"}},
				Any: []Match{{Method: "POST"}},
			},
			wantError: true,
		},
		{
			name: "mix logic and leaf",
			m: Match{
				All:    []Match{{Method: "GET"}},
				Domain: "example.com",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.m.validateMatch()

			if tt.wantError && err == nil {
				t.Fatalf("expected error")
			}

			if !tt.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
