package adapters

import (
	"context"
	"encoding/json"

	"github.com/leadtek-test/q1/common/consts"
	"github.com/leadtek-test/q1/common/handler/errors"
	"github.com/leadtek-test/q1/container/domain/container"
	"github.com/leadtek-test/q1/container/infrastructure/persistent"
)

type ContainerRepositoryPostgres struct {
	db *persistent.Postgres
}

func NewContainerRepositoryPostgres(db *persistent.Postgres) *ContainerRepositoryPostgres {
	return &ContainerRepositoryPostgres{db: db}
}

func (c ContainerRepositoryPostgres) Create(ctx context.Context, data *container.Container) error {
	commandData, err := normalizeAndMarshalCommand(data.Command)
	if err != nil {
		return err
	}
	envData, err := normalizeAndMarshalEnv(data.Env)
	if err != nil {
		return err
	}

	model := &persistent.ContainerModel{
		UserID:    data.UserID,
		Name:      data.Name,
		Image:     data.Image,
		Command:   commandData,
		Env:       envData,
		RuntimeID: data.RuntimeID,
		Status:    string(data.Status),
	}

	if err = c.db.CreateContainer(ctx, nil, model); err != nil {
		return err
	}

	data.ID = model.ID
	data.CreatedAt = model.CreatedAt
	data.UpdatedAt = model.UpdatedAt
	return nil
}

func (c ContainerRepositoryPostgres) GetByIDAndUser(ctx context.Context, id, userID uint) (container.Container, error) {
	model, err := c.db.GetContainerByIDAndUser(ctx, id, userID)
	if err != nil {
		return container.Container{}, err
	}
	return toDomainContainer(*model)
}

func (c ContainerRepositoryPostgres) Update(ctx context.Context, data *container.Container) error {
	commandData, err := normalizeAndMarshalCommand(data.Command)
	if err != nil {
		return err
	}
	envData, err := normalizeAndMarshalEnv(data.Env)
	if err != nil {
		return err
	}

	model := &persistent.ContainerModel{
		UserID:    data.UserID,
		Name:      data.Name,
		Image:     data.Image,
		Command:   commandData,
		Env:       envData,
		RuntimeID: data.RuntimeID,
		Status:    string(data.Status),
	}
	model.ID = data.ID

	return c.db.UpdateContainer(ctx, nil, model)
}

func (c ContainerRepositoryPostgres) Delete(ctx context.Context, id, userID uint) error {
	return c.db.DeleteContainer(ctx, nil, id, userID)
}

func (c ContainerRepositoryPostgres) ListByUser(ctx context.Context, userID uint) ([]container.Container, error) {
	models, err := c.db.BatchGetContainerByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]container.Container, 0, len(models))
	for _, model := range models {
		item, convErr := toDomainContainer(model)
		if convErr != nil {
			return nil, convErr
		}
		result = append(result, item)
	}
	return result, nil
}

func normalizeAndMarshalCommand(command []string) (string, error) {
	if command == nil {
		command = make([]string, 0)
	}
	data, err := json.Marshal(command)
	if err != nil {
		return "", errors.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return string(data), nil
}

func normalizeAndMarshalEnv(env map[string]string) (string, error) {
	if env == nil {
		env = map[string]string{}
	}
	data, err := json.Marshal(env)
	if err != nil {
		return "", errors.NewWithError(consts.ErrnoDatabaseError, err)
	}
	return string(data), nil
}

func toDomainContainer(model persistent.ContainerModel) (container.Container, error) {
	commandData := make([]string, 0)
	if model.Command != "" {
		if err := json.Unmarshal([]byte(model.Command), &commandData); err != nil {
			return container.Container{}, errors.NewWithError(consts.ErrnoDatabaseError, err)
		}
	}

	envData := map[string]string{}
	if model.Env != "" {
		if err := json.Unmarshal([]byte(model.Env), &envData); err != nil {
			return container.Container{}, errors.NewWithError(consts.ErrnoDatabaseError, err)
		}
	}

	return container.Container{
		ID:        model.ID,
		UserID:    model.UserID,
		Name:      model.Name,
		Image:     model.Image,
		Command:   commandData,
		Env:       envData,
		RuntimeID: model.RuntimeID,
		Status:    container.Status(model.Status),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}
