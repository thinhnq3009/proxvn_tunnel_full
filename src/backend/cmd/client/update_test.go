package main

import "testing"

func TestUpdateBinaryName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{"linux", "amd64", "proxvn-linux-amd64"},
		{"linux", "arm64", "proxvn-linux-arm64"},
		{"darwin", "amd64", "proxvn-darwin-amd64"},
		{"darwin", "arm64", "proxvn-darwin-arm64"},
		{"windows", "amd64", "proxvn-windows-amd64.exe"},
		{"android", "arm64", "proxvn-android-arm64"},
	}
	for _, tt := range tests {
		got, err := updateBinaryName(tt.goos, tt.goarch)
		if err != nil {
			t.Fatalf("%s/%s returned error: %v", tt.goos, tt.goarch, err)
		}
		if got != tt.want {
			t.Fatalf("%s/%s = %q, want %q", tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestUpdateBinaryNameUnsupported(t *testing.T) {
	if _, err := updateBinaryName("plan9", "amd64"); err == nil {
		t.Fatal("unsupported platform returned nil error")
	}
}

func TestChecksumForBinary(t *testing.T) {
	checksums := "abc123  proxvn-linux-amd64\nffff00  proxvn-darwin-arm64\n"
	if got := checksumForBinary(checksums, "proxvn-darwin-arm64"); got != "ffff00" {
		t.Fatalf("checksum = %q, want ffff00", got)
	}
	if got := checksumForBinary(checksums, "missing"); got != "" {
		t.Fatalf("missing checksum = %q, want empty", got)
	}
}
