# Contributing Guide

Thank you for contributing to this project.

We use **Conventional Commits** for all commit messages.

---

## Commit Format

`<type>(optional scope): <description>`

---

## Types

- `feat` – new feature
- `fix` – bug fix
- `ref` – code change that does not add feature or fix bug
- `perf` – performance improvement
- `test` – adding or fixing tests
- `docs` – documentation only changes
- `build` – build system or dependencies
- `ci` – CI configuration changes
- `chore` – maintenance tasks
- `revert` – revert previous commit

---

## Examples

- **feat(proxy):** add request interception middleware
- **fix(tunnel):** resolve connection leak in TCP handler
- **refactor(core):** simplify event dispatcher logic
- **docs(readme):** improve setup instructions
- **test(router):** add unit tests for routing layer

---

## Breaking Changes

Use `!` or `BREAKING CHANGE`:

- **feat(api)!:** change authentication flow
- **BREAKING CHANGE:** token format changed from JWT v1 to v2

---

## Rules

- Use present tense ("add", not "added")
- Keep first line ≤ 72 characters
- No period at the end
- Be specific and technical

---

## Go-specific notes

- Mention package when relevant: `feat(broker): ...`
- Prefer small commits aligned with Go packages
- Avoid mixing formatting changes with logic changes
