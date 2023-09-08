#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

STATIC_DIR="./website/static"

PACKAGES=(
    cunicu
    hawkes
    gont/v2
    go-babel
    go-pmtud
    go-rosenpass
    go-openpgp-card
)

function generate() {
    mkdir -p "${STATIC_DIR}/${1}"
    cat > "${STATIC_DIR}/${1}/index.html" <<EOF
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en-us">
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <meta name="go-import" content="cunicu.li/${1} git https://github.com/cunicu/${1/\/v[0-9]/}.git">
    <meta http-equiv="Refresh" content="0; url=https://github.com/cunicu/${1/\/v[0-9]/}" />
  </head>
  <body>
  </body>
  <!--
  SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
  SPDX-License-Identifier: Apache-2.0
  -->
</html>
EOF
}

for i in "${!PACKAGES[@]}"; do
    PACKAGE=${PACKAGES[$i]}
    PACKAGE_BASE=${PACKAGE/\/v[0-9]/}

    generate ${PACKAGE}
    if [ ${PACKAGE} != ${PACKAGE_BASE} ]; then
        generate ${PACKAGE_BASE}
    fi
done