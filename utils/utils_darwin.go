package utils

import (
	path "path/filepath"
)

func AppData() string {
	return path.Join(HomeDir(), "Library", "Application Support")
}
