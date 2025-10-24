# alertmanager-client-go

A Go client library for sending alerts to Prometheus Alertmanager.

## Installation

```bash
go get github.com/spectrocloud-labs/alertmanager-client-go
```

## Developer Guide

### Running Tests

```bash
# Run all tests
go test -v ./...
```

### Examples

The `examples/` directory contains multiple examples demonstrating different features of the library:

- **basic** - Basic usage of the library with simple alerts
- **time** - Demonstrates time-based alerts using `WithStartsAt()` and `WithEndsAt()`
- **audit** - Using Alertmanager as an audit log sink for CRUD operations
- **secure** - TLS and basic authentication configuration with security validation

See [examples/README.md](examples/README.md) for detailed information about each example and instructions on how to run them.
