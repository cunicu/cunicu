// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import BlogPostItem from '@theme-original/BlogPostItem';
import DiscourseComments from '@site/src/DiscourseComments';

export default function BlogPostItemWrapper(props) {
  return (
    <>
      <BlogPostItem {...props} />
      <DiscourseComments />
    </>
  );
}
