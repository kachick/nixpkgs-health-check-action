package update

import (
	"context"
	"testing"
)

func TestCheckClassifications(t *testing.T) {
	tests := []struct {
		name             string
		output           string
		exitCode         int
		hasPackageIgnore bool
		wantStatus       Status
		wantStale        bool
	}{
		{
			name:       "clean pass",
			output:     "\x1b[32m[OK]\x1b[0m foo",
			exitCode:   0,
			wantStatus: StatusOK,
			wantStale:  false,
		},
		{
			name:             "clean pass with unused package ignore is stale",
			output:           "[OK] foo",
			exitCode:         0,
			hasPackageIgnore: true,
			wantStatus:       StatusOK,
			wantStale:        true,
		},
		{
			name:             "warning with ANSI escape",
			output:           "\x1b[33m[WARN]\x1b[0m ignored failure",
			exitCode:         0,
			hasPackageIgnore: true,
			wantStatus:       StatusWarning,
			wantStale:        false,
		},
		{
			name:       "non-zero exit is error",
			output:     "[ERROR] failed to update",
			exitCode:   1,
			wantStatus: StatusError,
			wantStale:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			run := func(_ context.Context, _ string, _ string) (string, int, error) {
				return tc.output, tc.exitCode, nil
			}
			res, err := Check(context.Background(), run, "foo", nil, tc.hasPackageIgnore)
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			if res.Status != tc.wantStatus {
				t.Errorf("Status = %q, want %q", res.Status, tc.wantStatus)
			}
			if res.StaleIgnore != tc.wantStale {
				t.Errorf("StaleIgnore = %v, want %v", res.StaleIgnore, tc.wantStale)
			}
		})
	}
}
