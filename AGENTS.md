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
