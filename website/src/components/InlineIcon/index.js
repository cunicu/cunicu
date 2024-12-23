// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import styles from './styles.module.css';

export default function Icon({src}) {
    return (
        <img className={styles.inlineicon} src={src} />
    );
}
