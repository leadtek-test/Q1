package adapters

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leadtek-test/q1/common/consts"
	commonerrors "github.com/leadtek-test/q1/common/handler/errors"
	domaincontainer "github.com/leadtek-test/q1/container/domain/container"
	domainjob "github.com/leadtek-test/q1/container/domain/job"
	"github.com/sirupsen/logrus"
)

type fakeJobRepo struct {
	mu   sync.Mutex
	jobs map[string]domainjob.CreateContainerJob
}

func (f *fakeJobRepo) Create(_ context.Context, job *domainjob.CreateContainerJob) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.jobs == nil {
		f.jobs = map[string]domainjob.CreateContainerJob{}
	}
	f.jobs[job.JobID] = *job
	return nil
}

func (f *fakeJobRepo) GetByJobIDAndUser(_ context.Context, jobID string, userID uint) (domainjob.CreateContainerJob, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	job, ok := f.jobs[jobID]
	if !ok || job.UserID != userID {
		return domainjob.CreateContainerJob{}, commonerrors.New(consts.ErrnoContainerCreateJobNotFound)
	}
	return job, nil
}

func (f *fakeJobRepo) Update(_ context.Context, job *domainjob.CreateContainerJob) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	item, ok := f.jobs[job.JobID]
	if !ok || item.UserID != job.UserID {
		return commonerrors.New(consts.ErrnoContainerCreateJobNotFound)
	}
	item.Status = job.Status
	item.ErrorMessage = job.ErrorMessage
	item.ContainerID = job.ContainerID
	f.jobs[job.JobID] = item
	return nil
}

type fakeContainerRepoForDispatcher struct {
	createFn func(context.Context, *domaincontainer.Container) error
}

func (f fakeContainerRepoForDispatcher) Create(ctx context.Context, c *domaincontainer.Container) error {
	if f.createFn != nil {
		return f.createFn(ctx, c)
	}
	return nil
}
func (f fakeContainerRepoForDispatcher) GetByIDAndUser(context.Context, uint, uint) (domaincontainer.Container, error) {
	return domaincontainer.Container{}, nil
}
func (f fakeContainerRepoForDispatcher) Update(context.Context, *domaincontainer.Container) error {
	return nil
}
func (f fakeContainerRepoForDispatcher) Delete(context.Context, uint, uint) error { return nil }
func (f fakeContainerRepoForDispatcher) ListByUser(context.Context, uint) ([]domaincontainer.Container, error) {
	return nil, nil
}

type fakeContainerRuntimeForDispatcher struct {
	createFn func(context.Context, uint, domaincontainer.CreateSpec, string) (string, error)
}

func (f fakeContainerRuntimeForDispatcher) Create(ctx context.Context, userID uint, spec domaincontainer.CreateSpec, workspacePath string) (string, error) {
	if f.createFn != nil {
		return f.createFn(ctx, userID, spec, workspacePath)
	}
	return "runtime-id", nil
}
func (f fakeContainerRuntimeForDispatcher) Start(context.Context, string) error  { return nil }
func (f fakeContainerRuntimeForDispatcher) Stop(context.Context, string) error   { return nil }
func (f fakeContainerRuntimeForDispatcher) Delete(context.Context, string) error { return nil }

type fakeWorkspaceForDispatcher struct {
	ensureFn func(uint) (string, error)
}

func (f fakeWorkspaceForDispatcher) EnsureUserDir(userID uint) (string, error) {
	if f.ensureFn != nil {
		return f.ensureFn(userID)
	}
	return "/tmp/test", nil
}

func (f fakeWorkspaceForDispatcher) Save(uint, string, []byte) (string, error) {
	return "", nil
}

