#!/bin/sh

set -eu

repo="thinhnq3009/proxvn_tunnel_full"
ref="${PROXVN_REF:-v7.5.0}"
install_dir="${PROXVN_INSTALL_DIR:-/usr/local/bin}"

if [ "$(uname -s)" != "Linux" ]; then
    echo "Lỗi: installer này chỉ hỗ trợ Linux/Ubuntu." >&2
    exit 1
fi

case "$(uname -m)" in
    x86_64|amd64)
        arch="amd64"
        ;;
    aarch64|arm64)
        arch="arm64"
        ;;
    *)
        echo "Lỗi: kiến trúc $(uname -m) chưa được hỗ trợ (chỉ amd64 và arm64)." >&2
        exit 1
        ;;
esac

for command_name in curl sha256sum install mktemp awk; do
    if ! command -v "$command_name" >/dev/null 2>&1; then
        echo "Lỗi: thiếu lệnh '$command_name'." >&2
        exit 1
    fi
done

if [ ! -d "$install_dir" ]; then
    mkdir -p "$install_dir"
fi

if [ ! -w "$install_dir" ]; then
    echo "Lỗi: không có quyền ghi vào $install_dir. Hãy chạy installer bằng sudo." >&2
    exit 1
fi

binary_name="proxvn-linux-$arch"
base_url="https://raw.githubusercontent.com/$repo/$ref/bin"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/proxvn-install.XXXXXX")"

cleanup() {
    rm -rf "$tmp_dir"
}
trap cleanup EXIT HUP INT TERM

echo "Đang tải ProxVN cho Linux $arch..."
curl -fL --retry 3 --connect-timeout 15 \
    "$base_url/client/$binary_name" \
    -o "$tmp_dir/proxvn"
curl -fL --retry 3 --connect-timeout 15 \
    "$base_url/SHA256SUMS-client.txt" \
    -o "$tmp_dir/SHA256SUMS-client.txt"

expected_hash="$(awk -v name="$binary_name" '$2 == name { print $1 }' "$tmp_dir/SHA256SUMS-client.txt")"
if [ -z "$expected_hash" ]; then
    echo "Lỗi: không tìm thấy checksum cho $binary_name." >&2
    exit 1
fi

printf '%s  %s\n' "$expected_hash" "$tmp_dir/proxvn" | sha256sum -c -
install -m 0755 "$tmp_dir/proxvn" "$install_dir/proxvn"

echo "Đã cài ProxVN vào $install_dir/proxvn"
echo "Chạy thử: proxvn --help"
