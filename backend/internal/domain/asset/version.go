package asset

import (
	"fmt"
	"time"
)

type Version int

func NewVersion(value int) (Version, error) {
	if value < 1 {
		return 0, fmt.Errorf("version must be greater than zero")
	}

	return Version(value), nil
}

func (v Version) Int() int {
	return int(v)
}

type AssetVersion struct {
	ID              string
	AssetID         AssetID
	Version         Version
	State           InformationAsset
	CreatedBy       string
	CreatedAt       time.Time
	ChangeRequestID *string
}
