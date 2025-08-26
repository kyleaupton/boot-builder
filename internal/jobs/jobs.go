package jobs

import (
	"boot-builder/internal/core"
	"context"
	"sync"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

type Job struct {
	ID        string
	Plan      *core.Plan
	Status    Status
	Progress  float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Manager struct {
	mu   sync.Mutex
	jobs map[string]*Job
	emit func(ev core.Event)
}

func NewManager(emit func(ev core.Event)) *Manager {
	return &Manager{jobs: make(map[string]*Job), emit: emit}
}

func (m *Manager) List() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, *j)
	}
	return out
}

func (m *Manager) Enqueue(ctx context.Context, plan *core.Plan) (string, error) {
	m.mu.Lock()
	id := time.Now().Format("20060102150405.000")
	job := &Job{ID: id, Plan: plan, Status: StatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.jobs[id] = job
	m.mu.Unlock()

	go m.run(ctx, job)
	return id, nil
}

type execAdapter struct {
	emit  func(ev core.Event)
	jobID string
}

func (e execAdapter) Emit(ev core.Event) { ev.JobID = e.jobID; e.emit(ev) }

func (m *Manager) run(ctx context.Context, job *Job) {
	m.mu.Lock()
	job.Status = StatusRunning
	job.UpdatedAt = time.Now()
	m.mu.Unlock()
	m.emit(core.Event{JobID: job.ID, Type: "state", Message: string(job.Status)})

	e := execAdapter{emit: m.emit, jobID: job.ID}
	total := float64(len(job.Plan.Steps))
	for i, s := range job.Plan.Steps {
		e.Emit(core.Event{Type: "step-start", Message: s.Name()})
		err := s.Run(ctx, e)
		if err != nil {
			m.mu.Lock()
			job.Status = StatusFailed
			job.UpdatedAt = time.Now()
			m.mu.Unlock()
			m.emit(core.Event{JobID: job.ID, Type: "state", Message: string(job.Status), Error: err.Error()})
			return
		}
		m.mu.Lock()
		job.Progress = float64(i+1) / total
		job.UpdatedAt = time.Now()
		m.mu.Unlock()
		e.Emit(core.Event{Type: "progress", Percent: job.Progress * 100})
		e.Emit(core.Event{Type: "step-end", Message: s.Name()})
	}

	m.mu.Lock()
	job.Status = StatusSucceeded
	job.UpdatedAt = time.Now()
	m.mu.Unlock()
	m.emit(core.Event{JobID: job.ID, Type: "state", Message: string(job.Status)})
}
