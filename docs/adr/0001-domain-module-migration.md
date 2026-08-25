# ADR-0001: Migrate to domain modules

Status: Accepted

## Context

The original backend split each operation across handler, `*Service`, repository interface and a single Mongo adapter. Lifecycle and transaction invariants were therefore spread across shallow modules, while the frontend duplicated Material Request editor state and auth persisted refresh credentials in browser storage.

## Decision

- Each business domain owns a deep module whose Interface expresses complete operations and invariants.
- Material Request issuance is one atomic Interface: validate the Requester, allocate the maintenance-scoped Request Number, accumulate Material Profile Reality and transition Draft to Issued.
- Material Profile uses `(maintenance instance, Sector, Index Path)` as its business key. Estimate Import parses first and applies one transaction.
- User Session refresh credentials are opaque `sessionID.secret` values stored only as hashes and rotated with compare-and-set.
- HTTP and Mongo implementations live with the domain. A seam is retained only when tests need to control behaviour or when multiple adapters exist.
- Backend and frontend ship together; the caller-supplied Request Number Interface is removed.

## Consequences

MongoDB must run as a replica set. Deployment runs the read-only migration preflight before any write, backs up the database, then applies the idempotent migration. Existing access and refresh tokens are invalid after deployment.
