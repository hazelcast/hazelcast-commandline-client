#! /bin/bash

# Hazelcast CLC Install script
# (c) 2023 Hazelcast, Inc.

set -eu

check_ok () {
  local what="$1"
  local e=no
  which "$what" > /dev/null && e=yes
  case "$what" in
    awk*) state_awk_ok=$e;;
    curl*) state_curl_ok=$e;;
    sudo*) state_sudo_ok=$e;;
    tar*) state_tar_ok=$e;;
    unzip*) state_unzip_ok=$e;;
    wget*) state_wget_ok=$e;;
    xattr*) state_xattr_ok=$e;;
    *) log_debug "invalid check: $what"
  esac
}

log_info () {
    echo "INFO  $1" 1>&2
}

log_debug () {
    if [ "${state_debug}" == "yes" ]; then
        echo "DEBUG $1" 1>&2
    fi
}

bye () {
	if [[ "${1:-}" != "" ]]; then
		echo "ERROR $*" 1>&2
	fi
  exit 1
}

print_usage () {
  echo "This script installs Hazelcast CLC to a system or user directory."
  echo
	echo "Usage: $0 [--beta | --debug | --help]"
	echo
	echo "    --beta   Enable downloading BETA and PREVIEW releases"
	echo "    --debug  Enable DEBUG logging"
	echo "    --help   Show help"
	echo
	exit 0
}

setup () {
  detect_tmpdir
  for cmd in $DEPENDENCIES; do
      check_ok "$cmd"
  done
  detect_httpget
}

detect_tmpdir () {
	state_tmp_dir="${TMPDIR:-/tmp}"
}

do_curl () {
	curl -Ls "$1"
}

do_wget () {
	wget -O- "$1"
}

detect_uncompress () {
    local ext=${state_archive_ext}
    if [[ "$ext" == "tar.gz" ]]; then
        state_uncompress=do_untar
    elif [[ "$ext" == "zip" ]]; then
        state_uncompress=do_unzip
    else
        bye "$ext archive is not supported"
    fi
}

do_untar () {
  local path="$1"
  local base="$2"
  tar xf "$path" -C "$base"
}

do_unzip () {
  local path="$1"
  local base="$2"
  unzip -o -q "$path" -d "$base"
}

install_release () {
    # create base
    local tmp="${state_tmp_dir}"
    local base="$tmp/clc"
    mkdir -p "$base"
    # uncompress release package
    local path="${state_archive_path}"
    log_debug "UNCOMPRESS: $path => $base"
    ${state_uncompress} "$path" "$base"
    # move files to their place
    base="$base/${state_clc_name}"
    local bin="$home/bin/clc"
    mv_path "$base/clc" "$bin"
    local files="README.txt LICENSE.txt"
    for item in $files; do
        mv_path "$base/$item" "$home/$item"
    done
    # on MacOS remove the clc binary from quarantine
    if [[ "$state_xattr_ok" == "yes" && "$state_os" == "darwin" ]]; then
      set +e
      remove_from_quarantine "$bin"
      set -e
    fi
}

remove_from_quarantine () {
  local qa
  local path
  qa="com.apple.quarantine"
  path="$1"
  for a in $(xattr "$path"); do
    if [[ "$a" == "$qa" ]]; then
      log_debug "REMOVE FROM QUARANTINE: $path"
      xattr -d $qa "$path"
      break
    fi
  done
}

mv_path () {
    log_debug "MOVE $1 to $2"
    mv "$1" "$2"
}

detect_httpget () {
    if [[ "${state_curl_ok}" == "yes" ]]; then
      state_httpget=do_curl
    elif [[ "${state_wget_ok}" == "yes" ]]; then
      state_httpget=do_wget
    else
      bye "either curl or wget is required"
    fi
    log_debug "state_httpget=$state_httpget"
}

httpget () {
    log_debug "HTTP GET: ${state_httpget} $1"
    ${state_httpget} "$@"
}

print_banner () {
  echo
  echo "Hazelcast CLC Installer"
  echo "(c) 2023 Hazelcast, Inc."
  echo
}

print_success () {
  echo
  echo "Hazelcast CLC ${state_download_version} is installed at $home"
}

detect_last_release () {
  local re
  local text
  local v
  re='$1 ~ /tag_name/ { gsub(/[",]/, "", $2); print($2) }'
  text="$(httpget https://api.github.com/repos/hazelcast/hazelcast-commandline-client/releases)"
  if [[ "$state_beta" == "yes" ]]; then
    v=$(echo "$text" | awk "$re" | head -1)
  else
    v=$(echo "$text" | awk "$re" | grep -vi preview | grep -vi beta | head -1)
  fi
  if [[ "$v" == "" ]]; then
    bye "could not determine the latest version"
  fi
  state_download_version="$v"
  log_debug "state_download_version=$state_download_version"
}

detect_platform () {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux*) os=linux; ext="tar.gz";;
        Darwin*) os=darwin; ext="zip";;
        *) bye "This script supports only Linux and MacOS, not $os";;
    esac
    state_os=$os
    log_debug "state_os=$state_os"
    state_archive_ext=$ext
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
    state_arch="$arch"
    log_debug "state_arch=$state_arch"
}

make_download_url () {
    local v=${state_download_version}
    local clc_name=${state_clc_name}
    local ext=${state_archive_ext}
    state_download_url="https://github.com/hazelcast/hazelcast-commandline-client/releases/download/$v/${clc_name}.${ext}"
}

make_clc_name () {
    local v="${state_download_version}"
    local os="${state_os}"
    local arch="${state_arch}"
    state_clc_name="hazelcast-clc_${v}_${os}_${arch}"
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
    log_info "Downloading: ${state_download_url}"
    local tmp
    local ext
    tmp="${state_tmp_dir}"
    ext="${state_archive_ext}"
    state_archive_path="$tmp/clc.${ext}"
    httpget "${state_download_url}" > "${state_archive_path}"
}

process_flags () {
    for flag in "$@"; do
        case "$flag" in
          --beta*) state_beta=yes;;
          --debug*) state_debug=yes;;
      	  --help*) print_banner; print_usage;;
          *) bye "Unknown option: $flag";;
        esac
    done
}

DEPENDENCIES="awk wget curl sudo unzip tar xattr"

state_beta=no
state_debug=no
state_tmp_dir=
state_archive_ext=
state_archive_path=
state_download_url=
state_archive_path=
state_download_version=
state_clc_name=
state_archive_ext=
state_arch=
state_os=
state_httpget=
state_uncompress=

state_awk_ok=no
state_curl_ok=no
state_sudo_ok=no
state_tar_ok=no
state_unzip_ok=no
state_wget_ok=no
state_xattr_ok=no

home="${CLC_HOME:-$HOME/.hazelcast}"

process_flags "$@"
print_banner
setup
create_home
download_release
install_release
print_success
