// Package errors provides custom error types for Savanhi Shell.
// This package provides structured error handling with error codes,
// user-friendly messages, and error wrapping for context.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorCode represents a unique error code for programmatic handling.
type ErrorCode string

const (
	// General errors
	ErrUnknown         ErrorCode = "E0001"
	ErrInvalidArgument ErrorCode = "E0002"
	ErrNotImplemented  ErrorCode = "E0003"
	ErrCancelled       ErrorCode = "E0004"
	ErrTimeout         ErrorCode = "E0005"

	// Configuration errors (E001x)
	ErrConfigNotFound         ErrorCode = "E0010"
	ErrConfigInvalid          ErrorCode = "E0011"
	ErrConfigParseError       ErrorCode = "E0012"
	ErrConfigWriteError       ErrorCode = "E0013"
	ErrConfigPermissionDenied ErrorCode = "E0014"

	// Detection errors (E002x)
	ErrDetectionFailed         ErrorCode = "E0020"
	ErrOSDetectionFailed       ErrorCode = "E0021"
	ErrShellDetectionFailed    ErrorCode = "E0022"
	ErrTerminalDetectionFailed ErrorCode = "E0023"
	ErrFontDetectionFailed     ErrorCode = "E0024"
	ErrConfigDetectionFailed   ErrorCode = "E0025"

	// Installation errors (E003x)
	ErrInstallFailed           ErrorCode = "E0030"
	ErrDownloadFailed          ErrorCode = "E0031"
	ErrChecksumMismatch        ErrorCode = "E0032"
	ErrDependencyResolution    ErrorCode = "E0033"
	ErrBinaryNotFound          ErrorCode = "E0034"
	ErrInstallPermissionDenied ErrorCode = "E0035"
	ErrInstallTimeout          ErrorCode = "E0036"
	ErrInstallCanceled         ErrorCode = "E0037"

	// Rollback errors (E004x)
	ErrRollbackFailed  ErrorCode = "E0040"
	ErrBackupNotFound  ErrorCode = "E0041"
	ErrBackupCorrupted ErrorCode = "E0042"
	ErrRestoreFailed   ErrorCode = "E0043"

	// Preview errors (E005x)
	ErrPreviewFailed   ErrorCode = "E0050"
	ErrSubshellStart   ErrorCode = "E0051"
	ErrSubshellTimeout ErrorCode = "E0052"
	ErrThemeNotFound   ErrorCode = "E0053"
	ErrFontNotFound    ErrorCode = "E0054"

	// Persistence errors (E006x)
	ErrPersistenceFailed    ErrorCode = "E0060"
	ErrFileNotFound         ErrorCode = "E0061"
	ErrFilePermissionDenied ErrorCode = "E0062"
	ErrFileCorrupted        ErrorCode = "E0063"
	ErrJSONParseError       ErrorCode = "E0064"
	ErrJSONWriteError       ErrorCode = "E0065"

	// Shell errors (E007x)
	ErrShellNotFound     ErrorCode = "E0070"
	ErrRCNotFound        ErrorCode = "E0071"
	ErrRCReadError       ErrorCode = "E0072"
	ErrRCWriteError      ErrorCode = "E0073"
	ErrRCInvalidMarkers  ErrorCode = "E0074"
	ErrShellNotSupported ErrorCode = "E0075"

	// TUI errors (E008x)
	ErrTUIInitFailed  ErrorCode = "E0080"
	ErrTUIRenderError ErrorCode = "E0081"
	ErrTUIInputError  ErrorCode = "E0082"

	// Network errors (E009x)
	ErrNetworkError      ErrorCode = "E0090"
	ErrNetworkTimeout    ErrorCode = "E0091"
	ErrNetworkConnection ErrorCode = "E0092"
	ErrNetworkSSL        ErrorCode = "E0093"

	// System errors (E010x)
	ErrSystemPermission ErrorCode = "E0100"
	ErrSystemDiskFull   ErrorCode = "E0101"
	ErrSystemMemory     ErrorCode = "E0102"
)

// ErrorCategory groups related error codes.
type ErrorCategory string

const (
	CategoryGeneral       ErrorCategory = "general"
	CategoryConfiguration ErrorCategory = "configuration"
	CategoryDetection     ErrorCategory = "detection"
	CategoryInstallation  ErrorCategory = "installation"
	CategoryRollback      ErrorCategory = "rollback"
	CategoryPreview       ErrorCategory = "preview"
	CategoryPersistence   ErrorCategory = "persistence"
	CategoryShell         ErrorCategory = "shell"
	CategoryTUI           ErrorCategory = "tui"
	CategoryNetwork       ErrorCategory = "network"
	CategorySystem        ErrorCategory = "system"
)

