package domain

import "testing"

func TestValidateFindingStatus(t *testing.T) {
	for _, status := range FindingStatuses {
		if err := ValidateFindingStatus(status); err != nil {
			t.Errorf("ValidateFindingStatus(%q) = %v, want nil", status, err)
		}
	}

	for _, bad := range []string{"", "adressed", "resol ved", "Open", "closed"} {
		if err := ValidateFindingStatus(bad); err == nil {
			t.Errorf("ValidateFindingStatus(%q) = nil, want error", bad)
		}
	}
}
