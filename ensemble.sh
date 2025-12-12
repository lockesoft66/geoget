#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASEBOX_DIR="${SCRIPT_DIR}/basebox"

select_basebox_binary()
{
    local candidate
    if [ -x "${BASEBOX_DIR}/binl64/basebox" ]; then
        candidate="${BASEBOX_DIR}/binl64/basebox"
    elif [ -x "${BASEBOX_DIR}/binl/basebox" ]; then
        candidate="${BASEBOX_DIR}/binl/basebox"
    elif [ -x "${BASEBOX_DIR}/binmac/basebox" ]; then
        candidate="${BASEBOX_DIR}/binmac/basebox"
    elif [ -x "${BASEBOX_DIR}/binnt/basebox.exe" ]; then
        candidate="${BASEBOX_DIR}/binnt/basebox.exe"
    else
        candidate=""
    fi

    if [ -z "$candidate" ]; then
        printf 'Error: Unable to locate the Basebox executable.\n' >&2
        exit 1
    fi

    printf '%s' "$candidate"
}

BASEBOX_EXEC="$(select_basebox_binary)"
USER_CONFIG_FILE="${BASEBOX_DIR}/basebox.conf"

exec "$BASEBOX_EXEC" -noprimaryconf -nolocalconf -conf "$USER_CONFIG_FILE" "$@"