// Category returns the error category for an error code.
func (c ErrorCode) Category() ErrorCategory {
	switch {
	case strings.HasPrefix(string(c), "E001"):
		return CategoryConfiguration
	case strings.HasPrefix(string(c), "E002"):
		return CategoryDetection
	case strings.HasPrefix(string(c), "E003"):
		return CategoryInstallation
	case strings.HasPrefix(string(c), "E004"):
		return CategoryRollback
	case strings.HasPrefix(string(c), "E005"):
		return CategoryPreview
	case strings.HasPrefix(string(c), "E006"):
		return CategoryPersistence
	case strings.HasPrefix(string(c), "E007"):
		return CategoryShell
	case strings.HasPrefix(string(c), "E008"):
		return CategoryTUI
	case strings.HasPrefix(string(c), "E009"):
		return CategoryNetwork
	case strings.HasPrefix(string(c), "E010"):
		return CategorySystem
	default:
		return CategoryGeneral
	}
}

// SavanhiError is a custom error type with structured information.
type SavanhiError struct {
	// Code is the error code for programmatic handling.
	Code ErrorCode `json:"code"`

	// Message is the user-friendly error message.
	Message string `json:"message"`

	// Detail is the technical error detail (for debugging).
	Detail string `json:"detail,omitempty"`

	// Cause is the underlying error.
	Cause error `json:"-"`

	// Context contains additional context information.
	Context map[string]interface{} `json:"context,omitempty"`

	// Suggestion is a suggested action for the user.
	Suggestion string `json:"suggestion,omitempty"`

	// Recoverable indicates if the error can be recovered.
	Recoverable bool `json:"recoverable"`
}

// Error implements the error interface.
func (e *SavanhiError) Error() string {
	var sb strings.Builder

	sb.WriteString("[")
	sb.WriteString(string(e.Code))
	sb.WriteString("] ")
	sb.WriteString(e.Message)

	if e.Detail != "" {
		sb.WriteString(": ")
		sb.WriteString(e.Detail)
	}

	if e.Suggestion != "" {
		sb.WriteString("\n  Suggestion: ")
		sb.WriteString(e.Suggestion)
	}

	if e.Cause != nil {
		sb.WriteString("\n  Caused by: ")
		sb.WriteString(e.Cause.Error())
	}

	return sb.String()
}

// Unwrap implements the errors.Unwrap interface.
func (e *SavanhiError) Unwrap() error {
	return e.Cause
}

