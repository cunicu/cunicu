// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import {useBlogPost} from '@docusaurus/theme-common/internal';
import BlogPostItem from '@theme-original/BlogPostItem';
import DiscourseComments from '@site/src/DiscourseComments';

export default function BlogPostItemWrapper(props) {
  const {isBlogPostPage} = useBlogPost();
  return (
    <>
      <BlogPostItem {...props} />
      {isBlogPostPage && <DiscourseComments />}
    </>
  );
}
