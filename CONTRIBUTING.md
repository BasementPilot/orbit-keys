# Contributing to OrbitKeys

Thank you for considering contributing to OrbitKeys! This document outlines the process for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by the [GitHub Community Guidelines](https://docs.github.com/en/github/site-policy/github-community-guidelines).

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report. Following these guidelines helps maintainers understand your report.

* **Use a clear and descriptive title** for the issue to identify the problem.
* **Describe the exact steps which reproduce the problem** in as much detail as possible.
* **Provide specific examples to demonstrate the steps**.
* **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
* **Explain which behavior you expected to see instead and why.**
* **Include screenshots or animated GIFs** if possible.

### Reporting Security Issues

If you discover a security vulnerability within OrbitKeys, please follow these steps:

* **Do not disclose security vulnerabilities publicly** in the issue tracker.
* **Send an email** to the project maintainers directly.
* **Provide details** of the vulnerability including steps to reproduce.
* **If possible, include suggestions** for fixing the vulnerability.

We take security issues very seriously and will address them promptly.

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion, including completely new features and minor improvements to existing functionality.

* **Use a clear and descriptive title** for the issue to identify the suggestion.
* **Provide a step-by-step description of the suggested enhancement** in as much detail as possible.
* **Provide specific examples to demonstrate the steps** or point to similar features in other libraries.
* **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
* **Explain why this enhancement would be useful** to most OrbitKeys users.

### Pull Requests

* Fill in the required template
* Do not include issue numbers in the PR title
* Follow the Go style guide
* Include tests for new features or bug fixes
* Document new code
* End all files with a newline
* Include security considerations where applicable

## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line
* Consider starting the commit message with an applicable type:
  * `feat:` - new feature
  * `fix:` - bug fix
  * `docs:` - documentation changes
  * `security:` - security improvements
  * `test:` - adding or updating tests
  * `refactor:` - code changes that neither fix bugs nor add features

### Go Styleguide

* Code should be formatted with `gofmt`
* Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
* Use meaningful variable and function names
* Document all exported functions, types, and constants

### Security Guidelines

When contributing, please keep security in mind:

* **Use constants and secure algorithms** for cryptographic operations
* **Implement timeout mechanisms** for long-running operations
* **Use constant-time comparison** for comparing sensitive values like API keys
* **Validate and sanitize all inputs**, especially from environment variables
* **Use proper error messages** that don't leak sensitive information
* **Implement rate limiting** for authentication endpoints
* **Apply the principle of least privilege** when designing permissions
* **Include security tests** where applicable

## Running Tests

```bash
go test ./...
```

Make sure to write tests for security-critical functionality.

## Documentation

* Update documentation with any changes to the API
* Use godoc format for comments
* Document security considerations in code comments

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License. 