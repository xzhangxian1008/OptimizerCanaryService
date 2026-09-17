package diagnosis

import (
	"context"
	"fmt"
)

const (
	sampleCount = 3
)

type Validator struct {
	repository Repository
}

func NewValidator(repository Repository) *Validator {
	return &Validator{repository: repository}
}

func (v *Validator) Validate(ctx context.Context) ValidateResponse {
	response := ValidateResponse{
		Status:  "success",
		Sources: make(map[Source]SourceResult, len(sources)),
	}
	for _, source := range sources {
		response.Sources[source] = SourceResult{}
	}

	for _, source := range sources {
		result := SourceResult{}
		samples, err := v.repository.Sample(ctx, source, sampleCount)
		if err != nil {
			response.Status = "failed"
			response.Reason = fmt.Sprintf("%s sampling failed: %v", source, err)
			return response
		}

		result.Sampled = len(samples)
		for _, sample := range samples {
			if err := v.repository.Explain(ctx, sample); err != nil {
				response.Status = "failed"
				response.Reason = fmt.Sprintf("%s EXPLAIN failed after %d of %d samples: %v", source, result.Explained, result.Sampled, err)
				response.Sources[source] = result
				return response
			}
			result.Explained++
		}
		response.Sources[source] = result

		if result.Sampled < sampleCount {
			response.Status = "failed"
			response.Reason = fmt.Sprintf("%s has fewer than %d valid unredacted queries", source, sampleCount)
			return response
		}
	}
	return response
}
