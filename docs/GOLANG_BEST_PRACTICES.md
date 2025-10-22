# Golang Best Practices

## 📦 Code Organization
- Follow Clean Architecture
- Keep packages focused with single responsibilities
- Use `internal/` for private code
- Project structure:
  - `cmd/` - application entry points
  - `internal/` - private application code
  - `pkg/` - public libraries

## 🏷️ Naming Conventions
- **CamelCase** for unexported names
- **PascalCase** for exported names
- Short names in short scopes (`i`, `r`)
- Descriptive names in wider scopes (`userRepository`)
- Interfaces: `-er` suffix for single-method (`Reader`, `Writer`)

## ⚠️ Error Handling
- Always handle errors - never ignore them
- Use `fmt.Errorf` with `%w` to wrap and preserve error chains
- Use `errors.Is` and `errors.As` for checking
- Never panic in library code

## ⚡ Concurrency
- Always know when goroutines stop (use context)
- Prevent leaks with proper cancellation
- Use `sync.WaitGroup` for coordination
- Only sender closes channels
- Pass `context.Context` as first parameter

## 🧪 Testing
- Use table-driven tests with `t.Run()`
- Test files end in `_test.go`
- Create mocks via interfaces
- Run coverage: `go test -cover ./...`

## 🚀 Performance
- Preallocate slices with capacity
- Use `strings.Builder` for concatenation
- Leverage `sync.Pool` for reusable objects
- Profile before optimizing: `go test -cpuprofile`

## 🔒 Security
- Always validate input
- Use parameterized queries to prevent SQL injection
- Use `crypto/rand` for secure randomness
- Configure TLS with minimum version 1.3
- Do not hardcoded API key in code. Always set API key in .env and DO NOT push .env to GIT
- AI Coding tools ( Cursor, Claude, Winsurf,...) DO NOT READ .env file

## 💎 Code Style
- Run `gofmt`, `goimports`, and `golangci-lint`
- Use early returns to reduce nesting
- Group imports: stdlib → external → internal
- Make zero values useful

## 📚 Resources
- [Effective Go](https://golang.org/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- Tools: `gofmt`, `golangci-lint`, `go vet`
