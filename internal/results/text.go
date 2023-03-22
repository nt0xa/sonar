package results

import (
	"bytes"
	"context"
	"fmt"

	"github.com/russtone/sonar/internal/actions"
)

type Text struct {
	Templates map[string]Template
	OnText    func(ctx context.Context, id, message string)
}

func (h *Text) OnResult(ctx context.Context, res actions.Result) {
	tpl, ok := h.Templates[res.ResultID()]
	if !ok {
		h.OnText(ctx, actions.ErrorResultID, fmt.Sprintf("no template for %q", res.ResultID()))
		return
	}

	buf := &bytes.Buffer{}

	if err := tpl.Execute(buf, res); err != nil {
		h.OnText(ctx, actions.ErrorResultID, fmt.Sprintf("template error for %q: %v", res.ResultID(), err))
		return
	}

	h.OnText(ctx, res.ResultID(), buf.String())
}
