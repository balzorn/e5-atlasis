package asset

import "testing"

func TestLifecycleTransitions(t *testing.T) {
	tests := []struct {
		from AssetStatus
		to   AssetStatus
		want bool
	}{
		{AssetStatusDraft, AssetStatusActive, true},
		{AssetStatusDraft, AssetStatusRetired, true},
		{AssetStatusActive, AssetStatusSuspended, true},
		{AssetStatusActive, AssetStatusRetired, true},
		{AssetStatusSuspended, AssetStatusActive, true},
		{AssetStatusSuspended, AssetStatusRetired, true},
		{AssetStatusRetired, AssetStatusActive, false},
	}

	var a InformationAsset

	for _, tt := range tests {
		a.Status = tt.from

		if got := a.CanTransitionTo(tt.to); got != tt.want {
			t.Errorf(
				"CanTransitionTo(%s -> %s) = %v, want %v",
				tt.from,
				tt.to,
				got,
				tt.want,
			)
		}
	}
}
