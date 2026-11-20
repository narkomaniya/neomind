//go:build !windows

package main

import "errors"

// pickFolder is only available through the native Windows shell dialog.
// On other platforms the user can type an internal folder name directly.
func pickFolder(dialogTitle string) (string, error) {
	return "", errors.New("folder picker is available on Windows only")
}
