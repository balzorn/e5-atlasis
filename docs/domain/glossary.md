# Domain Glossary

This glossary defines the target domain. The current implementation still uses the legacy unified
`InformationAsset` / `IAxxxxx` model; that implementation detail is not the target domain model.

| Term | Definition |
|---|---|
| Tenant | Top-level data-isolation and authorization boundary for an independently administered registry customer/group |
| Organization | Independently identified entity belonging to exactly one tenant |
| Department | Organizational unit belonging to exactly one organization and tenant |
| Information System (IS) | A separately managed registry entity representing an information system |
| Object of Information Infrastructure (OII) | A separately managed registry entity representing an object of information infrastructure |
| OII composition | Versioned relationship between an OII and an IS included in that OII's effective composition |
| Critically Important Object of Informatization (КВОИ) | A distinct formal regulatory designation; not a synonym for OII and not inferred from internal criticality |
| Object of Informatization | A separate term used in normative documents; not introduced as a third primary AtlasIS entity by the current v1 decision |
| Information Asset (legacy IA) | Legacy unified domain entity used by the current implementation to represent either an IS or an OII |
| Subject | Immutable internal AtlasIS identity for an authenticated person or actor |
| External Identity | Identity-provider identity mapped to an internal subject, normally using validated issuer and subject claims |
| Role Assignment | Audited grant of a role to a subject within an explicit tenant and scope |
| Asset Version | Legacy implementation term; target IS and OII versions are separate immutable snapshots |
| Change Request | Controlled request to change exactly one target IS or OII |
| Approval | Explicit decision required for a Change Request |
| Discussion Thread | Discussion associated with a supported parent resource or Change Request |
| Owner | Organization accountable for the specific IS or OII; distinct from technical administrator and named responsible contacts |
| Criticality | Internal business-significance assessment; does not establish formal КВОИ designation |
| Risk Assessment | Internal assessment with methodology/version, assessment date, rationale and review metadata |
| Protection Status | Status of applicable information-protection measures, distinct from evidence such as an attestation document |
| Cybersecurity Center | Center providing cybersecurity services under the applicable service/contractual model |
