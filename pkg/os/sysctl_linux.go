// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package os

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SetSysctl(name string, value any) error {
	parts := strings.ReplaceAll(name, ".", string(filepath.Separator))
	path := filepath.Join("/proc/sys", parts)

	f, err := os.OpenFile(path, os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%v", value); err != nil {
		return err
	}

	return nil
}

func SetSysctlMap(m map[string]any) error {
	for k, v := range m {
		if err := SetSysctl(k, v); err != nil {
			return fmt.Errorf("failed to set sysctl '%s' to '%v': %w", k, v, err)
		}
	}

	return nil
}
