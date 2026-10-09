# Domain Glossary

This glossary includes target-domain terminology. The current implementation still contains legacy names such as InformationAsset and IAxxxxx; these are not the target identity model.

| Term | Definition |
|---|---|
| Tenant | Top-level data-isolation and authorization boundary for an independently administered registry customer/group |
| Organization | Independently identified entity belonging to exactly one tenant |
| Information System (IS) | A separately managed registry entity representing an information system |
| Object of Information Infrastructure (OII) | A separately managed registry entity representing an object of information infrastructure |
| OII composition | Versioned relationship between an OII and the IS included in its composition |
| Information Asset (legacy IA) | Legacy unified domain entity used by the current implementation to represent either an IS or an OII |
| Technical ID | Internal stable database identifier; target new entities use UUIDv7 |
| Display number | Human-readable identifier generated separately from the technical ID; target IS and OII namespaces are scoped by tenant and entity type |
| Subject | Authenticated actor represented by an immutable internal UUID in AtlasIS |
| External identity | Identity-provider identifier mapped to a subject; for OIDC, normally the validated (iss, sub) pair |
| Scoped role assignment | Grant of actions to a subject within an explicit tenant/resource/organization scope |
| Entity Version | Immutable snapshot of an entity's effective state |
| Change Request (CR) | Controlled request to change one target entity; applying it creates a new immutable version |
| Field Change | One proposed field-level change inside a CR |
| Discussion Thread | Contextual discussion associated with a CR or one of its field changes |
| Approval | Explicit decision associated with a CR |
| Owner | Person or role accountable for a registry entity, independent of access grants |
| Criticality | Business significance of an entity |
| Risk Level | Current assessed security risk level |
| Protection Status | State of required information protection measures |
| Attestation Status | State of applicable attestation/certification requirements |
| Cybersecurity Center | Holding cybersecurity center that interacts with the information security function |
