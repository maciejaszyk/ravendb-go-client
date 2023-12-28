package ravendb

// JSONArrayResult describes server's JSON response to batch command
type JSONArrayResult struct {
	Results          []map[string]interface{} `json:"Results"`
	TransactionIndex int64                    `json:"TransactionIndex"`
}

func (r *JSONArrayResult) getResults() []map[string]interface{} {
	return r.Results
}
