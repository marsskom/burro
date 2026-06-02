# Collaboration Guidelines

## License

This project is licensed under the **GNU General Public License v3.0 (GPLv3)**.

By contributing, you agree that your contributions will be licensed under GPLv3 as well.

See full license: https://www.gnu.org/licenses/gpl-3.0.en.html

---

## Contribution Principles

We value:

- Clean, readable Go code
- Minimal dependencies
- Explicit error handling
- Test coverage for new functionality

---

## Pull Request Rules

Before submitting a merge request:

- Ensure code compiles with latest Go version
- Run tests:

```shell
  go test ./...
```

- Run formatting:

```shell
go fmt ./...
```

- Keep changes focused and atomic

---

## Code Ownership

- No forced CLA (Contributor License Agreement)
- Contributions are accepted under GPLv3 automatically
- Maintainers may refactor submitted code for consistency

---

## Dependencies Policy

- Prefer standard library
- External dependencies must be:
  - Actively maintained
  - Compatible with GPLv3
  - Justified in PR description

---

## Security

Do not introduce:

- Hardcoded secrets
- Unsafe deserialization patterns
- Arbitrary remote code execution risks
- Unsafe system command execution without validation

Report vulnerabilities privately via GitLab security issues.

---

## Responsibility

Contributors are expected to:

- Write clear, readable code
- Respect project architecture decisions
- Communicate clearly in issues and merge requests
- Be open to feedback and iteration

---

## Review Process

All changes require:

- At least 1 maintainer approval
- Passing CI pipeline
