package service

type Service interface {
	PayloadsCreate
	PayloadsUpdate
	PayloadsDelete
	PayloadsClear
	PayloadsList

	UsersCreate
	UsersDelete

	ProfileGet

	EventsList
	EventsGet

	DNSRecordsCreate
	DNSRecordsDelete
	DNSRecordsClear
	DNSRecordsList

	HTTPRoutesCreate
	HTTPRoutesUpdate
	HTTPRoutesDelete
	HTTPRoutesClear
	HTTPRoutesList

	AuditRecordsList
	AuditRecordsGet

	AuthContextByAPIToken
	AuthContextByTelegramID
	AuthContextByLarkID
	AuthContextBySlackID
}
