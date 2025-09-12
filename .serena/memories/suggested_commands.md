# Suggested Commands for WFunc Go Development

## Build and Test
```bash
# Run full test suite with coverage
./build.sh

# Sync dependencies
./sync.sh

# Run specific package tests
go test -v github.com/wfunc/go/session
go test -v github.com/wfunc/go/sms
go test -v github.com/wfunc/go/email
go test -v github.com/wfunc/go/baseapi
go test -v github.com/wfunc/go/basedb
```

## Development Commands
```bash
# Install dependencies
go mod download

# Update dependencies
go mod tidy

# Build specific package
go build ./session
go build ./sms
go build ./email

# Run bot server
go run bot/botserver/main.go

# Run markdown converter
go run item2md/item2md.go <title> <file>

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Database Commands
```bash
# Run database migrations (from baseupgrade/)
cd baseupgrade
./dump.sh

# Apply latest SQL
psql -f latest.sql
psql -f clear.sql  # Clear database
```

## Git Commands
```bash
# Standard git workflow
git status
git add .
git commit -m "feat: description"
git push origin main
```

## Go Module Commands
```bash
# Initialize module
go mod init

# Vendor dependencies
go mod vendor

# Verify dependencies
go mod verify
```