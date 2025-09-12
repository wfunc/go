# Code Style and Conventions for WFunc Go

## Naming Conventions
- **Packages**: Lowercase, single word (e.g., `session`, `sms`, `email`)
- **Exported Functions**: PascalCase (e.g., `NewDbSessionBuilder`, `SendMessage`)
- **Unexported Functions**: camelCase (e.g., `newCode`, `writeCookie`)
- **Constants**: PascalCase or UPPER_CASE (e.g., `PhoneCodeTypeVerify`, `INIT`)
- **Variables**: camelCase for local, PascalCase for exported
- **Handlers**: Suffix with 'H' (e.g., `SendVerifySmsH`, `LoadPhoneCodeH`)

## File Organization
- One main type/interface per file
- Test files alongside implementation (`*_test.go`)
- SQL files for database schemas (`*.sql`)
- Shell scripts for automation (`*.sh`)

## Code Patterns
- Handler functions follow pattern: `func NameH(hs *web.Session)`
- Use `*web.Session` for HTTP handlers
- Return error as last value in functions
- Use context.Context as first parameter when needed
- Initialize Redis pools and database connections in `init()` functions

## Testing
- Test files must have `_test.go` suffix
- Use table-driven tests where appropriate
- Aim for >80% code coverage
- Mock external dependencies (Redis, Database, APIs)

## Documentation
- Package-level documentation in main file
- Function documentation before declaration
- Use meaningful variable names
- Comment complex logic

## Error Handling
- Return errors, don't panic
- Wrap errors with context
- Log errors at appropriate levels
- Use custom error types for domain errors

## Dependencies
- Use go.mod for dependency management
- Vendor critical dependencies
- Keep dependencies minimal
- Update regularly but cautiously