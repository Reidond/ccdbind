#!/bin/sh
# install.sh - Install ccdbind and ccdpin
# SPDX-License-Identifier: MIT

set -e

PROG_NAME="${0##*/}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
CONFIGDIR="${CONFIGDIR:-$HOME/.config/ccdbind}"
SYSTEMD_USER_DIR="${SYSTEMD_USER_DIR:-$HOME/.config/systemd/user}"

# Build flags
GO="${GO:-go}"
GOFLAGS="${GOFLAGS:-}"

# Colors (disabled if not a tty)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' NC=''
fi

info()  { printf "${BLUE}==>${NC} %s\n" "$*"; }
ok()    { printf "${GREEN}==>${NC} %s\n" "$*"; }
warn()  { printf "${YELLOW}==> WARNING:${NC} %s\n" "$*" >&2; }
die()   { printf "${RED}==> ERROR:${NC} %s\n" "$*" >&2; exit 1; }

usage() {
    cat <<EOF
Usage: $PROG_NAME [OPTIONS]

Install ccdbind and ccdpin binaries and systemd user units.

Options:
    -h, --help          Show this help message
    -n, --dry-run       Print actions without executing
    -s, --skip-build    Skip building binaries (use existing)
    -S, --skip-service  Skip systemd service setup
    --prefix=PATH       Install prefix (default: ~/.local)
    --bindir=PATH       Binary directory (default: PREFIX/bin)
    --configdir=PATH    Config directory (default: ~/.config/ccdbind)

Environment:
    PREFIX              Install prefix
    BINDIR              Binary directory
    CONFIGDIR           Config directory
    GO                  Go compiler (default: go)
    GOFLAGS             Additional flags for go build

Examples:
    $PROG_NAME
    $PROG_NAME --prefix=/usr/local
    $PROG_NAME --dry-run
EOF
    exit 0
}

# Parse a --key=value argument
parse_arg() {
    printf '%s' "${1#*=}"
}

# Run a command, or print it if dry-run
run() {
    if [ "$DRY_RUN" = 1 ]; then
        printf '%s\n' "$*"
    else
        "$@"
    fi
}

# Create directory if it doesn't exist
ensure_dir() {
    if [ ! -d "$1" ]; then
        run mkdir -p "$1"
    fi
}

# Install a file with mode
install_file() {
    _mode="$1"
    _src="$2"
    _dst="$3"
    ensure_dir "$(dirname "$_dst")"
    run install -m "$_mode" "$_src" "$_dst"
}

# Check for required commands
check_deps() {
    for cmd in "$@"; do
        command -v "$cmd" >/dev/null 2>&1 || die "Required command not found: $cmd"
    done
}

DRY_RUN=0
SKIP_BUILD=0
SKIP_SERVICE=0

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)      usage ;;
        -n|--dry-run)   DRY_RUN=1 ;;
        -s|--skip-build)   SKIP_BUILD=1 ;;
        -S|--skip-service) SKIP_SERVICE=1 ;;
        --prefix=*)     PREFIX="$(parse_arg "$1")"; BINDIR="$PREFIX/bin" ;;
        --bindir=*)     BINDIR="$(parse_arg "$1")" ;;
        --configdir=*)  CONFIGDIR="$(parse_arg "$1")" ;;
        -*)             die "Unknown option: $1" ;;
        *)              die "Unexpected argument: $1" ;;
    esac
    shift
done

# Determine script directory (where source files are)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

info "Installing ccdbind and ccdpin"
info "  BINDIR:         $BINDIR"
info "  CONFIGDIR:      $CONFIGDIR"
info "  SYSTEMD_USER:   $SYSTEMD_USER_DIR"

# Build
if [ "$SKIP_BUILD" = 0 ]; then
    check_deps "$GO"
    info "Running tests..."
    run "$GO" test $GOFLAGS ./...

    info "Building binaries..."
    run "$GO" build $GOFLAGS -o ccdbind ./cmd/ccdbind
    run "$GO" build $GOFLAGS -o ccdpin ./cmd/ccdpin
fi

# Verify binaries exist
for bin in ccdbind ccdpin; do
    [ -f "$SCRIPT_DIR/$bin" ] || die "Binary not found: $bin (run without --skip-build)"
done

# Install binaries
info "Installing binaries..."
install_file 755 "$SCRIPT_DIR/ccdbind" "$BINDIR/ccdbind"
install_file 755 "$SCRIPT_DIR/ccdpin"  "$BINDIR/ccdpin"

# Install systemd units
info "Installing systemd user units..."
install_file 644 "$SCRIPT_DIR/systemd/user/ccdbind.service" "$SYSTEMD_USER_DIR/ccdbind.service"
install_file 644 "$SCRIPT_DIR/systemd/user/game.slice"      "$SYSTEMD_USER_DIR/game.slice"

# Install example config (don't overwrite existing)
if [ ! -f "$CONFIGDIR/config.toml" ]; then
    info "Installing example configuration..."
    install_file 644 "$SCRIPT_DIR/config.example.toml" "$CONFIGDIR/config.toml"
else
    warn "Config file exists, not overwriting: $CONFIGDIR/config.toml"
fi

# Setup systemd service
if [ "$SKIP_SERVICE" = 0 ] && [ "$DRY_RUN" = 0 ]; then
    if command -v systemctl >/dev/null 2>&1; then
        info "Reloading systemd user daemon..."
        systemctl --user daemon-reload

        info "Enabling and starting ccdbind.service..."
        systemctl --user enable --now ccdbind.service
    else
        warn "systemctl not found, skipping service setup"
    fi
elif [ "$SKIP_SERVICE" = 1 ]; then
    info "Skipping systemd service setup (--skip-service)"
fi

ok "Installation complete!"
echo
info "Verify with:"
echo "    systemctl --user status ccdbind.service"
echo "    ccdbind status"
echo
info "Configuration:"
echo "    $CONFIGDIR/config.toml"
