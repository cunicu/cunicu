package pb

import (
	"fmt"
	"io"
)

func (s *StatusResp) Dump(wr io.Writer, verbosity int) error {
	for k, i := range s.Interfaces {
		if k > 0 {
			if _, err := fmt.Fprintln(wr); err != nil {
				return err
			}
		}

		if err := i.Dump(wr, verbosity); err != nil {
			return err
		}
	}

	return nil
}
