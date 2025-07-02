package environment

import (
	"os"
	"slices"
)

func IsDebug() bool {
	env := os.Getenv("PROFILE")
	return slices.Contains([]string{"", "debug"}, env)
}
