package change

import (
	"context"
	"errors"
	"testing"

	domainchange "github.com/balzorn/e5-atlasis/backend/internal/domain/change"
	"github.com/balzorn/e5-atlasis/backend/internal/ports"
)

type fakeChangeGetRepository struct {
	changeRequest *domainchange.ChangeRequest
	err           error
}

func (r *fakeChangeGetRepository) Create(
	_ context.Context,
	cr domainchange.ChangeRequest,
) error {
	r.changeRequest = &cr
	return nil
}

func (r *fakeChangeGetRepository) GetByID(
	_ context.Context,
	_ domainchange.ID,
) (*domainchange.ChangeRequest, error) {
	return r.changeRequest, r.err
}

func (r *fakeChangeGetRepository) Save(
	_ context.Context,
	_ domainchange.ChangeRequest,
) error {
	return nil
}

func TestGetChangeRequest(t *testing.T) {
	expected := &domainchange.ChangeRequest{
		ID:          "CR00001",
		AssetID:     "IA00001",
		BaseVersion: 1,
		Initiator:   "USR001",
		Title:       "Test change",
	}

	uc := NewGetChangeRequestUseCase(&fakeChangeGetRepository{
		changeRequest: expected,
	})

	got, err := uc.Execute(context.Background(), expected.ID)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got != expected {
		t.Fatal("Execute() returned unexpected change request")
	}
}

func TestGetChangeRequestNotFound(t *testing.T) {
	uc := NewGetChangeRequestUseCase(&fakeChangeGetRepository{})

	_, err := uc.Execute(context.Background(), "CR00001")
	if !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("Execute() error = %v, want ErrNotFound", err)
	}
}
