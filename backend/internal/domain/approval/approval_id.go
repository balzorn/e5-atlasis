package approval

import (
	"fmt"
	"regexp"
)

var approvalIDPattern = regexp.MustCompile(`^APR[0-9]{5}$`)

func ParseApprovalID(value string) (ID, error) {
	if !approvalIDPattern.MatchString(value) {
		return "", fmt.Errorf("invalid approval ID")
	}

	return ID(value), nil
}
