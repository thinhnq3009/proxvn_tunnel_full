package main

import "testing"

func TestNormalizeRequestedSubdomain(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty", input: "  ", want: ""},
		{name: "normalize", input: " Factorio-01 ", want: "factorio-01"},
		{name: "invalid character", input: "factorio.example", wantErr: true},
		{name: "leading hyphen", input: "-factorio", wantErr: true},
		{name: "trailing hyphen", input: "factorio-", wantErr: true},
		{name: "too long", input: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeRequestedSubdomain(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
