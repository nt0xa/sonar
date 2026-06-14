package cmd

import "context"

// resultSink carries a command's result out of cobra's RunE back to Exec.
type resultSink struct{ out any }

type sinkKeyType struct{}

var sinkKey = sinkKeyType{}

func withSink(ctx context.Context, s *resultSink) context.Context {
	return context.WithValue(ctx, sinkKey, s)
}

// setResult publishes a command's result. Called at the end of every leaf run.
func setResult(ctx context.Context, out any) error {
	if s, ok := ctx.Value(sinkKey).(*resultSink); ok {
		s.out = out
	}
	return nil
}
