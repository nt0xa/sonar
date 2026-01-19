package smtpx

// Meta contains SMTP-specific event metadata.
type Meta struct {
	Session Data  `json:"session"`
	Email   Email `json:"email"`
}