func TestCreateContainerDispatcherChannelSuccess(t *testing.T) {
	repo := &fakeJobRepo{jobs: map[string]domainjob.CreateContainerJob{}}
	dispatcher := NewCreateContainerDispatcherChannel(
		repo,
		fakeContainerRepoForDispatcher{
			createFn: func(_ context.Context, c *domaincontainer.Container) error {
				c.ID = 9
				return nil
			},
		},
		fakeContainerRuntimeForDispatcher{},
		fakeWorkspaceForDispatcher{},
		8,
		logrus.New(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go dispatcher.Listen(ctx)

	jobID, err := dispatcher.DispatchCreateContainer(context.Background(), domainjob.CreateContainerTask{
		UserID: 1,
		Name:   "demo",
		Image:  "busybox:latest",
	})
	if err != nil {
		t.Fatalf("DispatchCreateContainer unexpected error: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		job, getErr := repo.GetByJobIDAndUser(context.Background(), jobID, 1)
		if getErr != nil {
			t.Fatalf("GetByJobIDAndUser unexpected error: %v", getErr)
		}
		if job.Status == domainjob.CreateContainerJobStatusSucceeded {
			if job.ContainerID != 9 {
				t.Fatalf("unexpected container id: %d", job.ContainerID)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting job done, current status=%s", job.Status)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestCreateContainerDispatcherChannelFailure(t *testing.T) {
	repo := &fakeJobRepo{jobs: map[string]domainjob.CreateContainerJob{}}
	dispatcher := NewCreateContainerDispatcherChannel(
		repo,
		fakeContainerRepoForDispatcher{
			createFn: func(context.Context, *domaincontainer.Container) error {
				return nil
			},
		},
		fakeContainerRuntimeForDispatcher{
			createFn: func(context.Context, uint, domaincontainer.CreateSpec, string) (string, error) {
				return "", errors.New("boom")
			},
		},
		fakeWorkspaceForDispatcher{},
		8,
		logrus.New(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go dispatcher.Listen(ctx)

	jobID, err := dispatcher.DispatchCreateContainer(context.Background(), domainjob.CreateContainerTask{
		UserID: 1,
		Name:   "demo",
		Image:  "busybox:latest",
	})
	if err != nil {
		t.Fatalf("DispatchCreateContainer unexpected error: %v", err)
	}

	deadline := time.After(2 * time.Second)
	for {
		job, getErr := repo.GetByJobIDAndUser(context.Background(), jobID, 1)
		if getErr != nil {
			t.Fatalf("GetByJobIDAndUser unexpected error: %v", getErr)
		}
		if job.Status == domainjob.CreateContainerJobStatusFailed {
			if job.ErrorMessage == "" {
				t.Fatalf("expected failed job has error message")
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timeout waiting job failed, current status=%s", job.Status)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestCreateContainerDispatcherChannelQueueFull(t *testing.T) {
	repo := &fakeJobRepo{jobs: map[string]domainjob.CreateContainerJob{}}
	dispatcher := NewCreateContainerDispatcherChannel(
		repo,
		fakeContainerRepoForDispatcher{
			createFn: func(_ context.Context, c *domaincontainer.Container) error {
				c.ID = 1
				return nil
			},
		},
		fakeContainerRuntimeForDispatcher{},
		fakeWorkspaceForDispatcher{},
		1,
		logrus.New(),
	)

	_, err := dispatcher.DispatchCreateContainer(context.Background(), domainjob.CreateContainerTask{UserID: 1, Image: "img"})
	if err != nil {
		t.Fatalf("first dispatch unexpected error: %v", err)
	}

	_, err = dispatcher.DispatchCreateContainer(context.Background(), domainjob.CreateContainerTask{UserID: 1, Image: "img"})
	if commonerrors.Errno(err) != consts.ErrnoContainerCreateJobQueueFull {
		t.Fatalf("expected queue full errno, got err=%v", err)
	}
}

func TestCreateContainerDispatcherChannelDrainOnShutdown(t *testing.T) {
	repo := &fakeJobRepo{jobs: map[string]domainjob.CreateContainerJob{}}
	var nextID atomic.Uint32

	dispatcher := NewCreateContainerDispatcherChannel(
		repo,
		fakeContainerRepoForDispatcher{
			createFn: func(_ context.Context, c *domaincontainer.Container) error {
				c.ID = uint(nextID.Add(1))
				return nil
			},
		},
		fakeContainerRuntimeForDispatcher{
			createFn: func(context.Context, uint, domaincontainer.CreateSpec, string) (string, error) {
				time.Sleep(80 * time.Millisecond)
				return "runtime-id", nil
			},
		},
		fakeWorkspaceForDispatcher{},
		8,
		logrus.New(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		dispatcher.Listen(ctx)
	}()

	jobIDs := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		jobID, err := dispatcher.DispatchCreateContainer(context.Background(), domainjob.CreateContainerTask{
			UserID: 1,
			Name:   "demo",
			Image:  "busybox:latest",
		})
		if err != nil {
			t.Fatalf("DispatchCreateContainer unexpected error: %v", err)
		}
		jobIDs = append(jobIDs, jobID)
	}

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("dispatcher listen did not exit after draining jobs")
	}

	for _, jobID := range jobIDs {
		job, err := repo.GetByJobIDAndUser(context.Background(), jobID, 1)
		if err != nil {
			t.Fatalf("GetByJobIDAndUser unexpected error: %v", err)
		}
		if job.Status != domainjob.CreateContainerJobStatusSucceeded {
			t.Fatalf("expected job=%s succeeded, got status=%s", jobID, job.Status)
		}
	}
}
