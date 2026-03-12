package logging

import (
	"context"
	stderrors "errors"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

type formatterArg struct {
	value string
	err   error
}

func (f formatterArg) FormatArg() (string, error) {
	return f.value, f.err
}

func TestFormatArgAndArgs(t *testing.T) {
	v := formatArg(map[string]int{"a": 1})
	if v == "" {
		t.Fatalf("formatArg should not be empty")
	}

	v = formatArg(formatterArg{value: "ok"})
	if v != "ok" {
		t.Fatalf("unexpected formatter value: %s", v)
	}

	v = formatArg(formatterArg{err: stderrors.New("x")})
	if v != "" {
		t.Fatalf("expected current behavior empty value, got: %s", v)
	}

	joined := formatArgs([]any{"a", 1})
	if joined == "" {
		t.Fatalf("joined args should not be empty")
	}
}

func TestWhenPostgresAndFormatter(t *testing.T) {
	logger := logrus.New()
	SetFormatter(logger)
	if logger.Formatter == nil {
		t.Fatalf("formatter should be set")
	}

	_ = os.Setenv("LOCAL_ENV", "false")
	SetFormatter(logger)
	if logger.Formatter == nil {
		t.Fatalf("formatter should be set")
	}

	fields, done := WhenPostgres(context.Background(), "m", "x")
	if fields[Method] != "m" {
		t.Fatalf("unexpected fields: %+v", fields)
	}

	var err error
	done("resp", &err)
	err = stderrors.New("fail")
	done(nil, &err)
}
