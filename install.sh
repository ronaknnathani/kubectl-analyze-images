#!/bin/sh
set -eu

REPO="ronaknnathani/kubectl-analyze-images"
BINARY="kubectl-analyze-images"

detect_os() {
	case "$(uname -s)" in
		Darwin) echo "darwin" ;;
		Linux) echo "linux" ;;
		*)
			echo "Unsupported OS: $(uname -s). Install manually from https://github.com/$REPO/releases" >&2
			exit 1
			;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64 | amd64) echo "amd64" ;;
		arm64 | aarch64) echo "arm64" ;;
		*)
			echo "Unsupported architecture: $(uname -m). Install manually from https://github.com/$REPO/releases" >&2
			exit 1
			;;
	esac
}

latest_tag() {
	curl -fsSLI -o /dev/null -w "%{url_effective}" "https://github.com/$REPO/releases/latest" | sed 's#.*/##'
}

install_dir() {
	if [ "${INSTALL_DIR:-}" ]; then
		echo "$INSTALL_DIR"
	elif [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
		echo "/usr/local/bin"
	else
		echo "$HOME/.local/bin"
	fi
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
TAG="${VERSION:-$(latest_tag)}"
VERSION_NUMBER="${TAG#v}"
ARCHIVE="${BINARY}_${VERSION_NUMBER}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$TAG/$ARCHIVE"
DEST_DIR="$(install_dir)"
TMP_DIR="$(mktemp -d 2>/dev/null || mktemp -d -t kubectl-analyze-images)"

cleanup() {
	rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

mkdir -p "$DEST_DIR"

echo "Downloading $BINARY $TAG for $OS/$ARCH..."
curl -fsSL "$URL" -o "$TMP_DIR/$ARCHIVE"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$BINARY" ]; then
	echo "Release archive did not contain $BINARY" >&2
	exit 1
fi

chmod +x "$TMP_DIR/$BINARY"
cp "$TMP_DIR/$BINARY" "$DEST_DIR/$BINARY"

echo "Installed $BINARY to $DEST_DIR/$BINARY"
if ! command -v "$BINARY" >/dev/null 2>&1; then
	echo "Add $DEST_DIR to your PATH to run $BINARY directly." >&2
fi
