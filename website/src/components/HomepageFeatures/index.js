// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Easy to Use',
    Svg: require('@site/static/img/rocket.svg').default,
    description: (
      <>
        cunīcu was designed from the ground up to be easily installed and
        used to get your website up and running quickly.
      </>
    ),
  },
  {
    title: 'Connectivity everywhere',
    Svg: require('@site/static/img/webrtc_logo.svg').default,
    description: (
      <>
        cunīcu embraces open standards and uses various WebRTC-related RFCs like the Interactive Connectivity Establishment (ICE) to establish peer-to-peer connections even in restrictive network environments.
      </>
    ),
  },
  {
    title: 'Powered by WireGuard®',
    Svg: require('@site/static/img/wireguard_logo.svg').default,
    description: (
      <>
        cunīcu is using user- or kernelspace WireGuard® implementation to provide state-of-the-art security and performance.
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
