package diagnosis

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	samples      []Sample
	sampleErr    error
	explainErrAt int
	explainCalls int
}

func (f *fakeRepository) Sample(_ context.Context, _ Source, limit int) ([]Sample, error) {
	if f.sampleErr != nil {
		return nil, f.sampleErr
	}
	samples := f.samples
	if len(samples) > limit {
		samples = samples[:limit]
	}
	return samples, nil
}

func (f *fakeRepository) Explain(_ context.Context, _ Sample) error {
	f.explainCalls++
	if f.explainErrAt > 0 && f.explainCalls == f.explainErrAt {
		return errors.New("explain error")
	}
	return nil
}

func TestValidatorSuccess(t *testing.T) {
	repo := &fakeRepository{samples: []Sample{{SQL: "select 1"}, {SQL: "select 2"}, {SQL: "select 3"}}}

	response := NewValidator(repo).Validate(context.Background())

	if response.Status != "success" {
		t.Fatalf("expected success, got %+v", response)
	}
	if repo.explainCalls != 9 {
		t.Fatalf("expected 9 EXPLAIN calls, got %d", repo.explainCalls)
	}
}

func TestValidatorFailsWhenSamplesAreInsufficient(t *testing.T) {
	repo := &fakeRepository{samples: []Sample{{SQL: "select 1"}}}

	response := NewValidator(repo).Validate(context.Background())

	if response.Status != "failed" || response.Sources[SourceSlowQuery].Sampled != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestValidatorReportsSamplingFailure(t *testing.T) {
	repo := &fakeRepository{sampleErr: errors.New("unavailable")}

	response := NewValidator(repo).Validate(context.Background())

	if response.Status != "failed" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestValidatorReportsExplainFailure(t *testing.T) {
	repo := &fakeRepository{
		samples:      []Sample{{SQL: "select 1"}, {SQL: "select 2"}, {SQL: "select 3"}},
		explainErrAt: 2,
	}

	response := NewValidator(repo).Validate(context.Background())

	if response.Status != "failed" || response.Sources[SourceSlowQuery].Explained != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
}
