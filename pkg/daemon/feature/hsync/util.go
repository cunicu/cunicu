// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package hsync

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

func isWritable(fn string) (bool, error) {
	f, err := os.OpenFile(fn, os.O_WRONLY, 0o000)
	if err != nil {
		if errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, f.Close()
}

func readLines(fn string) ([]string, error) {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_RDONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	lines := []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	return lines, scanner.Err()
}

func writeLines(fn string, lines []string) error {
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}
