package detect

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
)

// DetectOptions contains the options for detecting the environment.
type DetectOptions struct {
	// Platform limits the detection to find only environments for the given platform.
	Platform string

	// TargetType limits the detection to find only environments for the given target type (pull request or commit SHA)
	TargetType string
}

// DetectResult contains the result of a detection.
// It contains the platform, project (repo), target type and target ref (pull request number/commit SHA).
// It also contains any extra values the detector can detect, e.g. a token or API URL.
type DetectResult struct {
	Platform   string
	Project    string
	TargetType string
	TargetRef  string
	Extra      interface{}
}

// DetectError is an error that is returned when the environment could not be detected
// by the current detector.
type DetectError struct {
	err error
}

// Error returns the string message of the error.
func (e *DetectError) Error() string {
	return e.err.Error()
}

// Detector is the interface that must be implemented by a detector.
type Detector interface {
	DisplayName() string
	Detect(ctx context.Context, opts DetectOptions) (DetectResult, error)
}

// DetectorRegistryItem represents an item in the detector registry.
// It maps the detector to the platforms it detects.
type detectorRegistryItem struct {
	supportedPlatforms []string
	detector           Detector
}

// detectorRegistry contains the list of all detectors.
// Detectors are registered using the registerDetector function in the init()
// function of the files containing them.
var detectorRegistry = []detectorRegistryItem{}

// registerDetector registers a new detector in the detector registry,
// mapping it to the platforms it detects.
func registerDetector(supportedPlatforms []string, detector Detector) {
	detectorRegistry = append(detectorRegistry, detectorRegistryItem{
		supportedPlatforms: supportedPlatforms,
		detector:           detector,
	})
}

// DetectEnvironment detects the environment for a given platform and target type.
// It iterates through the detectors and returns the first one that detects the
// environment.
func DetectEnvironment(ctx context.Context, opts DetectOptions) (DetectResult, error) {
	for _, detectorRegistryItem := range detectorRegistry {
		if opts.Platform != "" && !contains(detectorRegistryItem.supportedPlatforms, opts.Platform) {
			continue
		}

		detector := detectorRegistryItem.detector

		log.Ctx(ctx).Debug().Msgf("Checking for %s", detector.DisplayName())

		result, err := detector.Detect(ctx, opts)
		if err != nil {
			if e, ok := err.(*DetectError); ok {
				log.Ctx(ctx).Debug().Err(e).Msgf("Could not detect %s environment", detector.DisplayName())
				continue
			} else {
				return DetectResult{}, err
			}
		} else {
			log.Ctx(ctx).Info().Msgf("Detected %s (platform: %s, target type: %s, target ref: %s)", detector.DisplayName(), result.Platform, result.TargetType, result.TargetRef)
			return result, nil
		}
	}

	return DetectResult{}, &DetectError{errors.New("Could not to detect environment")}
}

// contains returns true if the given string slice contains the given string.
func contains(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}
