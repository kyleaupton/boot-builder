package service

import (
	"boot-builder/internal/core"
	"boot-builder/internal/eventbus"
	"boot-builder/internal/installers/linux/ubuntu_stub"
	"boot-builder/internal/jobs"
	"context"
)

type InstallerMeta struct {
	ID      string
	Name    string
	Targets []core.Target
}

type StartJobRequest struct {
	InstallerID string
	SourceLocal string
}

type JobsService struct {
	mgr        *jobs.Manager
	installers map[string]core.Installer
}

func NewJobsService() *JobsService {
	svc := &JobsService{
		installers: map[string]core.Installer{},
	}
	// register minimal installer(s)
	u := ubuntu_stub.UbuntuStub{}
	svc.installers[u.ID()] = u
	svc.mgr = jobs.NewManager(func(ev core.Event) {
		eventbus.Emit("job:event", ev)
	})
	return svc
}

func (s *JobsService) ListInstallers() []InstallerMeta {
	out := []InstallerMeta{}
	for _, inst := range s.installers {
		out = append(out, InstallerMeta{ID: inst.ID(), Name: inst.Name(), Targets: inst.Targets()})
	}
	return out
}

func (s *JobsService) StartJob(ctx context.Context, req StartJobRequest) (string, error) {
	inst, ok := s.installers[req.InstallerID]
	if !ok {
		return "", nil
	}
	plan, err := inst.Plan(ctx, core.CreateRequest{Source: core.SourceSpec{Local: req.SourceLocal}})
	if err != nil {
		return "", err
	}
	return s.mgr.Enqueue(ctx, plan)
}

func (s *JobsService) ListJobs() []jobs.Job { return s.mgr.List() }