// Is implements the errors.Is interface for comparison.
func (e *SavanhiError) Is(target error) bool {
	t, ok := target.(*SavanhiError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// UserMessage returns a user-friendly error message.
func (e *SavanhiError) UserMessage() string {
	var sb strings.Builder

	sb.WriteString(e.Message)

	if e.Suggestion != "" {
		sb.WriteString("\n\nHint: ")
		sb.WriteString(e.Suggestion)
	}

	return sb.String()
}

// WithContext adds context information to the error.
func (e *SavanhiError) WithContext(key string, value interface{}) *SavanhiError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion to the error.
func (e *SavanhiError) WithSuggestion(suggestion string) *SavanhiError {
	e.Suggestion = suggestion
	return e
}

// WithCause wraps an underlying error.
func (e *SavanhiError) WithCause(cause error) *SavanhiError {
	e.Cause = cause
	return e
}

// WithDetail adds technical detail to the error.
func (e *SavanhiError) WithDetail(detail string) *SavanhiError {
	e.Detail = detail
	return e
}

// New creates a new SavanhiError.
func New(code ErrorCode, message string) *SavanhiError {
	return &SavanhiError{
		Code:    code,
		Message: message,
	}
}

// NewWithCause creates a new SavanhiError with a cause.
func NewWithCause(code ErrorCode, message string, cause error) *SavanhiError {
	return &SavanhiError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Wrap wraps an existing error with a SavanhiError.
func Wrap(code ErrorCode, message string, err error) *SavanhiError {
	return &SavanhiError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// WrapWithContext wraps an error with additional context.
func WrapWithContext(code ErrorCode, message string, err error, context map[string]interface{}) *SavanhiError {
	return &SavanhiError{
		Code:    code,
		Message: message,
		Cause:   err,
		Context: context,
	}
}

// IsSavanhiError checks if an error is a SavanhiError.
func IsSavanhiError(err error) bool {
	var e *SavanhiError
	return errors.As(err, &e)
}

// GetErrorCode extracts the error code from an error.
func GetErrorCode(err error) ErrorCode {
	var e *SavanhiError
	if errors.As(err, &e) {
		return e.Code
	}
	return ErrUnknown
}

// GetUserMessage extracts the user message from an error.
func GetUserMessage(err error) string {
	var e *SavanhiError
	if errors.As(err, &e) {
		return e.UserMessage()
	}
	return err.Error()
}

// IsRecoverable checks if an error is recoverable.
func IsRecoverable(err error) bool {
	var e *SavanhiError
	if errors.As(err, &e) {
		return e.Recoverable
	}
	return false
}

// IsCode checks if an error has a specific code.
func IsCode(err error, code ErrorCode) bool {
	var e *SavanhiError
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

// IsCategory checks if an error belongs to a specific category.
func IsCategory(err error, category ErrorCategory) bool {
	var e *SavanhiError
	if errors.As(err, &e) {
		return e.Code.Category() == category
	}
	return false
}

// Common error constructors

// ConfigNotFound returns a configuration not found error.
func ConfigNotFound(path string) *SavanhiError {
	return New(ErrConfigNotFound,
		"Configuration file not found").
		WithContext("path", path).
		WithSuggestion("Create a configuration file using 'savanhi config init'")
}

// ConfigInvalid returns a configuration invalid error.
func ConfigInvalid(path string, reason string) *SavanhiError {
	return New(ErrConfigInvalid,
		"Configuration file is invalid").
		WithContext("path", path).
		WithDetail(reason).
		WithSuggestion("Check the configuration file format and syntax")
}

// DetectionFailed returns a detection failed error.
func DetectionFailed(component string, err error) *SavanhiError {
	return NewWithCause(ErrDetectionFailed,
		fmt.Sprintf("Failed to detect %s", component), err).
		WithContext("component", component).
		WithSuggestion("Ensure the component is properly installed and accessible")
}

// InstallFailed returns an installation failed error.
func InstallFailed(component string, err error) *SavanhiError {
	return NewWithCause(ErrInstallFailed,
		fmt.Sprintf("Failed to install %s", component), err).
		WithContext("component", component).
		WithSuggestion("Check your internet connection and try again")
}

// DownloadFailed returns a download failed error.
func DownloadFailed(url string, err error) *SavanhiError {
	return NewWithCause(ErrDownloadFailed,
		"Download failed", err).
		WithContext("url", url).
		WithSuggestion("Check your internet connection and try again")
}

// ChecksumMismatch returns a checksum mismatch error.
func ChecksumMismatch(component string, expected, actual string) *SavanhiError {
	return New(ErrChecksumMismatch,
		"Download verification failed").
		WithContext("component", component).
		WithContext("expected", expected).
		WithContext("actual", actual).
		WithSuggestion("The downloaded file may be corrupted. Try again or use --skip-checksum")
}

// RollbackFailed returns a rollback failed error.
func RollbackFailed(reason string, err error) *SavanhiError {
	return NewWithCause(ErrRollbackFailed,
		"Rollback failed", err).
		WithContext("reason", reason).
		WithSuggestion("Your system may be in an inconsistent state. Check backup files manually")
}

// BackupNotFound returns a backup not found error.
func BackupNotFound(backupID string) *SavanhiError {
	return New(ErrBackupNotFound,
		"Backup not found").
		WithContext("backup_id", backupID).
		WithSuggestion("Run 'savanhi rollback --original' to restore from the original backup")
}

// PreviewFailed returns a preview failed error.
func PreviewFailed(reason string, err error) *SavanhiError {
	return NewWithCause(ErrPreviewFailed,
		"Preview failed", err).
		WithContext("reason", reason).
		WithSuggestion("Try running in non-interactive mode with --non-interactive")
}

// SubshellTimeout returns a subshell timeout error.
func SubshellTimeout(duration string) *SavanhiError {
	return New(ErrSubshellTimeout,
		"Preview timed out").
		WithContext("duration", duration).
		WithSuggestion("The preview took too long. Try a simpler configuration")
}

// RCNotFound returns an RC file not found error.
func RCNotFound(shell string) *SavanhiError {
	return New(ErrRCNotFound,
		"Shell configuration file not found").
		WithContext("shell", shell).
		WithSuggestion(fmt.Sprintf("Create the configuration file for %s first", shell))
}

// ShellNotSupported returns a shell not supported error.
func ShellNotSupported(shell string) *SavanhiError {
	return New(ErrShellNotSupported,
		"Shell not supported").
		WithContext("shell", shell).
		WithSuggestion("Supported shells are: zsh, bash")
}

// NetworkError returns a network error.
func NetworkError(operation string, err error) *SavanhiError {
	return NewWithCause(ErrNetworkError,
		fmt.Sprintf("Network error during %s", operation), err).
		WithSuggestion("Check your internet connection and try again")
}

// PermissionDenied returns a permission denied error.
func PermissionDenied(path string, operation string) *SavanhiError {
	return New(ErrSystemPermission,
		fmt.Sprintf("Permission denied for %s", operation)).
		WithContext("path", path).
		WithContext("operation", operation).
		WithSuggestion("Try running with elevated privileges or check file permissions")
}

// Canceled returns a canceled error.
func Canceled(operation string) *SavanhiError {
	return New(ErrCancelled,
		"Operation canceled").
		WithContext("operation", operation).
		WithSuggestion("Run the operation again if you want to continue")
}

// Timeout returns a timeout error.
func Timeout(operation string, duration string) *SavanhiError {
	return New(ErrTimeout,
		"Operation timed out").
		WithContext("operation", operation).
		WithContext("duration", duration).
		WithSuggestion("Try increasing the timeout with --timeout flag")
}
