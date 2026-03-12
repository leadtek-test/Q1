package logging

import (
	"context"
	"strings"
	"time"

	"github.com/leadtek-test/q1/common/util"
	"github.com/sirupsen/logrus"
)

func WhenPostgres(ctx context.Context, method string, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		Method: method,
		Args:   formatArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "postgres_success"
		fields[Cost] = time.Since(start).Milliseconds()
		fields[Response] = resp

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "postgres_error"
			fields[Error] = (*err).Error()
		}

		logf(ctx, level, fields, "%s", msg)
	}
}

func formatArgs(args []any) string {
	var item []string
	for _, arg := range args {
		item = append(item, formatArg(arg))
	}
	return strings.Join(item, "||")
}

func formatArg(arg any) string {
	var (
		str string
		err error
	)
	defer func() {
		if err != nil {
			str = "unsupported type in formatPostgresArg||err=" + err.Error()
		}
	}()
	switch v := arg.(type) {
	default:
		str, err = util.MarshalString(v)
	case ArgFormatter:
		str, err = v.FormatArg()
	}
	return str
}
