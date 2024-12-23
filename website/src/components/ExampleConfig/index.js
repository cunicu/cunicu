// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import CodeBlock from '@theme/CodeBlock';
import ExampleConfigSource from '!!raw-loader!../../../../etc/cunicu.yaml';
import ExampleAdvancedConfigSource from '!!raw-loader!../../../../etc/cunicu.advanced.yaml';

export default function ExampleConfig(props) {
    let codeProps = {...props};

    codeProps.language ??= 'yaml';

    let content;
    if (codeProps.advanced) {
        content = ExampleAdvancedConfigSource;
        codeProps.title = '/etc/cunicu.advanced.yaml';
    } else {
        content = ExampleConfigSource;
        codeProps.title = '/etc/cunicu.yaml';
    }

    return <CodeBlock {...codeProps}>{content}</CodeBlock>;
}