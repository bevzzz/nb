package nb

import (
	// Currently supported nbformat versions:
	_ "github.com/bevzzz/nb/schema/v4"
)

// Version returns current release version.
func Version() string {
	return "v0.1.0"
}