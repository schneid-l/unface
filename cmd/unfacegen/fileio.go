package main

import "os"

// execWrite is a thin wrapper around os.WriteFile so tests can reuse a
// named helper without pulling in os at call sites.
func execWrite(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
