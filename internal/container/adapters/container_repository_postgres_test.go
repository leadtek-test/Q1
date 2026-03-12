package adapters

import (
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
)

func TestNormalizeAndMarshalHelpers(t *testing.T) {
	command, err := normalizeAndMarshalCommand(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if command != "[]" {
		t.Fatalf("unexpected command json: %s", command)
	}

	env, err := normalizeAndMarshalEnv(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != "{}" {
		t.Fatalf("unexpected env json: %s", env)
	}
}

func TestToDomainContainer(t *testing.T) {
	model := persistent.ContainerModel{
		UserID:    1,
		Name:      "n",
		Image:     "img",
		Command:   `["a","b"]`,
		Env:       `{"k":"v"}`,
		RuntimeID: "rid",
		Status:    string(domaincontainer.StatusRunning),
	}
	model.ID = 9
	model.CreatedAt = time.Unix(1, 0)
	model.UpdatedAt = time.Unix(2, 0)

	result, err := toDomainContainer(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 9 || len(result.Command) != 2 || result.Env["k"] != "v" {
		t.Fatalf("unexpected result: %+v", result)
	}

	_, err = toDomainContainer(persistent.ContainerModel{Command: "{bad json}"})
	if got := commonerrors.Errno(err); got != consts.ErrnoDatabaseError {
		t.Fatalf("unexpected errno: %d", got)
	}
}
