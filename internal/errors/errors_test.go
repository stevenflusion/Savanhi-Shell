// Package errors_test provides tests for the errors package.
package errors_test

import (
	stderrors "errors"
	"testing"

	savanhierrors "github.com/savanhi/shell/internal/errors"
)

func TestErrorCodeCategory(t *testing.T) {
	tests := []struct {
		code     savanhierrors.ErrorCode
		expected savanhierrors.ErrorCategory
	}{
		{savanhierrors.ErrConfigNotFound, savanhierrors.CategoryConfiguration},
		{savanhierrors.ErrConfigInvalid, savanhierrors.CategoryConfiguration},
		{savanhierrors.ErrDetectionFailed, savanhierrors.CategoryDetection},
		{savanhierrors.ErrOSDetectionFailed, savanhierrors.CategoryDetection},
		{savanhierrors.ErrInstallFailed, savanhierrors.CategoryInstallation},
		{savanhierrors.ErrDownloadFailed, savanhierrors.CategoryInstallation},
		{savanhierrors.ErrRollbackFailed, savanhierrors.CategoryRollback},
		{savanhierrors.ErrBackupNotFound, savanhierrors.CategoryRollback},
		{savanhierrors.ErrPreviewFailed, savanhierrors.CategoryPreview},
		{savanhierrors.ErrPersistenceFailed, savanhierrors.CategoryPersistence},
		{savanhierrors.ErrShellNotFound, savanhierrors.CategoryShell},
		{savanhierrors.ErrTUIInitFailed, savanhierrors.CategoryTUI},
		{savanhierrors.ErrNetworkError, savanhierrors.CategoryNetwork},
		{savanhierrors.ErrSystemPermission, savanhierrors.CategorySystem},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := tt.code.Category(); got != tt.expected {
				t.Errorf("ErrorCode.Category() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "configuration file not found")

	if err.Code != savanhierrors.ErrConfigNotFound {
		t.Errorf("New().Code = %v, want %v", err.Code, savanhierrors.ErrConfigNotFound)
	}
	if err.Message != "configuration file not found" {
		t.Errorf("New().Message = %v, want %v", err.Message, "configuration file not found")
	}
}

func TestNewWithCause(t *testing.T) {
	cause := stderrors.New("underlying error")
	err := savanhierrors.NewWithCause(savanhierrors.ErrInstallFailed, "installation failed", cause)

	if err.Code != savanhierrors.ErrInstallFailed {
		t.Errorf("NewWithCause().Code = %v, want %v", err.Code, savanhierrors.ErrInstallFailed)
	}
	if err.Cause != cause {
		t.Errorf("NewWithCause().Cause = %v, want %v", err.Cause, cause)
	}
}

func TestWrap(t *testing.T) {
	cause := stderrors.New("underlying error")
	err := savanhierrors.Wrap(savanhierrors.ErrDownloadFailed, "download failed", cause)

	if err.Code != savanhierrors.ErrDownloadFailed {
		t.Errorf("Wrap().Code = %v, want %v", err.Code, savanhierrors.ErrDownloadFailed)
	}
	if err.Cause != cause {
		t.Errorf("Wrap().Cause = %v, want %v", err.Cause, cause)
	}
}

func TestSavanhiErrorError(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "file not found")
	errStr := err.Error()

	if errStr == "" {
		t.Error("Error() returned empty string")
	}
}

func TestSavanhiErrorUnwrap(t *testing.T) {
	cause := stderrors.New("underlying")
	err := savanhierrors.Wrap(savanhierrors.ErrInstallFailed, "install failed", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap() did not return the cause")
	}
}

func TestSavanhiErrorIs(t *testing.T) {
	err1 := savanhierrors.New(savanhierrors.ErrConfigNotFound, "error 1")
	err2 := savanhierrors.New(savanhierrors.ErrConfigNotFound, "error 2")

	if !stderrors.Is(err1, err2) {
		t.Error("Is() should return true for same error codes")
	}

	err3 := savanhierrors.New(savanhierrors.ErrConfigInvalid, "error 3")
	if stderrors.Is(err1, err3) {
		t.Error("Is() should return false for different error codes")
	}
}

func TestSavanhiErrorUserMessage(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "file not found").
		WithSuggestion("Create the file first")

	msg := err.UserMessage()

	if msg == "" {
		t.Error("UserMessage() returned empty string")
	}
}

func TestSavanhiErrorWithContext(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "file not found").
		WithContext("path", "/path/to/config")

	if err.Context == nil {
		t.Error("WithContext() did not set context")
	}
	if err.Context["path"] != "/path/to/config" {
		t.Errorf("Context[path] = %v, want %v", err.Context["path"], "/path/to/config")
	}
}

func TestSavanhiErrorWithDetail(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "file not found").
		WithDetail("file does not exist at path")

	if err.Detail != "file does not exist at path" {
		t.Errorf("WithDetail() = %v, want %v", err.Detail, "file does not exist at path")
	}
}

