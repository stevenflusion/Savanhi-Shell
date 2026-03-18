# Contributing to Savanhi Shell

Thank you for your interest in contributing to Savanhi Shell! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)

## Code of Conduct

This project and everyone participating in it is governed by basic principles of respect and inclusivity. By participating, you are expected to:

- Be respectful of differing viewpoints
- Gracefully accept constructive criticism
- Focus on what is best for the community
- Show empathy towards other community members

## How Can I Contribute?

### Reporting Bugs

See [Reporting Bugs](#reporting-bugs) below.

### Suggesting Enhancements

See [Requesting Features](#requesting-features) below.

### Writing Code

- Fix bugs
- Implement new features
- Improve documentation
- Add tests

### Reviewing Code

- Review pull requests
- Provide constructive feedback
- Help maintain code quality

## Development Setup

### Prerequisites

- **Go 1.21+**: Required for building the project
- **Make**: For build commands
- **Git**: For version control

### Getting Started

1. **Fork the repository**

   Click the "Fork" button on GitHub.

2. **Clone your fork**

   ```bash
   git clone https://github.com/YOUR_USERNAME/shell.git
   cd shell
   ```

3. **Add upstream remote**

   ```bash
   git remote add upstream https://github.com/savanhi/shell.git
   ```

4. **Install dependencies**

   ```bash
   go mod download
   ```

5. **Build the project**

   ```bash
   make build
   ```

6. **Run tests**

   ```bash
   make test
   ```

### Project Structure

```
savanhi-shell/
├── cmd/savanhi-shell/      # Entry point
├── internal/               # Internal packages
│   ├── cli/               # CLI logic
│   ├── detector/          # System detection
│   ├── errors/            # Error handling
│   ├── installer/         # Installation logic
│   ├── persistence/       # Data storage
│   ├── preview/           # Live preview
│   ├── staging/           # Change staging
│   └── tui/               # Terminal UI
├── pkg/shell/              # Public packages (shell manipulation)
├── configs/                # Bundled configurations
├── scripts/                # Install scripts
└── tests/                  # E2E tests
    └── e2e/
```

## Development Workflow

1. **Create a feature branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**

   Write code, following the [Code Style Guidelines](#code-style-guidelines).

3. **Write/update tests**

   Ensure all tests pass and new code has test coverage.

4. **Run the test suite**

   ```bash
   make test
   make coverage  # To see coverage report
   ```

5. **Run linters**

   ```bash
   make lint
   ```

6. **Commit your changes**

   Follow [Commit Guidelines](#commit-guidelines).

7. **Push to your fork**

   ```bash
   git push origin feature/your-feature-name
   ```

8. **Open a Pull Request**

   See [Pull Request Process](#pull-request-process).

## Code Style Guidelines

### Go Code Style

1. **Follow standard Go conventions**

   - Use `gofmt` for formatting
   - Follow [Effective Go](https://golang.org/doc/effective_go)
   - Use `golangci-lint` for linting

2. **Package naming**

   - Use short, lowercase names
   - No underscores or mixedCaps
   - Nouns preferred over verbs (e.g., `detector` not `detect`)

3. **Exported names**

   - Document all exported types, functions, and constants
   - Use doc comments that start with the name

   ```go
   // Detector is the interface for system detection.
   type Detector interface {
       // DetectOS returns information about the operating system.
       DetectOS() (*OSInfo, error)
   }
   ```

4. **Error handling**

   - Use the `internal/errors` package for structured errors
   - Wrap errors with context using `fmt.Errorf` or `errors.NewWithCause`
   - Never ignore errors

   ```go
   if err != nil {
       return errors.NewWithCause(errors.ErrDetectionFailed,
           "failed to detect shell", err)
   }
   ```

5. **Constants over magic values**

   ```go
   // Good
   const (
       DefaultTimeout = 10 * time.Minute
       MaxHistoryEntries = 1000
   )

   // Bad
   ctx, cancel := context.WithTimeout(context.Background(), 600000000000)
   ```

### Documentation

1. **Package documentation**

   Every package should have a doc comment explaining its purpose:

   ```go
   // Package detector provides system detection capabilities for Savanhi Shell.
   // It detects OS, shell, terminal, fonts, and existing configurations.
   package detector
   ```

2. **Function documentation**

   Document all exported functions:

   ```go
   // NewDetector creates a new DefaultDetector with all sub-detectors.
   // It returns a fully initialized detector ready for use.
   func NewDetector() *DefaultDetector {
       // ...
   }
   ```

3. **README updates**

   Update README.md if you:
   - Add new features
   - Change existing behavior
   - Add new command-line options

## Testing Requirements

### Unit Tests

- All new code must have unit tests
- Aim for >80% code coverage
- Use table-driven tests for multiple cases

```go
func TestDetectShell(t *testing.T) {
    tests := []struct {
        name    string
        env     map[string]string
        want    *ShellInfo
        wantErr bool
    }{
        {
            name: "detects zsh",
            env:  map[string]string{"SHELL": "/bin/zsh"},
            want: &ShellInfo{Name: "zsh"},
        },
        {
            name: "detects bash",
            env:  map[string]string{"SHELL": "/bin/bash"},
            want: &ShellInfo{Name: "bash"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

- Test module interactions
- Use temp directories for file operations
- Clean up resources in tests

### E2E Tests

Located in `tests/e2e/`. These tests:
- Build the binary
- Run actual commands
- Verify real behavior

Run with:

```bash
go test -v ./tests/e2e/
```

### Running All Tests

```bash
# Unit tests
make test

# With coverage
make coverage

# E2E tests
go test -v ./tests/e2e/

# All tests including E2E
make test-all
```

## Commit Guidelines

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

### Examples

```
feat(detector): add Windows Terminal detection

Add support for detecting Windows Terminal on WSL environments.
The detector now checks for WT_SESSION environment variable.

Closes #42
```

```
fix(installer): correct RC file backup path

The backup path was using the wrong home directory on some systems.
Now properly uses os.UserHomeDir() for cross-platform support.
```

```
docs(readme): update installation instructions

Add Homebrew installation method and update Go version requirement.
```

## Pull Request Process

### Before Submitting

1. **Ensure all tests pass**

   ```bash
   make test
   make lint
   ```

2. **Update documentation**

   - Update README.md if needed
   - Add/update godoc comments
   - Update configuration.md for config changes

3. **Squash commits**

   Keep your PR history clean. Squash related commits.

4. **Write a good PR description**

   - Describe what the PR does
   - Reference related issues
   - List any breaking changes

### PR Template

```markdown
## Description

[Describe your changes]

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist

- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] All tests pass
- [ ] No new warnings

## Related Issues

Closes #XXX
```

### Review Process

1. **Automated checks**: CI runs tests, linters, and builds
2. **Code review**: Maintainers review your code
3. **Feedback**: Address any requested changes
4. **Approval**: PR requires approval from at least one maintainer
5. **Merge**: Maintainer will merge the PR

### After Merge

- Delete your feature branch
- Update your local main branch:

```bash
git checkout main
git pull upstream main
```

## Reporting Bugs

### Before Reporting

1. **Search existing issues**

   Check if the bug has already been reported.

2. **Try the latest version**

   The bug might already be fixed.

### How to Report

Create a new issue with:

1. **Title**: Clear, descriptive title

2. **Description**:
   - What did you expect to happen?
   - What actually happened?
   - Steps to reproduce

3. **Environment**:
   - OS and version
   - Shell (zsh/bash)
   - Terminal emulator
   - Savanhi Shell version

4. **Logs**:
   - Command output
   - Error messages
   - `~/.config/savanhi/logs/` contents (if applicable)

### Bug Report Template

```markdown
## Description

[Clear description of the bug]

## Steps to Reproduce

1. Run `savanhi-shell --detect`
2. Select option X
3. Observe error

## Expected Behavior

[What you expected to happen]

## Actual Behavior

[What actually happened]

## Environment

- OS: macOS 14.0
- Shell: zsh 5.9
- Terminal: iTerm2 3.4.19
- Savanhi Shell: v1.0.0

## Logs

```
[Paste relevant logs here]
```

## Additional Context

[Any other context about the problem]
```

## Requesting Features

### Before Requesting

1. **Search existing issues**

   Check if the feature has already been requested.

2. **Check the roadmap**

   See if it's already planned.

### How to Request

Create a new issue with:

1. **Title**: Clear, descriptive title

2. **Description**:
   - What problem does it solve?
   - Proposed solution
   - Alternatives considered

3. **Use cases**:
   - Who would use this?
   - How would it be used?

### Feature Request Template

```markdown
## Problem Description

[Describe the problem this feature would solve]

## Proposed Solution

[Describe your proposed solution]

## Use Cases

1. As a [user type], I want to [action] so that [benefit].

## Alternatives Considered

[Other solutions you've considered]

## Additional Context

[Any other context or screenshots]
```

## Getting Help

- **GitHub Discussions**: For questions and discussions
- **GitHub Issues**: For bug reports and feature requests
- **Documentation**: Check `docs/` for detailed guides

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT License).

---

Thank you for contributing to Savanhi Shell! 🎉