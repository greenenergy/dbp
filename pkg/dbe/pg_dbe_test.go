package dbe

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/lib/pq"
)

func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		undefined bool
		retryable bool
	}{
		{name: "missing table", err: &pq.Error{Code: "42P01"}, undefined: true},
		{name: "starting up", err: &pq.Error{Code: "57P03"}, retryable: true},
		{name: "bad password", err: &pq.Error{Code: "28P01"}},
		{name: "network", err: &net.OpError{Op: "dial", Err: errors.New("refused")}, retryable: true},
		{name: "wrapped network", err: fmt.Errorf("x: %w", &net.OpError{Op: "dial", Err: errors.New("refused")}), retryable: true},
		{name: "other", err: errors.New("boom")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUndefinedTable(tt.err); got != tt.undefined {
				t.Errorf("isUndefinedTable = %v, want %v", got, tt.undefined)
			}
			if got := isRetryableConnError(tt.err); got != tt.retryable {
				t.Errorf("isRetryableConnError = %v, want %v", got, tt.retryable)
			}
		})
	}
}

func TestToConnStrQuoting(t *testing.T) {
	tests := []struct {
		password string
		want     string
	}{
		{"plain", "password=plain"},
		{"with space", `password='with space'`},
		{`it's`, `password='it\'s'`},
		{`back\slash`, `password='back\\slash'`},
		{"a=b@c", "password=a=b@c"},
	}

	for _, tt := range tests {
		ea := &EngineArgs{Host: "h", Password: tt.password}
		if got := ea.ToConnStr(false); !strings.Contains(got, tt.want) {
			t.Errorf("ToConnStr(%q) = %q, want it to contain %q", tt.password, got, tt.want)
		}
	}

	ea := &EngineArgs{Host: "h", Password: "with space"}
	if got := ea.ToConnStr(true); strings.Contains(got, "with space") {
		t.Errorf("protected conn string leaks the password: %q", got)
	}
}
