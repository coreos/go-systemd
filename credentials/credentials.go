// Package credentials provides functions for working with systemd service credentials as documented by https://www.freedesktop.org/software/systemd/man/systemd.exec.html#Credentials.
package credentials

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func directory() (string, bool) {
	return os.LookupEnv("CREDENTIALS_DIRECTORY")
}

// Available returns whether credentials are available to the process
func Available() bool {
	_, available := directory()
	return available
}

// All returns a map of all available credentials
func All() map[string]string {
	dir, ok := directory()
	if !ok {
		return nil
	}

	filesystem := os.DirFS(dir)
	matches, _ := fs.Glob(filesystem, "*")

	creds := make(map[string]string)
	for _, match := range matches {
		data, _ := fs.ReadFile(filesystem, match)
		str := strings.TrimRight(string(data), "\n")

		basename := filepath.Base(match)
		creds[basename] = str
	}

	return creds
}

// Get returns a credential for a specified name
func Get(key string) (string, bool) {
	creds := All()
	cred, ok := creds[key]
	return cred, ok
}

// AsEnvironment loads all available credentials into the process environment, prefixing
// each credential name with SYSTEMD_CREDENTIAL_
// Note that this may nullify some of the security benefits of using LoadCredential/SetCredential
func AsEnvironment() error {
	creds := All()

	for key, value := range creds {
		if err := os.Setenv("SYSTEMD_CREDENTIAL_"+key, value); err != nil {
			return fmt.Errorf("error setting credential: %w", err)
		}
	}

	return nil
}
