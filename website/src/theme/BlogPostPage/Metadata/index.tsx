// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import {PageMetadata} from '@docusaurus/theme-common';
import Metadata from '@theme-original/BlogPostPage/Metadata';
import type MetadataType from '@theme/BlogPostPage/Metadata';
import type {WrapperProps} from '@docusaurus/types';
import {useBlogPost} from '@docusaurus/theme-common/internal';

type Props = WrapperProps<typeof MetadataType>;

export default function MetadataWrapper(props: Props): JSX.Element {
  const post = useBlogPost();
  const username = post.metadata.authors[0].discourse || "";

  return (
    <>
      <Metadata {...props} />
      <PageMetadata>
      <meta
        name="discourse-username"
        content={username}
      />
      </PageMetadata>
    </>
  );
}
