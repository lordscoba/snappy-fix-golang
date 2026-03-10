package adapters

import (
	"fmt"
	"net/url"
	"strconv"

	logutil "github.com/snappy-fix-golang/pkg/logger"
)

// ResolvePortParsing attempts to parse a port string.
// If the string is not a simple integer, it tries to parse it as a URL
// to extract the port. It panics on critical parsing failures.
func ResolvePortParsing(port string, logger *logutil.Logger) string {
	// Try to convert the port string directly to an integer.
	// If it's a valid integer (e.g., "8080"), no further parsing is needed.
	if _, err := strconv.Atoi(port); err != nil {
		// If it's not an integer, assume it might be a URL and try to parse it.
		u, err := url.Parse(port)
		if err != nil {
			// If URL parsing itself fails, log and panic with the actual URL parsing error.
			logutil.LogAndPrint(logger, fmt.Sprintf("parsing url %v to get port failed with: %v", port, err))
			panic(err)
		}

		// Attempt to extract the port from the parsed URL.
		detectedPort := u.Port()
		if detectedPort == "" {
			// If no port is detected from the URL, create a new error for this specific failure.
			newErr := fmt.Errorf("detecting port from url %v failed: no port found", port)
			logutil.LogAndPrint(logger, newErr.Error()) // Log the new error.
			panic(newErr)                               // Panic with the newly created error.
		}
		// If a port was successfully detected, use it.
		port = detectedPort
	}
	// Return the resolved port string.
	return port
}
