package runner

import (
	"errors"
	"fmt"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

// Sleeper and action injection keeps ordering/failure semantics testable without
// wall-clock delays. A cancelled wait always aborts, even on a continue step.
func executeScheduleTasks(tasks []database.ScheduleTask, sleep func(time.Duration) error, execute func(database.ScheduleTask) error) error {
	if err := database.ValidateScheduleTasks(tasks); err != nil {
		return err
	}
	var failures []error
	for i, task := range tasks {
		if err := sleep(time.Duration(task.TimeOffset) * time.Second); err != nil {
			return errors.Join(append(failures, fmt.Errorf("task %d delay: %w", i+1, err))...)
		}
		if err := execute(task); err != nil {
			failures = append(failures, fmt.Errorf("task %d (%s): %w", i+1, task.Action, err))
			if !task.ContinueOnFailure {
				break
			}
		}
	}
	return errors.Join(failures...)
}

func (m *ServerManager) waitScheduleDelay(delay time.Duration) error {
	if m.isShuttingDown() {
		return fmt.Errorf("daemon shutting down")
	}
	if delay == 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-m.shutdownCh:
		return fmt.Errorf("daemon shutting down")
	case <-timer.C:
		return nil
	}
}
