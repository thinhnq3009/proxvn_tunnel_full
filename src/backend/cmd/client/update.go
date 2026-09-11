package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultUpdateBaseURL     = "https://raw.githubusercontent.com/thinhnq3009/proxvn_tunnel_full/develop/bin/client"
	defaultUpdateChecksumURL = "https://raw.githubusercontent.com/thinhnq3009/proxvn_tunnel_full/develop/bin/SHA256SUMS-client.txt"
)

func runSelfUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	baseURL := fs.String("url", defaultUpdateBaseURL, "base URL chứa binary client")
	checksumURL := fs.String("checksum-url", defaultUpdateChecksumURL, "URL chứa SHA256SUMS-client.txt")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("không hỗ trợ tham số: %s", strings.Join(fs.Args(), " "))
	}

	binaryName, err := updateBinaryName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("không xác định được executable hiện tại: %w", err)
	}
	target, err := filepath.EvalSymlinks(executable)
	if err != nil {
		target = executable
	}

	base := strings.TrimRight(*baseURL, "/")
	binaryURL := base + "/" + binaryName

	fmt.Printf("[update] Đang tải %s\n", binaryURL)
	data, err := downloadUpdateFile(binaryURL)
	if err != nil {
		return err
	}
	if err := verifyUpdateChecksum(data, binaryName, strings.TrimSpace(*checksumURL)); err != nil {
		return err
	}

	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("không đọc được file hiện tại %s: %w", target, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".proxvn-update-*")
	if err != nil {
		return fmt.Errorf("không tạo được file tạm cạnh %s: %w", target, err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("ghi file cập nhật lỗi: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("đóng file cập nhật lỗi: %w", err)
	}
	if err := os.Chmod(tmpPath, info.Mode().Perm()); err != nil {
		return fmt.Errorf("chmod file cập nhật lỗi: %w", err)
	}
	if err := os.Rename(tmpPath, target); err != nil {
		return fmt.Errorf("thay binary hiện tại lỗi: %w", err)
	}
	cleanup = false

	fmt.Printf("[update] Xong. Đã cập nhật %s\n", target)
	return nil
}

func updateBinaryName(goos, goarch string) (string, error) {
	switch goos + "/" + goarch {
	case "linux/amd64":
		return "proxvn-linux-amd64", nil
	case "linux/arm64":
		return "proxvn-linux-arm64", nil
	case "darwin/amd64":
		return "proxvn-darwin-amd64", nil
	case "darwin/arm64":
		return "proxvn-darwin-arm64", nil
	case "windows/amd64":
		return "proxvn-windows-amd64.exe", nil
	case "android/arm64":
		return "proxvn-android-arm64", nil
	default:
		return "", fmt.Errorf("chưa hỗ trợ cập nhật tự động cho %s/%s", goos, goarch)
	}
}

func downloadUpdateFile(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("tải %s lỗi: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tải %s lỗi: HTTP %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func verifyUpdateChecksum(data []byte, binaryName, checksumURL string) error {
	checksumData, err := downloadUpdateFile(checksumURL)
	if err != nil {
		return err
	}
	want := checksumForBinary(string(checksumData), binaryName)
	if want == "" {
		return fmt.Errorf("không tìm thấy checksum cho %s", binaryName)
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("checksum không khớp cho %s", binaryName)
	}
	return nil
}

func checksumForBinary(checksums, binaryName string) string {
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == binaryName {
			return fields[0]
		}
	}
	return ""
}
