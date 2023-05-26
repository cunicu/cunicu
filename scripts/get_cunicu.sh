#!/usr/bin/env bash

# Copyright The cunicu Authors.
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

# The install script is based off of the Apache-2.0-licensed script from Helm:
#  https://github.com/helm/helm/blob/main/scripts/get-helm-3

: ${BINARY_NAME:="cunicu"}
: ${USE_SUDO:="true"}
: ${DEBUG:="false"}
: ${VERIFY_CHECKSUM:="true"}
: ${VERIFY_SIGNATURES:="true"}
: ${INSTALL_DIR:="/usr/local/bin"}

HAS_WGET="$(type "wget" &> /dev/null && echo true || echo false)"
HAS_SHA256SUM="$(type "sha256sum" &> /dev/null && echo true || echo false)"
HAS_GPG="$(type "gpg" &> /dev/null && echo true || echo false)"
HAS_TAR="$(type "tar" &> /dev/null && echo true || echo false)"

# Settings
GITHUB_URL="https://github.com/stv0g/cunicu"
SUPPORTED_PLATFORMS=(darwin-amd64 darwin-arm64 linux-amd64 linux-arm linux-armv6 linux-arm64 windows-amd64 windows-arm64)

# detectArch discovers the architecture for this system.
function detectArch() {
  ARCH=$(uname -m)
  case ${ARCH} in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac

  echo "Detected architecture: ${ARCH}"
}

# detectOS discovers the operating system for this system.
function detectOS() {
  OS=$(uname | tr '[:upper:]' '[:lower:]')

  case "${OS}" in
    # Minimalist GNU for Windows
    mingw*|cygwin*)
      OS="windows"
      ;;
  esac

  echo "Detected operating system: ${OS}"
}

# runs the given command as root (detects if we are root already)
function runAsRoot() {
  if (( EUID != 0 )) && [[ "${USE_SUDO}" == "true" ]]; then
    sudo "${@}"
  else
    "${@}"
  fi
}

# verifySupported checks that the os/arch combination is supported for
# binary builds, as well whether or not necessary tools are present.
function verifySupported() {
  if ! [[ ${SUPPORTED_PLATFORMS[*]} =~ (^|[[:space:]])${OS}-${ARCH}($|[[:space:]]) ]]; then
    echo -e "No prebuilt binary for ${OS}-${ARCH}."
    echo -e "To build from source, go to ${GITHUB_URL}"
    exit 1
  fi

  if [[ "${HAS_WGET}" != "true" ]]; then
    echo -e "wget is required"
    exit 1
  fi

  if [[ "${HAS_TAR}" != "true" ]]; then
    echo -e "tar is required"
    exit 1
  fi

  if [[ "${VERIFY_CHECKSUM}" == "true" && "${HAS_SHA256SUM}" != "true" ]]; then
    echo -e "In order to verify checksum, sha256sum must first be installed."
    echo -e "Please install sha256sum or set VERIFY_CHECKSUM=false in your environment."
    exit 1
  fi

  if [[ "${VERIFY_SIGNATURES}" == "true" ]]; then
    if [[ "${OS}" != "linux" ]]; then
      echo -e "Signature verification is currently only supported on Linux."
      echo -e "Please set VERIFY_SIGNATURES=false or verify the signatures manually."
      exit 1
    elif [[ "${HAS_GPG}" != "true" ]]; then
      echo -e "In order to verify signatures, gpg must first be installed."
      echo -e "Please install gpg or set VERIFY_SIGNATURES=false in your environment."
      exit 1
    fi
  fi
}

