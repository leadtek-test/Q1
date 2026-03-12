package decorator

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/sirupsen/logrus"
)

type fakeCommandHandler struct {
	handleFn func(context.Context, sampleCommand) (int, error)
}

func (f fakeCommandHandler) Handle(ctx context.Context, cmd sampleCommand) (int, error) {
	return f.handleFn(ctx, cmd)
}

type fakeQueryHandler struct {
	handleFn func(context.Context, sampleQuery) (string, error)
}

func (f fakeQueryHandler) Handle(ctx context.Context, query sampleQuery) (string, error) {
	return f.handleFn(ctx, query)
}

type sampleCommand struct{ Name string }
type sampleQuery struct{ Name string }

type fakeMetrics struct {
	store map[string]int
}

func (f *fakeMetrics) Inc(key string, value int) {
	if f.store == nil {
		f.store = map[string]int{}
	}
	f.store[key] += value
}

func TestApplyDecorators(t *testing.T) {
	logger := logrus.New()

	cmd := ApplyCommandDecorators[sampleCommand, int](fakeCommandHandler{
		handleFn: func(context.Context, sampleCommand) (int, error) { return 7, nil },
	}, logger)

	got, err := cmd.Handle(context.Background(), sampleCommand{Name: "x"})
	if err != nil || got != 7 {
		t.Fatalf("unexpected command result: got=%d err=%v", got, err)
	}

	query := ApplyQueryDecorators[sampleQuery, string](fakeQueryHandler{
		handleFn: func(context.Context, sampleQuery) (string, error) { return "ok", nil },
	}, logger)

	s, err := query.Handle(context.Background(), sampleQuery{Name: "q"})
	if err != nil || s != "ok" {
		t.Fatalf("unexpected query result: got=%s err=%v", s, err)
	}
}

func TestMetricsDecorators(t *testing.T) {
	client := &fakeMetrics{}

	cmd := commandMetricsDecorator[sampleCommand, int]{
		base: fakeCommandHandler{
			handleFn: func(context.Context, sampleCommand) (int, error) { return 1, nil },
		},
		client: client,
	}
	if _, err := cmd.Handle(context.Background(), sampleCommand{Name: "X"}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	query := queryMetricsDecorator[sampleQuery, string]{
		base: fakeQueryHandler{
			handleFn: func(context.Context, sampleQuery) (string, error) { return "", stderrors.New("fail") },
		},
		client: client,
	}
	if _, err := query.Handle(context.Background(), sampleQuery{Name: "Y"}); err == nil {
		t.Fatalf("expected error")
	}

	if len(client.store) == 0 {
		t.Fatalf("metrics not recorded")
	}
}

func TestGenerateActionName(t *testing.T) {
	name := generateActionName(sampleCommand{})
	if name != "sampleCommand" {
		t.Fatalf("unexpected action name: %s", name)
	}
}
