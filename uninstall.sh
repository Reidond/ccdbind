#!/bin/sh
# uninstall.sh - Uninstall ccdbind and ccdpin
# SPDX-License-Identifier: MIT

set -e

PROG_NAME="${0##*/}"
PREFIX="${PREFIX:-$HOME/.local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
CONFIGDIR="${CONFIGDIR:-$HOME/.config/ccdbind}"
STATEDIR="${STATEDIR:-$HOME/.local/state/ccdbind}"
STATEDIR_PIN="${STATEDIR_PIN:-$HOME/.local/state/ccdpin}"
SYSTEMD_USER_DIR="${SYSTEMD_USER_DIR:-$HOME/.config/systemd/user}"

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

Uninstall ccdbind and ccdpin binaries and systemd user units.

Options:
    -h, --help          Show this help message
    -n, --dry-run       Print actions without executing
    -p, --purge         Also remove configuration and state files
    -f, --force         Don't prompt for confirmation
    --prefix=PATH       Install prefix (default: ~/.local)
    --bindir=PATH       Binary directory (default: PREFIX/bin)
    --configdir=PATH    Config directory (default: ~/.config/ccdbind)
    --statedir=PATH     State directory (default: ~/.local/state/ccdbind)

Environment:
    PREFIX              Install prefix
    BINDIR              Binary directory
    CONFIGDIR           Config directory
    STATEDIR            State directory

Examples:
    $PROG_NAME
    $PROG_NAME --purge
    $PROG_NAME --dry-run --purge
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

# Remove file if it exists
rm_file() {
    if [ -f "$1" ] || [ -L "$1" ]; then
        run rm -f "$1"
        return 0
    fi
    return 1
}

# Remove directory if it exists and is empty
rm_dir_if_empty() {
    if [ -d "$1" ]; then
        if [ -z "$(ls -A "$1" 2>/dev/null)" ]; then
            run rmdir "$1"
        else
            warn "Directory not empty, not removing: $1"
        fi
    fi
}

# Remove directory recursively
rm_dir() {
    if [ -d "$1" ]; then
        run rm -rf "$1"
    fi
}

# Prompt for confirmation
confirm() {
    if [ "$FORCE" = 1 ]; then
        return 0
    fi
    printf '%s [y/N] ' "$1"
    read -r answer
    case "$answer" in
        [Yy]|[Yy][Ee][Ss]) return 0 ;;
        *) return 1 ;;
    esac
}

DRY_RUN=0
PURGE=0
FORCE=0

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)      usage ;;
        -n|--dry-run)   DRY_RUN=1 ;;
        -p|--purge)     PURGE=1 ;;
        -f|--force)     FORCE=1 ;;
        --prefix=*)     PREFIX="$(parse_arg "$1")"; BINDIR="$PREFIX/bin" ;;
        --bindir=*)     BINDIR="$(parse_arg "$1")" ;;
        --configdir=*)  CONFIGDIR="$(parse_arg "$1")" ;;
        --statedir=*)   STATEDIR="$(parse_arg "$1")" ;;
        -*)             die "Unknown option: $1" ;;
        *)              die "Unexpected argument: $1" ;;
    esac
    shift
done

info "Uninstalling ccdbind and ccdpin"
info "  BINDIR:         $BINDIR"
info "  CONFIGDIR:      $CONFIGDIR"
info "  STATEDIR:       $STATEDIR"
info "  SYSTEMD_USER:   $SYSTEMD_USER_DIR"

if [ "$PURGE" = 1 ]; then
    warn "Purge mode enabled - config and state will be removed"
fi

if ! confirm "Proceed with uninstall?"; then
    info "Aborted."
    exit 0
fi

# Stop and disable systemd service
if [ "$DRY_RUN" = 0 ] && command -v systemctl >/dev/null 2>&1; then
    if systemctl --user is-active --quiet ccdbind.service 2>/dev/null; then
        info "Stopping ccdbind.service..."
        systemctl --user stop ccdbind.service || true
    fi
    if systemctl --user is-enabled --quiet ccdbind.service 2>/dev/null; then
        info "Disabling ccdbind.service..."
        systemctl --user disable ccdbind.service || true
    fi
fi

# Remove systemd units
info "Removing systemd user units..."
rm_file "$SYSTEMD_USER_DIR/ccdbind.service" && info "  Removed ccdbind.service"
rm_file "$SYSTEMD_USER_DIR/game.slice" && info "  Removed game.slice"

# Reload systemd
if [ "$DRY_RUN" = 0 ] && command -v systemctl >/dev/null 2>&1; then
    info "Reloading systemd user daemon..."
    systemctl --user daemon-reload || true
fi

# Remove binaries
info "Removing binaries..."
rm_file "$BINDIR/ccdbind" && info "  Removed ccdbind"
rm_file "$BINDIR/ccdpin" && info "  Removed ccdpin"

# Remove config and state if purging
if [ "$PURGE" = 1 ]; then
    info "Removing configuration..."
    rm_dir "$CONFIGDIR" && info "  Removed $CONFIGDIR"

    info "Removing state..."
    rm_dir "$STATEDIR" && info "  Removed $STATEDIR"
    rm_dir "$STATEDIR_PIN" && info "  Removed $STATEDIR_PIN"
else
    info "Config and state preserved (use --purge to remove)"
    info "  Config: $CONFIGDIR"
    info "  State:  $STATEDIR"
fi

ok "Uninstall complete!"
