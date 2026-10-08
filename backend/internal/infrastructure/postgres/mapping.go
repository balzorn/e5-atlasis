package postgres

import (
	"encoding/json"
	"fmt"

	domainasset "github.com/balzorn/e5-atlasis/backend/internal/domain/asset"
)

type storedFieldValue struct {
	Kind  string          `json:"kind"`
	Value json.RawMessage `json:"value"`
}

func encodeFieldValue(
	value domainasset.FieldValue,
) ([]byte, error) {
	var raw json.RawMessage

	switch value.Kind() {
	case domainasset.FieldValueKindString:
		v, err := value.StringValue()
		if err != nil {
			return nil, err
		}

		raw, err = json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal string field value: %w", err)
		}

	case domainasset.FieldValueKindBoolean:
		v, err := value.BoolValue()
		if err != nil {
			return nil, err
		}

		raw, err = json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal boolean field value: %w", err)
		}

	default:
		return nil, fmt.Errorf(
			"unsupported field value kind %q",
			value.Kind(),
		)
	}

	return json.Marshal(storedFieldValue{
		Kind:  string(value.Kind()),
		Value: raw,
	})
}

func decodeFieldValue(
	data []byte,
) (domainasset.FieldValue, error) {
	var stored storedFieldValue

	if err := json.Unmarshal(data, &stored); err != nil {
		return domainasset.FieldValue{}, fmt.Errorf(
			"unmarshal field value: %w",
			err,
		)
	}

	switch domainasset.FieldValueKind(stored.Kind) {
	case domainasset.FieldValueKindString:
		var value string

		if err := json.Unmarshal(stored.Value, &value); err != nil {
			return domainasset.FieldValue{}, fmt.Errorf(
				"decode string field value: %w",
				err,
			)
		}

		return domainasset.NewStringFieldValue(value), nil

	case domainasset.FieldValueKindBoolean:
		var value bool

		if err := json.Unmarshal(stored.Value, &value); err != nil {
			return domainasset.FieldValue{}, fmt.Errorf(
				"decode boolean field value: %w",
				err,
			)
		}

		return domainasset.NewBoolFieldValue(value), nil

	default:
		return domainasset.FieldValue{}, fmt.Errorf(
			"unsupported field value kind %q",
			stored.Kind,
		)
	}
}
