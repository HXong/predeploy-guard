package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	RunTaskQueued    = "queued"
	RunTaskRunning   = "running"
	RunTaskCompleted = "completed"
	RunTaskFailed    = "failed"
)

type RunManager struct {
	service *RunService
	mu      sync.RWMutex
	tasks   map[string]RunTask
}

type RunTask struct {
	TaskID     string     `json:"taskId"`
	Profile    string     `json:"profile"`
	Status     string     `json:"status"`
	Passed     bool       `json:"passed"`
	Error      string     `json:"error,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Message    string     `json:"message,omitempty"`
	Logs       []string   `json:"logs"`
}

func NewRunManager(service *RunService) *RunManager {
	return &RunManager{
		service: service,
		tasks:   make(map[string]RunTask),
	}
}

func (m *RunManager) Start(profile string) RunTask {
	now := time.Now()

	task := RunTask{
		TaskID:    newRunTaskID(),
		Profile:   runTaskProfile(profile),
		Status:    RunTaskQueued,
		CreatedAt: now,
		Message:   "validation run queued",
		Logs:      []string{"validation run queued"},
	}

	m.mu.Lock()
	m.tasks[task.TaskID] = task
	m.mu.Unlock()

	go m.run(task.TaskID, profile)

	return copyRunTask(task)
}

func (m *RunManager) Get(taskID string) (RunTask, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return RunTask{}, false
	}

	return copyRunTask(task), true
}

func (m *RunManager) GetLogs(taskID string) ([]string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, false
	}

	return copyLogs(task.Logs), true
}

func (m *RunManager) List() []RunTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]RunTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, copyRunTask(task))
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	return tasks
}

func (m *RunManager) run(taskID string, profile string) {
	startedAt := time.Now()
	m.update(taskID, func(task RunTask) RunTask {
		task.Status = RunTaskRunning
		task.StartedAt = &startedAt
		task.Message = "validation run started"
		task.Logs = appendRunTaskLog(task, "validation run started")
		return task
	})

	result, err := m.service.Run(profile)
	finishedAt := time.Now()

	m.update(taskID, func(task RunTask) RunTask {
		task.FinishedAt = &finishedAt

		if err != nil {
			task.Status = RunTaskFailed
			task.Passed = false
			task.Error = err.Error()
			task.Message = ""
			task.Logs = appendRunTaskLog(task, fmt.Sprintf("validation run failed: %s", err.Error()))
			return task
		}

		task.Profile = result.Profile
		task.Passed = result.Passed
		task.Error = result.Error
		task.Message = result.Message

		if result.Passed {
			task.Status = RunTaskCompleted
			task.Logs = appendRunTaskLog(task, "validation run completed")
		} else {
			failureMessage := result.Error
			if failureMessage == "" {
				failureMessage = "validation checks failed"
			}

			task.Status = RunTaskFailed
			task.Logs = appendRunTaskLog(task, fmt.Sprintf("validation run failed: %s", failureMessage))
		}

		return task
	})
}

func (m *RunManager) update(taskID string, update func(RunTask) RunTask) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return
	}

	m.tasks[taskID] = update(task)
}

func newRunTaskID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return "task-" + hex.EncodeToString(bytes[:])
	}

	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

func runTaskProfile(profile string) string {
	if profile == "" {
		return "base"
	}

	return profile
}

func appendRunTaskLog(task RunTask, message string) []string {
	logs := copyLogs(task.Logs)
	return append(logs, message)
}

func copyRunTask(task RunTask) RunTask {
	task.Logs = copyLogs(task.Logs)
	return task
}

func copyLogs(logs []string) []string {
	copied := make([]string, len(logs))
	copy(copied, logs)
	return copied
}
