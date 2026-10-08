# Testing

## Backend

Run all tests:

```bash
cd backend
go test ./...
```

## Rules
- New domain behavior must have unit tests.
- New application use cases must have unit tests.
- Tests must cover both successful and invalid scenarios.
- Business invariants must not be tested only through infrastructure tests.
