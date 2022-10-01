#!/bin/bash

set -e

function request() {
    RESOURCE=$1
    shift

    curl --silent \
         --location \
         --header "Accept: application/vnd.github+json" \
         --header "Authorization: Bearer ${GITHUB_TOKEN}" \
         "$@" https://api.github.com/repos/stv0g/cunicu/${RESOURCE}
}

function undraft_release() {
    request releases/$1 -X PATCH -d '{ "draft": false }'
}

function get_draft_release() {
    request releases | jq '. | map(select(.draft == false)) | first'
}

function download_asset() {
    ASSET_NAME=$1

    ASSET_URL=$(jq -r ".assets | map(select(.name == \"${ASSET_NAME}\")) | first | .browser_download_url")

    curl --silent \
         --location \
         --output ${ASSET_NAME} \
         --header "Authorization: Bearer ${GITHUB_TOKEN}" \
         ${ASSET_URL}
}

function upload_asset() {
    RELEASE_ID=$1
    FILENAME=$2
    MIME_TYPE=$(file -b --mime-type ${FILENAME})

    curl --silent \
         --location \
         --request POST \
         --header "Content-Type: ${MIME_TYPE}" \
         --header "Accept: application/vnd.github+json" \
         --header "Authorization: Bearer ${GITHUB_TOKEN}" \
         --data-binary @${FILENAME} \
         "https://uploads.github.com/repos/stv0g/cunicu/releases/${RELEASE_ID}/assets?name=${FILENAME}" | \
    jq .
}

RELEASE=$(get_draft_release)
if [[ -z "${RELEASE}" ]]; then
    echo -e "No drafted releases available"
    exit -1
fi

RELEASE_ID=$(jq .id <<< "${RELEASE}")
echo "Release ID: ${RELEASE_ID}"

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