# checkDesiredVersion checks if the desired version is available.
function checkDesiredVersion() {
  local latest_release_url="${GITHUB_URL}/releases"

  if [[ -z "${DESIRED_TAG}" ]]; then
    TAG=$(wget ${latest_release_url} -O - 2>&1 | grep 'href="/stv0g/cunicu/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | sed -E 's/.*\/stv0g\/cunicu\/releases\/tag\/(v[0-9\.]+)".*/\1/g' | head -1)
    echo "Latest available version of ${BINARY_NAME} is ${TAG}"
  else
    TAG=${DESIRED_TAG}
    echo "Installing requested version ${TAG}"
  fi

  VERSION=${TAG#v}
}

# checkInstalledVersion checks which version of cunicu is installed and
# if it needs to be changed.
function checkInstalledVersion() {
  if [[ -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
    local installed_version=$("${INSTALL_DIR}/${BINARY_NAME}" version -s)
    if [[ "${installed_version}" == "${TAG}" ]]; then
      echo "Installed version of ${BINARY_NAME} is ${installed_version} which is already ${DESIRED_TAG:-latest}"
      return 0
    else
      echo "New version of ${BINARY_NAME} is available: ${TAG}."
      echo "Updating from ${BINARY_NAME} version ${installed_version} to ${TAG}"
      return 1
    fi
  else
    return 1
  fi
}

# downloadArchive downloads the latest binary package and also the checksum
# for that binary.
function downloadArchive() {
  local suffix="tar.gz"
  if [[ ${OS} == "windows" ]]; then
    suffix="zip"
  fi

  DIST_FILE="cunicu_${VERSION}_${OS}_${ARCH}.${suffix}"

  downloadFile "${DIST_FILE}"
}

# verifyArchive verifies the SHA256 checksum of the binary package
# and the GPG signatures for both the package and checksum file
# (depending on settings in environment).
function verifyArchive() {
  if [[ "${VERIFY_CHECKSUM}" == "true" ]]; then
    verifyChecksum
  fi

  if [[ "${VERIFY_SIGNATURES}" == "true" ]]; then
    setupKeyring
    verifyChecksumSignature
  fi
}

# installBinary installs the cunicu binary.
function installBinary() {
  tar -xzf "${TMP_ROOT}/${DIST_FILE}" -C "${TMP_ROOT}"

  runAsRoot cp "${TMP_ROOT}/${BINARY_NAME}" "${INSTALL_DIR}"

  if [[ ${OS} != "windows" ]]; then
    runAsRoot chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
  fi

  echo "Installed ${BINARY_NAME} into ${INSTALL_DIR}"
}

# verifyChecksum verifies the SHA256 checksum of the binary package.
function verifyChecksum() {
  CHECKSUM_FILE="checksums.txt"

  downloadFile "${CHECKSUM_FILE}"

  local sum=$(sha256sum "${TMP_ROOT}/${DIST_FILE}" | cut -f1 -d" ")
  local sum_expected=$(grep "${DIST_FILE}" "${TMP_ROOT}/${CHECKSUM_FILE}" | cut -f1 -d" ")
  
  if [[ "${sum}" != "${sum_expected}" ]]; then
    echo -e "SHA sum of ${DIST_FILE} does not match. Aborting."
    exit 1
  fi
  
  echo "Verified binary checksum"
}

# downloadFile downloads a file from the Github releases
function downloadFile() {
  local filename=$1
  local github_release_url="${GITHUB_URL}/releases/download/${TAG}"
  local download_url="${github_release_url}/${filename}"

  echo "Downloading ${download_url}"
  wget -q -O "${TMP_ROOT}/${filename}" "${download_url}"
}

# setupKeyring initialize the GPG keyring for signature verification
function setupKeyring() {
  local keys_url="https://keys.openpgp.org/vks/v1/by-fingerprint/09BE3BAE8D55D4CD8579285A9675EAC34897E6E2"
  local gpg_homedir="${TMP_ROOT}/gnupg"

  GPG_KEYRING="${TMP_ROOT}/keyring.gpg"

  wget -q -O "${TMP_ROOT}/keys.asc" "${keys_url}"

  mkdir -p -m 0700 "${gpg_homedir}"

  gpg --batch --quiet --homedir="${gpg_homedir}" --import "${TMP_ROOT}/keys.asc"
  gpg --batch --no-default-keyring --keyring "${gpg_homedir}/pubring.kbx" --export > "${GPG_KEYRING}"
}

# verifySignatures checks that the signature of the checksum file matches
function verifyChecksumSignature() {
  CHECKSUM_SIG_FILE="${CHECKSUM_FILE}.asc"
  
  downloadFile "${CHECKSUM_SIG_FILE}"
   
  if ! gpg --verify --keyring="${GPG_KEYRING}" "${TMP_ROOT}/${CHECKSUM_SIG_FILE}" "${TMP_ROOT}/${CHECKSUM_FILE}"; then
    echo -e "Checksum signature in ${CHECKSUM_SIG_FILE} is invalid. Aborting"
    exit 1
  fi

  echo "Verified checksum signature"
}

# failTrap is executed if an error occurs.
function failTrap() {
  result=$?
  if (( result != 0 )); then
    if [[ -n "${INPUT_ARGUMENTS}" ]]; then
      echo -e "Failed to install ${BINARY_NAME} with the arguments provided: ${INPUT_ARGUMENTS}"
      help
    else
      echo -e "Failed to install ${BINARY_NAME}"
    fi
    echo -e "\tFor support, go to ${GITHUB_URL}."
  fi
  
  cleanup

  exit ${result}
}

# testVersion tests the installed client to make sure it is working.
function testVersion() {
  if ! checkInstalledVersion; then
    echo -e "Failed to install new version. Is ${INSTALL_DIR} in your PATH?"
    exit 1
  fi
}

# showHelp provides possible cli installation arguments
function showHelp () {
  echo -e "Accepted cli arguments are:"
  echo -e "\t[--help|-h ] ->> prints this help"
  echo -e "\t[--version|-v <DESIRED_TAG>] . When not defined it fetches the latest release from GitHub"
  echo -e "\te.g. --version v0.1.0"
  echo -e "\t[--no-sudo]  ->> install without sudo"
}

# cleanup removes the 1orary directory
function cleanup() {
  if [[ -d "${TMP_ROOT:-}" ]]; then
    rm -rf "${TMP_ROOT}"
  fi
}

# Execution

# Stop execution on any error
trap "failTrap" EXIT
set -e

# Set debug if desired
if [[ "${DEBUG}" == "true" ]]; then
  set -x
fi

# Parsing input arguments (if any)
INPUT_ARGUMENTS="${@}"
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--version'|-v)
      shift
      if (( # != 0 )); then
        DESIRED_TAG="${1}"
      else
        echo -e "Please provide the desired version. e.g. --version v0.1.0"
        exit 0
      fi
      ;;

    '--no-sudo')
      USE_SUDO="false"
      ;;

    '--help'|-h)
      help
      exit 0
      ;;

    *) exit 1
      ;;
  esac
  shift
done

set +u

TMP_ROOT="$(mktemp -dt cunicu-installer-XXXXXX)"  

detectArch
detectOS
verifySupported
checkDesiredVersion

if ! checkInstalledVersion; then
  downloadArchive
  verifyArchive
  installBinary
  testVersion
fi

cleanup
