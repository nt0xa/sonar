package results

import (
	"context"
	"encoding/json"

	"github.com/russtone/sonar/internal/actions"
)

type JSON struct {
	Encoder *json.Encoder
}

func (h *JSON) OnResult(ctx context.Context, res actions.Result) {
	h.Encoder.Encode(res)
}

