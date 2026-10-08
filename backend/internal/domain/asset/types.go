package asset

type AssetType string

const (
	AssetTypeInformationSystem               AssetType = "information_system"
	AssetTypeInformationInfrastructureObject AssetType = "information_infrastructure_object"
)

type AssetStatus string

const (
	AssetStatusDraft     AssetStatus = "DRAFT"
	AssetStatusActive    AssetStatus = "ACTIVE"
	AssetStatusSuspended AssetStatus = "SUSPENDED"
	AssetStatusRetired   AssetStatus = "RETIRED"
)

type Criticality string

const (
	CriticalityLow      Criticality = "LOW"
	CriticalityMedium   Criticality = "MEDIUM"
	CriticalityHigh     Criticality = "HIGH"
	CriticalityCritical Criticality = "CRITICAL"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

type ProtectionStatus string

const (
	ProtectionStatusNotRequired ProtectionStatus = "NOT_REQUIRED"
	ProtectionStatusRequired    ProtectionStatus = "REQUIRED"
	ProtectionStatusInProgress  ProtectionStatus = "IN_PROGRESS"
	ProtectionStatusImplemented ProtectionStatus = "IMPLEMENTED"
)

type AttestationStatus string

const (
	AttestationStatusNotRequired AttestationStatus = "NOT_REQUIRED"
	AttestationStatusRequired    AttestationStatus = "REQUIRED"
	AttestationStatusInProgress  AttestationStatus = "IN_PROGRESS"
	AttestationStatusAttested    AttestationStatus = "ATTESTED"
	AttestationStatusExpired     AttestationStatus = "EXPIRED"
)

type SecurityProfile struct {
	ProtectionRequired  bool
	ProtectionStatus    ProtectionStatus
	AttestationStatus   AttestationStatus
	CyberCenterRequired bool
}
