# Security

- Use TLS for Redis connections over untrusted networks. Supply a
  `tls.Config` with the expected trust roots and server name when needed; do
  not set `InsecureSkipVerify` in production.
- Keep Redis and Sentinel credentials in a secret store. Redis and Sentinel
  credentials can differ, so configure the two pairs separately.
- Restrict network access to Redis and Sentinel endpoints. Authentication
  alone does not replace network controls.
- Sentinel mode follows a primary and supports logical database 0 only. It
  does not provide Redis Cluster routing.
- `Client.Raw()` exposes the full go-redis API. Apply application-specific
  command, key, and data validation before passing untrusted input to it.
