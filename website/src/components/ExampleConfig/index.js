// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import CodeBlock from '@theme/CodeBlock';
import ExampleConfigSource from '!!raw-loader!../../../../etc/cunicu.yaml';

export default function ExampleConfig(props) {
    let codeProps = {...props};
    if (!codeProps.language) {
        codeProps.language = 'yaml';
    }

    codeProps.title = '/etc/cunicu.yaml';

    let content = ExampleConfigSource;

    if (codeProps.section) {
        const contentLines = content.split('\n');
        let commentLines = [];
        let sectionLines = [];
        let inSection = false;

        for (let line of contentLines) {
            let startsSection = false;
            let endsSection = false;
            let commentLine = line.startsWith('#');
            let emptyLine = line.trim() === '';

            let matches = line.match(/^([a-zA-z]+):/);
            if (matches !== null) {
                startsSection = matches[1] == codeProps.section;
                endsSection = matches[1] != codeProps.section;
            }

            if (commentLine) {
                inSection = false;
                commentLines.push(line);
            }
            
            if (startsSection) {
                inSection = true;

                sectionLines.push(...commentLines);
                commentLines = [];
            }

            if (endsSection)
                inSection = false;

            if (emptyLine)
                commentLines = [];

            if (inSection)
                sectionLines.push(line);
        }

        if (sectionLines[sectionLines.length - 1] == '')
            sectionLines = sectionLines.slice(0, -1);

        content = sectionLines.join('\n');

        codeProps.title = `Section "${codeProps.section}" of ${codeProps.title}`;
    }

    return <CodeBlock {...codeProps}>{content}</CodeBlock>;
}