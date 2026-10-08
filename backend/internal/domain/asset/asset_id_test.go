package asset

import "testing"

func TestParseAssetID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "IA00001", false},
		{"valid upper bound", "IA99999", false},
		{"too short", "IA0001", true},
		{"too long", "IA000001", true},
		{"wrong prefix", "AB00001", true},
		{"non numeric", "IA12ABC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAssetID(tt.value)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ParseAssetID(%q) error = %v, wantErr %v",
					tt.value,
					err,
					tt.wantErr,
				)
			}
		})
	}
}