func TestIsSavanhiError(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "test")
	if !savanhierrors.IsSavanhiError(err) {
		t.Error("IsSavanhiError() returned false for SavanhiError")
	}

	stdErr := stderrors.New("standard error")
	if savanhierrors.IsSavanhiError(stdErr) {
		t.Error("IsSavanhiError() returned true for standard error")
	}
}

func TestGetErrorCode(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "test")
	code := savanhierrors.GetErrorCode(err)

	if code != savanhierrors.ErrConfigNotFound {
		t.Errorf("GetErrorCode() = %v, want %v", code, savanhierrors.ErrConfigNotFound)
	}

	stdErr := stderrors.New("standard error")
	code = savanhierrors.GetErrorCode(stdErr)

	if code != savanhierrors.ErrUnknown {
		t.Errorf("GetErrorCode() for standard error = %v, want %v", code, savanhierrors.ErrUnknown)
	}
}

func TestGetUserMessage(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "file not found").
		WithSuggestion("create it")
	msg := savanhierrors.GetUserMessage(err)

	if msg == "" {
		t.Error("GetUserMessage() returned empty string")
	}
}

func TestIsRecoverable(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "test")
	err.Recoverable = true

	if !savanhierrors.IsRecoverable(err) {
		t.Error("IsRecoverable() returned false for recoverable error")
	}

	err.Recoverable = false
	if savanhierrors.IsRecoverable(err) {
		t.Error("IsRecoverable() returned true for non-recoverable error")
	}
}

func TestIsCode(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "test")

	if !savanhierrors.IsCode(err, savanhierrors.ErrConfigNotFound) {
		t.Error("IsCode() returned false for matching code")
	}

	if savanhierrors.IsCode(err, savanhierrors.ErrConfigInvalid) {
		t.Error("IsCode() returned true for non-matching code")
	}
}

func TestIsCategory(t *testing.T) {
	err := savanhierrors.New(savanhierrors.ErrConfigNotFound, "test")

	if !savanhierrors.IsCategory(err, savanhierrors.CategoryConfiguration) {
		t.Error("IsCategory() returned false for matching category")
	}

	if savanhierrors.IsCategory(err, savanhierrors.CategoryInstallation) {
		t.Error("IsCategory() returned true for non-matching category")
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	t.Run("ConfigNotFound", func(t *testing.T) {
		err := savanhierrors.ConfigNotFound("/path/to/config")
		if err.Code != savanhierrors.ErrConfigNotFound {
			t.Errorf("ConfigNotFound().Code = %v, want %v", err.Code, savanhierrors.ErrConfigNotFound)
		}
	})

	t.Run("ConfigInvalid", func(t *testing.T) {
		err := savanhierrors.ConfigInvalid("/path/to/config", "invalid syntax")
		if err.Code != savanhierrors.ErrConfigInvalid {
			t.Errorf("ConfigInvalid().Code = %v, want %v", err.Code, savanhierrors.ErrConfigInvalid)
		}
	})

	t.Run("DetectionFailed", func(t *testing.T) {
		cause := stderrors.New("underlying")
		err := savanhierrors.DetectionFailed("system", cause)
		if err.Code != savanhierrors.ErrDetectionFailed {
			t.Errorf("DetectionFailed().Code = %v, want %v", err.Code, savanhierrors.ErrDetectionFailed)
		}
	})

	t.Run("InstallFailed", func(t *testing.T) {
		cause := stderrors.New("underlying")
		err := savanhierrors.InstallFailed("oh-my-posh", cause)
		if err.Code != savanhierrors.ErrInstallFailed {
			t.Errorf("InstallFailed().Code = %v, want %v", err.Code, savanhierrors.ErrInstallFailed)
		}
	})

	t.Run("DownloadFailed", func(t *testing.T) {
		cause := stderrors.New("network error")
		err := savanhierrors.DownloadFailed("https://example.com/file", cause)
		if err.Code != savanhierrors.ErrDownloadFailed {
			t.Errorf("DownloadFailed().Code = %v, want %v", err.Code, savanhierrors.ErrDownloadFailed)
		}
	})

	t.Run("ChecksumMismatch", func(t *testing.T) {
		err := savanhierrors.ChecksumMismatch("file", "expected", "actual")
		if err.Code != savanhierrors.ErrChecksumMismatch {
			t.Errorf("ChecksumMismatch().Code = %v, want %v", err.Code, savanhierrors.ErrChecksumMismatch)
		}
	})

	t.Run("PermissionDenied", func(t *testing.T) {
		err := savanhierrors.PermissionDenied("/etc/config", "write")
		if err.Code != savanhierrors.ErrSystemPermission {
			t.Errorf("PermissionDenied().Code = %v, want %v", err.Code, savanhierrors.ErrSystemPermission)
		}
	})

	t.Run("Canceled", func(t *testing.T) {
		err := savanhierrors.Canceled("installation")
		if err.Code != savanhierrors.ErrCancelled {
			t.Errorf("Canceled().Code = %v, want %v", err.Code, savanhierrors.ErrCancelled)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		err := savanhierrors.Timeout("download", "30s")
		if err.Code != savanhierrors.ErrTimeout {
			t.Errorf("Timeout().Code = %v, want %v", err.Code, savanhierrors.ErrTimeout)
		}
	})
}
