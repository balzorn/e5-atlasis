package asset

import (
	"fmt"
	"regexp"
)

var assetIDPattern = regexp.MustCompile(`^IA[0-9]{5}$`)

type AssetID string

func ParseAssetID(value string) (AssetID, error) {
	if !assetIDPattern.MatchString(value) {
		return "", fmt.Errorf("invalid asset ID %q: expected format IA00000", value)
	}

	return AssetID(value), nil
}

func (id AssetID) String() string {
	return string(id)
}
