package actions

type Actions interface {
	PayloadsActions
	UsersActions
	DNSActions
	UserActions
}

type ResultHandler interface {
	PayloadsHandler
	DNSRecordsHandler
	UsersHandler
	UserHandler
}
