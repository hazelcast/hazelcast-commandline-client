#! /bin/bash

# Hazelcast CLC Install script
# (c) 2023 Hazelcast, Inc.

set -eu

check_ok () {
  local what="$1"
  local e=no
  which "$what" > /dev/null && e=yes
  state[${what}_ok]=$e
}

log_info () {
    echo "INFO  $1" 1>&2
}

log_debug () {
    if [ "${state[debug]}" == "yes" ]; then
        echo "DEBUG $1" 1>&2
    fi
}

bye () {
	if [ "${1:-}x" != "x" ]; then
		echo "ERROR $*" 1>&2
	fi
  exit 1
}

setup () {
  detect_tmpdir
  commands="wget curl sudo uname mkdir unzip tar"
  for cmd in $commands; do
      check_ok "$cmd"
  done
  detect_httpget
  for key in "${!state[@]}"; do
      log_debug "$key: ${state[${key}]}"
  done
}

detect_tmpdir () {
	state[tmp_dir]="${TMPDIR:-/tmp}"
}

do_curl () {
	curl -Ls "$1"
}

do_wget () {
	wget -O- "$1"
}

detect_uncompress () {
    local ext=${state[archive_ext]}
    if [ "$ext" == "tar.gz" ]; then
        state[uncompress]=do_untar
    elif [ "$ext" == "zip" ]; then
        state[uncompress]=do_unzip
    else
        bye "$ext archive is not supported"
    fi
}

do_untar () {
    local tmp="${state[tmp_dir]}"
    local base="$tmp/clc"
    mkdir -p "$base"
    tar xf "$1" -C "$base"
    base="$base/${state[clc_name]}"
    mv_path "$base/clc" "$home/bin/clc"
    local files="README.txt LICENSE.txt"
    for item in $files; do
        mv_path "$base/$item" "$home/$item"
    done
}

do_unzip () {
    unzip
}

uncompress_release () {
    local path="${state[archive_path]}"
    log_debug "UNCOMPRESS: $path"
    ${state[uncompress]} "$path"
}

mv_path () {
    log_debug "MOVE $1 to $2"
    mv "$1" "$2"
}

detect_httpget () {
    if [ "${state[curl_ok]}x" != "x" ]; then
      state[httpget]=do_curl
    elif [ "${state[wget_ok]}x" != "x" ]; then
      state[httpget]=do_wget
    else
      bye "either curl or wget is required"
    fi
}

httpget () {
    log_debug "HTTP GET: $1"
    ${state[httpget]} "$@"
}

print_banner () {
    echo "Hazelcast CLC Installer"
    echo "(c) 2023 Hazelcast, Inc."
    echo
}

print_usage () {
    echo "This script installs CLC to a system or user directory."
    echo
    echo "Usage: $0 [-v version]"
    exit 1
}

print_success () {
    echo
    echo "Hazelcast CLC ${state[download_version]} is installed at $home"
}

detect_last_release () {
        local v
        v=$(httpget https://api.github.com/repos/hazelcast/hazelcast-commandline-client/releases | sed -n 's/^\s*"tag_name":\s*"\(.*\)",/\1/p' | grep -v BETA | head -1)
        state[download_version]="$v"
}

detect_platform () {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux*) os=linux; ext="tar.gz";;
        Darwin*) os=darwin; ext="zip";;
        *) bye "This script supports only Linux and MacOS, not $os";;
    esac
    state[os]=$os
    state[archive_ext]=$ext
    arch="$(uname -m)"
    case "$arch" in
        x86_64*) arch=amd64;;
        amd64*) arch=amd64;;
        armv6l*) arch=arm;;
        armv7l*) arch=arm;;
        arm64*) arch=arm64;;
        aarch64*) arch=arm64;;
        *) bye "This script supports only 64bit Intel and 32/64bit ARM architecture, not $arch"
    esac
    state[arch]="$arch"
}

make_download_url () {
    local v=${state[download_version]}
    local clc_name=${state[clc_name]}
    local ext=${state[archive_ext]}
    state[download_url]="https://github.com/hazelcast/hazelcast-commandline-client/releases/download/$v/${clc_name}.${ext}"
}

make_clc_name () {
    local v="${state[download_version]}"
    local os="${state[os]}"
    local arch="${state[arch]}"
    state[clc_name]="hazelcast-clc_${v}_${os}_${arch}"
}

create_home () {
    log_info "Creating the Home directory: $home"
    mkdir -p "$home/bin" "$home/etc"
    echo "install-script" > "$home/etc/.source"
}

download_release () {
    detect_tmpdir
    detect_platform
    detect_uncompress
    detect_last_release
    make_clc_name
    make_download_url
    log_info "Downloading: ${state[download_url]}"
    local tmp="${state[tmp_dir]}"
    local ext="${state[archive_ext]}"
    state[archive_path]="$tmp/clc.${ext}"
    httpget "${state[download_url]}" > "${state[archive_path]}"
}

process_flags () {
    for flag in "$@"; do
        case "$flag" in
          --beta*) state[beta]=yes;;
          --debug*) state[debug]=yes;;
          *) bye "Unknown option: $flag";;
        esac
    done
}

# an associative array to store the existence of commands
declare -A state=()
state[beta]=no
state[debug]=no

home="${CLC_HOME:-$HOME/.hazelcast}"

process_flags "$@"
print_banner
setup
create_home
download_release
uncompress_release
print_success
