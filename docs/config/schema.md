---
sidebar_position: 4
# SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# JSON Schema

[JSON Schema](https://json-schema.org/) is a declarative language that allows you to annotate and validate JSON documents.

JSON Schema can also be used to validate YAML documents and as such cunīcu's configuration file.
YAML Ain't Markup Language (YAML) is a powerful data serialization language that aims to be human friendly.

Most JSON is syntactically valid YAML, but idiomatic YAML follows very different conventions.
While YAML has advanced features that cannot be directly mapped to JSON, most YAML files use features that can be validated by JSON Schema.
JSON Schema is the most portable and broadly supported choice for YAML validation.

The schema of cunīcu's configuration file is available at:

- [`etc/cunicu.schema.yaml`](https://github.com/cunicu/cunicu/blob/main/etc/cunicu.schema.yaml)
- https://cunicu.li/schemas/config.yaml

## Editor / Language Server support

Redhat's [YAML Visual Studio Code extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) provides comprehensive YAML Language support, via the [yaml-language-server](https://github.com/redhat-developer/yaml-language-server).

It provides completion, validation and code lenses based on JSON Schemas.

To make use of it, you need to associate your config file with the JSON Schema by adding the following line into your config:

```yaml
# yaml-language-server: $schema=https://cunicu.li/schemas/config.yaml
---

watch_interval: 1s
```

## Reference

Here is a rendered reference based on this schema:

import ApiSchema from '@theme/ApiSchema';

<ApiSchema pointer="#/components/schemas/Config" />
