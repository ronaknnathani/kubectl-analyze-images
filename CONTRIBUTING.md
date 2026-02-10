# Contributing to kubectl-analyze-images

Thank you for your interest in contributing to kubectl-analyze-images! This document provides guidelines and instructions for contributing.

## Prerequisites

- **Go 1.23+** -- this project uses Go 1.23 features
- **kubectl** -- for testing against a Kubernetes cluster
- **Access to a Kubernetes cluster** -- for integration testing (minikube, kind, or a remote cluster)
- **golangci-lint** -- for code linting (install via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

## Development Setup

1. **Fork and clone the repository:**

   ```bash
   git clone https://github.com/<your-username>/kubectl-analyze-images.git
   cd kubectl-analyze-images
   ```

2. **Install dependencies:**

   ```bash
   make deps
   ```

3. **Build the plugin:**

   ```bash
   make build
   ```

4. **Run the tests:**

   ```bash
   make test
   ```

## Making Changes

1. **Fork the repository** on GitHub
2. **Create a feature branch** from `main`:

   ```bash
   git checkout -b feature/my-new-feature
   ```

3. **Make your changes** -- write clear, well-documented code
4. **Add tests** for any new functionality
5. **Run all checks** to ensure nothing is broken:

   ```bash
   make check
   ```

   This runs both `make test` (unit tests) and `make lint` (linter).

6. **Commit your changes** with a descriptive message:

   ```bash
   git commit -m "feat: add support for image age analysis"
   ```

7. **Push to your fork:**

   ```bash
   git push origin feature/my-new-feature
   ```

8. **Open a Pull Request** against the `main` branch

## Code Style

This project uses [golangci-lint](https://golangci-lint.run/) for code quality enforcement. The linter configuration is defined in `.golangci.yml`.

Before submitting a pull request, ensure your code passes the linter:

```bash
make lint
```

Key style conventions:

- Follow standard Go conventions and the [Effective Go](https://go.dev/doc/effective_go) guide
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and reasonably sized

## Testing

Run unit tests:

```bash
make test
```

Run tests with coverage report:

```bash
make test-coverage
```

Testing guidelines:

- **Target >70% coverage** for new code
- Use table-driven tests where appropriate
- Use the `testify` assertion library (already a project dependency)
- Place test files alongside the code they test (`*_test.go`)
- Use the fake Kubernetes client (`fake.NewClientset()`) for cluster interaction tests

## Pull Request Guidelines

When submitting a pull request, please ensure:

- **Clear description** -- explain what the change does and why
- **Test coverage** -- include tests for new functionality
- **One feature per PR** -- keep pull requests focused on a single change
- **Reference related issues** -- link to any relevant GitHub issues
- **Passing CI** -- all GitHub Actions checks must pass (tests, lint, build)
- **Clean commits** -- squash fixup commits before requesting review

## Reporting Issues

If you find a bug or have a feature request:

1. **Check existing issues** to avoid duplicates
2. **Open a new issue** with a clear title and description
3. **Include reproduction steps** for bugs
4. **Provide context** -- Go version, Kubernetes version, OS, etc.

## License

By contributing to kubectl-analyze-images, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
