package service

import (
	"encoding/base64"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

func httpRoute(m database.HTTPRoute, payloadSubdomain string) *types.HTTPRoute {
	return &types.HTTPRoute{
		Index:            int64(m.Index),
		PayloadSubdomain: payloadSubdomain,
		Method:           m.Method,
		Path:             m.Path,
		Code:             m.Code,
		Headers:          m.Headers,
		Body:             base64.StdEncoding.EncodeToString(m.Body),
		IsDynamic:        m.IsDynamic,
		CreatedAt:        m.CreatedAt,
	}
}
