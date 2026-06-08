````md
---
inclusion: always
---

# Golang Backend Engineering Standards

## Technology Requirements
- All backend services and applications must use Golang (Go).
- Minimum Go version is 1.22+.
- Do not introduce additional backend languages without architecture approval.

---

# Architecture Rules

## SOLID Principles
All code must follow SOLID principles:
- Single Responsibility Principle
- Open/Closed Principle
- Liskov Substitution Principle
- Interface Segregation Principle
- Dependency Inversion Principle

Requirements:
- Keep components modular and maintainable.
- Avoid tightly coupled implementations.
- Separate business logic from infrastructure concerns.

---

# Dependency Injection Rules

- All dependencies must be injected.
- Constructor injection is the preferred approach.
- Do not instantiate repositories, services, or external clients directly inside business logic.
- Depend on interfaces instead of concrete implementations.

Preferred example:

```go
type UserService struct {
    repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{
        repo: repo,
    }
}
````

---

# Unit Testing Rules

## Mandatory Unit Testing

* Every service, use case, and business logic component must include unit tests.
* All new features must include corresponding unit test coverage.
* Bug fixes must include regression tests.
* External dependencies must be mocked during testing.

## Coverage Requirements

* Minimum unit test coverage target is 80%.
* Critical business logic should have high coverage and edge-case validation.

## Recommended Testing Libraries

* testing
* testify
* gomock

---

# SonarQube Rules

## Mandatory SonarQube Compliance

All projects must pass SonarQube Quality Gate validation.

The following conditions are mandatory:

* No blocker issues
* No critical vulnerabilities
* No unresolved major security hotspots
* No excessive duplicated code
* Coverage must meet project thresholds

## Code Quality Requirements

Code must:

* Handle errors explicitly
* Avoid dead code
* Avoid unused variables
* Avoid overly complex functions
* Avoid hardcoded credentials or secrets
* Use meaningful naming conventions

---

# Project Structure

Recommended structure:

```text
/cmd
/internal
    /handler
    /service
    /repository
    /model
    /config
/pkg
/test
/docs
```

---

# Security Rules

* Never hardcode secrets or credentials.
* Use environment variables for configuration.
* Validate all external input.
* Use structured logging.
* Apply proper timeout and context handling.

