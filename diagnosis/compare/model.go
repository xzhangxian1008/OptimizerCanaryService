package compare

// Sample identifies a SQL statement and the schema in which it runs.
type Sample struct {
	Schema string
	SQL    string
}
