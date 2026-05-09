package cli

import "os"

// removeFile is a tiny wrapper to avoid pulling os into agent.go just for this.
func removeFile(path string) error {
	return os.Remove(path)
}
