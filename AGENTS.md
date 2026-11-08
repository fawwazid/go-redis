# Agent instructions

This repository is the `go-redis` Go module. Keep changes within this module
unless a task explicitly spans other repositories. Preserve existing user
changes. Do not create or edit changelogs; Release Please generates them.

## Security and API

- Preserve certificate verification; never add insecure TLS or plaintext
  fallback behavior.
- Keep Redis and Sentinel authentication separate and do not log credentials.
- Sentinel mode must reject logical databases other than 0. This module does
  not implement Redis Cluster routing.
- Preserve `MaxActiveConns` behavior: zero is unlimited, and go-redis pipeline
  connections use a separate pool.
- `Raw` intentionally exposes go-redis commands; document the boundary and
  preserve input validation in package-owned methods.

## Development

- Target the Go version in `go.mod`; format Go changes with `gofmt`.
- Keep README and security documentation aligned with the exported API.
- Preserve exported APIs unless a breaking change is explicitly requested.
- Add regression coverage for code changes and run the module checks before
  finishing.
- Use Conventional Commit titles for release automation. Do not edit the
  generated changelog manually.

## Branches and pull requests

- Start from the repository's default branch and create a branch for each
  change. Do not commit directly to the default branch.
- Use lowercase type prefixes and concise kebab-case descriptions for branch
  names: `feat/<description>`, `fix/<description>`, `docs/<description>`,
  `refactor/<description>`, `test/<description>`, `chore/<description>`, or
  `security/<description>`.
- Follow Conventional Commits for commit messages:
  `<type>(<optional-scope>): <imperative summary>`. Keep the summary concise,
  lowercase, and without a final period. Use `!` before the colon and a
  `BREAKING CHANGE:` footer for breaking changes.
- Open a pull request for each branch. Use a Conventional Commit title so
  squash merges and release automation produce the intended version change.
  Follow the release types and exceptions described above in this file.
- Describe what changed and why, note user or API impact, and report relevant
  checks and their results. Mention security, compatibility, or migration
  considerations when applicable, and link related issues when available.
- Keep pull requests focused and satisfy the repository's required checks and
  review policy before merging.
