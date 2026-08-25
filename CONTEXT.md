# Material Management

This context describes how maintenance material needs are prepared, issued, and reflected in actual material usage.

## Language

**Material Request**:
A document that records materials requested for one maintenance instance and sector.
_Avoid_: Material order, requisition

**Draft Material Request**:
A Material Request that has not yet been assigned a Request Number and may still be edited.
_Avoid_: Unnumbered request

**Issued Material Request**:
A Material Request that has been assigned a Request Number and can no longer be edited.
_Avoid_: Numbered request, completed request

**Request Number**:
The next consecutive number assigned by the system within one maintenance instance when a Material Request is issued.
_Avoid_: Order number, material request count

**Material Profile Reality**:
The accumulated actual material quantities recorded against a Material Profile.
_Avoid_: Actual inventory

**Material Profile Estimate**:
The current estimated material quantities recorded against a Material Profile.
_Avoid_: Material Profile Reality, imported sheet

**Material Profile**:
A maintenance-specific occurrence of Equipment/Machinery at one Sector and Index Path, with estimated and actual material quantities.
_Avoid_: Equipment record, material master

**Index Path**:
The hierarchical position that distinguishes a Material Profile from other occurrences of the same Equipment/Machinery.
_Avoid_: Database index, profile ID

**Estimate Import**:
A spreadsheet patch that updates matching entries in Material Profile Estimate while preserving entries absent from the spreadsheet.
_Avoid_: Catalogue replacement, full snapshot

**Estimate Variance**:
The non-blocking difference between requested or actual material quantity and Material Profile Estimate.
_Avoid_: Quantity limit, validation error

**User Session**:
One authenticated login on one device, independently refreshable and revocable from other logins by the same user.
_Avoid_: User account, access token

**Requester**:
The user who creates and owns a Material Request.
_Avoid_: Approver, issuer

**Signature Copy**:
A DOCX rendering of a Draft Material Request exported for external approval and signing before issuance.
_Avoid_: Issued document, final request

## Relationships

- A **Material Request** belongs to exactly one maintenance instance and one sector.
- Multiple **Material Profiles** may reference the same Equipment/Machinery and are distinguished by their Sector and **Index Path**.
- An **Index Path** has one to ten hierarchical segments, each in the inclusive range 1-63.
- An **Index Path** identifies at most one **Material Profile** within the same maintenance instance and Sector.
- An **Estimate Import** matches **Material Profiles** by maintenance instance, Sector, and **Index Path**.
- An **Estimate Import** replaces estimated material entries present in the spreadsheet and leaves all other entries unchanged.
- An **Estimate Import** fails if an existing **Index Path** names different Equipment/Machinery, and no part of that import is applied.
- An **Estimate Import** creates missing Equipment/Machinery automatically as part of the same atomic import.
- Each **Material Profile Reality** belongs to exactly one **Material Profile**.
- A **Material Request** may contain only materials that already exist in its referenced **Material Profiles**.
- Creating a missing material is a separate workflow that must complete before it can be selected for a **Material Request**.
- A requested quantity may exceed **Material Profile Estimate**; the frontend presents the **Estimate Variance** for future estimate revisions instead of rejecting the request.
- A **Draft Material Request** keeps its original maintenance instance and Sector; only its description, selected materials, and requested quantities may be edited.
- Logging out revokes only the current **User Session** and leaves other sessions for the same user active.
- Refreshing a **User Session** rotates its refresh token atomically and immediately invalidates the previous refresh token.
- Reuse of an invalidated refresh token is rejected without revoking the current **User Session** or its successor refresh token.
- The refresh token for a **User Session** is stored only in an HttpOnly, Secure, SameSite cookie and is not exposed to frontend JavaScript.
- A **Draft Material Request** becomes an **Issued Material Request** when exactly one **Request Number** is assigned.
- Each maintenance instance owns an independent consecutive sequence of **Request Numbers**.
- An **Issued Material Request** is immutable.
- An **Issued Material Request** cannot be cancelled or withdrawn.
- Issuing a **Material Request** contributes its material quantities to **Material Profile Reality**.
- Until digital approval is supported, only the **Requester** may issue their own **Draft Material Request**.
- A **Draft Material Request** may be exported as a **Signature Copy** before it is issued.
- Exporting a **Signature Copy** does not freeze the **Draft Material Request** or change its lifecycle state.

## Example dialogue

> **Dev:** "Can a user edit a **Material Request** after assigning its **Request Number**?"
> **Domain expert:** "No. Assigning the **Request Number** issues it as an **Issued Material Request**, so it is immutable."

## Flagged ambiguities

- "Cancel" applies only to a **Draft Material Request**; returning unused materials after maintenance is a separate future lifecycle outside the current scope.
- Management approval is required by the real-world process but is temporarily represented by self-issuance because the software does not yet support direct digital approval or signing.
- A **Signature Copy** is intentionally not version-bound to its **Draft Material Request** in the current workflow.
