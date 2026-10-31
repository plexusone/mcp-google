// Package envflag resolves a setting that a command-line flag and an
// environment variable can both supply.
package envflag

// Bool resolves a boolean setting. A flag given on the command line wins over
// the environment variable, so --flag=false turns the setting off even when the
// environment turns it on. Without the flag, the setting is on only when the
// variable is exactly "true".
func Bool(given, value bool, env string) bool {
	if given {
		return value
	}
	return env == "true"
}
