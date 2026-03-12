package errors

import (
	stderrors "errors"
	"strings"
	"testing"

	"github.com/leadtek-test/q1/common/consts"
)

func TestErrorHelpers(t *testing.T) {
	err := New(consts.ErrnoUserNotFound)
	if Errno(err) != consts.ErrnoUserNotFound {
		t.Fatalf("unexpected errno: %d", Errno(err))
	}

	err = NewWithError(consts.ErrnoDatabaseError, stderrors.New("db down"))
	if !strings.Contains(err.Error(), "db down") {
		t.Fatalf("unexpected error message: %s", err.Error())
	}

	err = NewWithMsgf(consts.ErrnoRequestValidateError, "bad %s", "arg")
	if !strings.Contains(err.Error(), "bad arg") {
		t.Fatalf("unexpected message: %s", err.Error())
	}

	errno, msg := Output(nil)
	if errno != consts.ErrnoSuccess || msg == "" {
		t.Fatalf("unexpected output for nil err: %d %s", errno, msg)
	}

	unknown := stderrors.New("unknown")
	errno, msg = Output(unknown)
	if errno != consts.ErrnoUnknownError || msg != "unknown" {
		t.Fatalf("unexpected output for unknown err: %d %s", errno, msg)
	}
}
