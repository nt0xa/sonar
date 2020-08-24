package actions

type MessageResult struct {
	Message string
}

type Actions interface {
	PayloadsActions
	UsersActions
	DNSActions
}
