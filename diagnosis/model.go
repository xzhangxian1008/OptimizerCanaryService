package diagnosis

type Source string

const (
	SourceSlowQuery        Source = "slow_query"
	SourceTopSQL           Source = "top_sql"
	SourceStatementSummary Source = "statement_summary"
)

// TODO Add SourceSlowQuery in the future, as the slow_query table is not accessible in the cloud environment.
var sources = []Source{SourceStatementSummary}

type SourceResult struct {
	Sampled   int `json:"sampled"`
	Explained int `json:"explained"`
}

type ValidateResponse struct {
	Status  string                  `json:"status"`
	Reason  string                  `json:"reason,omitempty"`
	Sources map[Source]SourceResult `json:"sources"`
}

type Sample struct {
	Schema string
	SQL    string
}
