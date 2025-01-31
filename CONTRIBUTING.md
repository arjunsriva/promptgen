# Contributing to promptgen

Thank you for your interest in contributing to promptgen! We welcome contributions from the community and are pleased to have you join us.

## Code of Conduct

By participating in this project, you agree to abide by the code of conduct principles:

- Be respectful and inclusive
- Exercise consideration and empathy
- Focus on what is best for the community
- Give and gracefully accept constructive feedback

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/promptgen.git`
3. Create a new branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes: `git commit -m "Add some feature"`
7. Push to your fork: `git push origin feature/your-feature-name`
8. Open a Pull Request

## Development Setup

1. Ensure you have Go 1.21 or later installed
2. Install dependencies: `go mod download`
3. Set up your OpenAI API key for testing:
   ```bash
   export OPENAI_API_KEY=your-api-key
   ```

## Pre-commit Hooks

We use pre-commit hooks to ensure code quality. To set them up:

1. Install pre-commit:
   ```bash
   # macOS
   brew install pre-commit

   # Linux
   pip install pre-commit

   # Windows
   pip install pre-commit
   ```

2. Install golangci-lint:
   ```bash
   # macOS
   brew install golangci-lint

   # Windows & Linux
   # See: https://golangci-lint.run/usage/install/#local-installation
   ```

3. Set up the pre-commit hooks:
   ```bash
   pre-commit install
   ```

4. (Optional) Run against all files:
   ```bash
   pre-commit run --all-files
   ```

The pre-commit hooks will now run automatically on `git commit`. They include:
- Code formatting (gofmt)
- Import organization (goimports)
- Linting (golangci-lint)
- Unit tests
- Build verification
- Test coverage checks
- And more...

If any checks fail, the commit will be aborted. Fix the issues and try committing again.
