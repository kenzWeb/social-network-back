package services

import "context"

type NoopTxRunner struct{}

func (NoopTxRunner) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
