package types

type DataPoint struct {
	Timestamp     float64            `json:"timestamp"`
	Data          map[string]float64 `json:"data"` // All columns as key-value pairs
	ParticipantID string             `json:"participant_id"`
	Condition     string             `json:"condition"`
}

type Dataset struct {
	Points   []DataPoint            `json:"points"`
	Columns  []string               `json:"columns"`
	Metadata map[string]interface{} `json:"metadata"`
}
