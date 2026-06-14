package dbsvc

import (
	"encoding/base64"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

func httpRoute(m database.HTTPRoute, payloadSubdomain string) *service.HTTPRoute {
	return &service.HTTPRoute{
		Index:            int64(m.Index),
		PayloadSubdomain: payloadSubdomain,
		Method:           service.HTTPMethod(m.Method),
		Path:             m.Path,
		Code:             m.Code,
		Headers:          m.Headers,
		Body:             base64.StdEncoding.EncodeToString(m.Body),
		IsDynamic:        m.IsDynamic,
		CreatedAt:        m.CreatedAt,
	}
}
