// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import { useEffect } from 'react';

export default function DiscourseCommentSystem(props) {
    useEffect(() => {
        window.DiscourseEmbed = {
            discourseUrl: 'https://discuss.cunicu.li/',
            // discourseEmbedUrl: 'https://cunicu.li',
            discourseEmbedUrl: window.location.toString(),
        };
    
        const d = document.createElement('script');
        d.type = 'text/javascript';
        d.async = true;
        d.src = window.DiscourseEmbed.discourseUrl + 'javascripts/embed.js';
        (document.getElementsByTagName('head')[0] || document.getElementsByTagName('body')[0]).appendChild(d);
      }, []);

  return (
    <>
      <div className="docusaurus-mt-lg">
        <div id="discourse-comments" />
      </div>
    </>
  );
}



