// Package cli provides command-line interface functionality for Savanhi Shell.
package cli

// Exit codes for programmatic handling.
const (
	// ExitSuccess indicates successful execution.
	ExitSuccess = 0

	// ExitError indicates a general error.
	ExitError = 1

	// ExitConfigError indicates a configuration error.
	ExitConfigError = 2

	// ExitDetectionError indicates a detection error.
	ExitDetectionError = 3

	// ExitInstallError indicates an installation error.
	ExitInstallError = 4

	// ExitRollbackError indicates a rollback error.
	ExitRollbackError = 5

	// ExitNetworkError indicates a network error.
	ExitNetworkError = 6

	// ExitPermissionError indicates a permission error.
	ExitPermissionError = 7

	// ExitTimeoutError indicates a timeout error.
	ExitTimeoutError = 8

	// ExitCanceledError indicates user cancellation.
	ExitCanceledError = 130
)

// ExitCodeFromError returns the appropriate exit code for an error.
func ExitCodeFromError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Import the errors package for error type checking
	// This maps error codes to exit codes
	code := GetErrorCodeFromError(err)

	switch code {
	case "E0010", "E0011", "E0012", "E0013", "E0014": // Config errors
		return ExitConfigError
	case "E0020", "E0021", "E0022", "E0023", "E0024", "E0025": // Detection errors
		return ExitDetectionError
	case "E0030", "E0031", "E0032", "E0033", "E0034", "E0035", "E0036", "E0037": // Install errors
		return ExitInstallError
	case "E0040", "E0041", "E0042", "E0043": // Rollback errors
		return ExitRollbackError
	case "E0090", "E0091", "E0092", "E0093": // Network errors
		return ExitNetworkError
	case "E0100": // Permission error
		return ExitPermissionError
	case "E0004": // Canceled
		return ExitCanceledError
	case "E0005": // Timeout
		return ExitTimeoutError
	default:
		return ExitError
	}
}

// GetErrorCodeFromError extracts the error code string from an error.
func GetErrorCodeFromError(err error) string {
	// This is a simplified version
	// In real implementation, we would check for SavanhiError type
	if serr, ok := err.(interface{ Code() string }); ok {
		return serr.Code()
	}
	return "E0001" // Unknown error
}
