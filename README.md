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

### Running the Examples

The `examples/` directory contains multiple examples demonstrating different features of the library. Each example has its own README with specific instructions.

Available examples:
- **basic** - Basic usage of the library with simple alerts
- **time** - Demonstrates time-based alerts using `WithStartsAt()` and `WithEndsAt()`
- **audit** - Using Alertmanager as an audit log sink for CRUD operations

To run an example:
```bash
# Start Alertmanager (only needs to be done once)
cd examples
docker-compose up -d

# Run any example
cd <example-name>
go run main.go

# Clean up when done
cd ..
docker-compose down
```

See each example's README for detailed information about what it demonstrates.
