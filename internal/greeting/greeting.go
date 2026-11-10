// Package greeting builds greeting messages.
package greeting

import "strings"

// Message returns a greeting for name. A blank name is treated as "Go".
func Message(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Go"
	}

	return "Hello, " + name + "!"
}
