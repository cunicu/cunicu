#!/bin/bash

# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

set -e

function request() {
    RESOURCE=$1
    shift

    curl --silent \
         --location \
         --header "Accept: application/vnd.github+json" \
         --header "Authorization: Bearer ${GITHUB_TOKEN}" \
         "$@" "https://api.github.com/repos/${REPO}/${RESOURCE}"
}

function undraft_release() {
    request "releases/$1" -X PATCH -d '{ "draft": false }'  | \
    jq .
}

function get_draft_release() {
    request releases | jq '. | map(select(.draft == true)) | first'
}

function download_asset() {
    ASSET_NAME=$1
    ASSET_ID=$(jq -r ".assets | map(select(.name == \"${ASSET_NAME}\")) | first | .id")

    curl --silent \
         --location \
         --output "${ASSET_NAME}" \
         --header "Authorization: Bearer ${GITHUB_TOKEN}" \
         --header "Accept:application/octet-stream" \
         "https://api.github.com/repos/${REPO}/releases/assets/${ASSET_ID}"
}

function upload_asset() {
    RELEASE_ID=$1
    FILENAME=$2
    MIME_TYPE=$(file -b --mime-type "${FILENAME}")

    curl --silent \
         --location \
         --request POST \
         --header "Content-Type: ${MIME_TYPE}" \
         --header "Accept: application/vnd.github+json" \
         --header "Authorization: Bearer ${GITHUB_TOKEN}" \
         --data-binary "@${FILENAME}" \
         "https://uploads.github.com/repos/${REPO}/releases/${RELEASE_ID}/assets?name=${FILENAME}" | \
    jq .
}

REPO="stv0g/cunicu"

if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo -e "Missing GITHUB_TOKEN environment variable"
    exit -1
fi

RELEASE=$(get_draft_release)
if [[ -z "${RELEASE}" ]]; then
    echo -e "No drafted releases available"
    exit -1
fi

RELEASE_ID=$(jq -r .id <<< "${RELEASE}")
RELEASE_AUTHOR=$(jq -r .author.login <<< "${RELEASE}")
RELEASE_NAME=$(jq -r .name <<< "${RELEASE}")
RELEASE_CREATED_AT=$(jq -r .created_at <<< "${RELEASE}")
echo "Release ${RELEASE_NAME} (${RELEASE_ID}) created by ${RELEASE_AUTHOR} at ${RELEASE_CREATED_AT}"

download_asset checksums.txt <<< "${RELEASE}"
echo "Checksums:"
cat checksums.txt

gpg --batch \
    --yes \
    --detach-sign \
    --armor checksums.txt

echo "Checksum signature:"
cat checksums.txt.asc

upload_asset "${RELEASE_ID}" "checksums.txt.asc"
echo "Signature added to release."

undraft_release "${RELEASE_ID}"
echo "Release published."
