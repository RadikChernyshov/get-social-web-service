package environment

import "os"

const (
	development = "development"
	production  = "production"
)

// Structure that represents the application environment with additional methods
type Environment struct {
}

// Gets the current value that was set in the application environment,
// if the requested variable doesn't exist returns the default value
func (e Environment) get(name string, defaultValue string) string {
	value, exist := os.LookupEnv(name)
	if !exist {
		return defaultValue
	}
	return value
}

// Check is the current application environment is in development mode
func Development() bool {
	env := Environment{}
	return env.get("GO_ENV", development) == development
}

// Check is the current application environment is in production mode
func Production() bool {
	env := Environment{}
	return env.get("GO_ENV", development) == production
}
