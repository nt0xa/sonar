package dnsx

// Question represents a DNS question.
type Question struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Answer represents a DNS answer.
type Answer struct {
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  uint32 `json:"ttl"`
}

// Meta contains DNS-specific event metadata.
type Meta struct {
	Question Question `json:"question"`
	Answer   []Answer `json:"answer,omitempty"`
}
