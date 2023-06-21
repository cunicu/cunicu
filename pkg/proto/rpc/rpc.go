// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"fmt"
	"io"

	"github.com/stv0g/cunicu/pkg/log"
)

func (s *GetStatusResp) Dump(wr io.Writer, level log.Level) error {
	for k, i := range s.Interfaces {
		if k > 0 {
			if _, err := fmt.Fprintln(wr); err != nil {
				return err
			}
		}

		if err := i.Dump(wr, level); err != nil {
			return err
		}
	}

	return nil
}
