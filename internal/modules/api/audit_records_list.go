package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) AuditRecordsList(w http.ResponseWriter, r *http.Request) {
	actorID, err := queryInt64Ptr(r, "actorId")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	from, err := queryTimePtr(r, "from")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	to, err := queryTimePtr(r, "to")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	records, err := api.svc.AuditRecordsList(r.Context(), service.AuditRecordsListInput{
		ActorID:      actorID,
		ActorName:    r.URL.Query().Get("actorName"),
		ResourceType: service.AuditResourceType(r.URL.Query().Get("resourceType")),
		Action:       service.AuditAction(r.URL.Query().Get("action")),
		From:         from,
		To:           to,
		Page:         queryUint(r, "page"),
		PerPage:      queryUint(r, "perPage"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, records)
}
