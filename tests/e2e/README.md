# E2E Tests for Savanhi Shell

These tests verify the complete installation flow end-to-end.

## Prerequisites

- Docker
- Go 1.21+
- Make

## Running Tests

```bash
# Run all E2E tests
make e2e

# Run specific test
go test -v ./tests/e2e/... -run TestInstallFlow

# Run with Docker
make e2e-docker
```

## Test Structure

```
tests/e2e/
├── install_test.go     # Installation flow tests
├── rollback_test.go    # Rollback flow tests
├── preview_test.go     # Preview functionality tests
├── noninteractive_test.go  # Non-interactive mode tests
├── docker/             # Docker test environments
│   ├── Dockerfile.ubuntu
│   ├── Dockerfile.arch
│   └── Dockerfile.fedora
└── fixtures/           # Test fixtures
    ├── config.json
    └── themes/
```

## Docker Images

Test environments are defined in `docker/`:

- `Dockerfile.ubuntu` - Ubuntu latest
- `Dockerfile.arch` - Arch Linux
- `Dockerfile.fedora` - Fedora latest