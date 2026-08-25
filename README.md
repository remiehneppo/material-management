# Material Management backend

Domain language is defined in [CONTEXT.md](CONTEXT.md); decisions are recorded in [docs/adr](docs/adr).

Run `go run . migrate-domain -c config.yaml` for the read-only deployment preflight. After backing up the database and resolving every conflict, run `go run . migrate-domain -c config.yaml --apply`. The migration refuses to write unless MongoDB is a replica set and the preflight is clean.

`CORS_ALLOWED_ORIGINS` is a comma-separated allowlist. Production must set it explicitly.
