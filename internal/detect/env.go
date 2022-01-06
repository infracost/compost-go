package detect

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// sanitizeValue returns a sanitized version of the given value.
// If the value is a secret, it will be replaced with a placeholder.
func sanitizeValue(val string, isSecret bool) string {
	if isSecret {
		return "************"
	}

	return val
}

// checkEnvVarExists checks if the given environment variable exists.
// If the variable is a secret, any logs will be replaced with a placeholder.
// If the variable exists the value of the variable is returned, if not then
// an error is thrown.
func checkEnvVarExists(ctx context.Context, name string, isSecret bool) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf("%s environment variable is not set", name)
	}

	log.Ctx(ctx).Debug().Msgf("%s environment variable is set to %s", name, sanitizeValue(val, isSecret))

	return val, nil
}

// checkEnvVarExistsOrEmpty checks if the given environment variable exists
// and has the expected value.
// If the variable is a secret, any logs will be replaced with a placeholder.
// If the variable does not exist or has an unexpected value then an error is thrown.
func checkEnvVarValue(ctx context.Context, name string, expected string, isSecret bool) error {
	val, err := checkEnvVarExists(ctx, name, isSecret)
	if err != nil {
		return err
	}

	if val != expected {
		return fmt.Errorf("%s environment variable is set to %s, expected %s", name, val, expected)
	}

	return nil
}
