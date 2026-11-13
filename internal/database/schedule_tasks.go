package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const MaxScheduleTasks = 20
const MaxCommandPayload = 4096

// Order is a stable, one-based position within a schedule. IDs survive reads and
// unrelated updates; replacing the task array assigns new IDs.
type ScheduleTask struct {
	ID                int64  `json:"id"`
	Order             int    `json:"order"`
	Action            string `json:"action"`
	Payload           string `json:"payload"`
	TimeOffset        int    `json:"time_offset"`
	ContinueOnFailure bool   `json:"continue_on_failure"`
}

func (db *DB) MigrateScheduleTasks() error {
	_, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS schedule_tasks (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 schedule_id TEXT NOT NULL REFERENCES server_schedules(id) ON DELETE CASCADE,
 task_order INTEGER NOT NULL,
 action TEXT NOT NULL,
 payload TEXT NOT NULL DEFAULT '',
 time_offset INTEGER NOT NULL DEFAULT 0 CHECK(time_offset BETWEEN 0 AND 3600),
 continue_on_failure INTEGER NOT NULL DEFAULT 0,
 UNIQUE(schedule_id, task_order)
 );
 CREATE TRIGGER IF NOT EXISTS delete_schedule_tasks AFTER DELETE ON server_schedules
 BEGIN DELETE FROM schedule_tasks WHERE schedule_id = OLD.id; END;`)
	return err
}

func ValidateScheduleTasks(tasks []ScheduleTask) error {
	if len(tasks) == 0 || len(tasks) > MaxScheduleTasks {
		return fmt.Errorf("tasks must contain 1 to %d steps", MaxScheduleTasks)
	}
	for i, t := range tasks {
		if t.TimeOffset < 0 || t.TimeOffset > 3600 {
			return fmt.Errorf("task %d: time_offset must be an integer from 0 to 3600 seconds", i+1)
		}
		if len(t.Payload) > MaxCommandPayload {
			return fmt.Errorf("task %d: payload exceeds %d bytes", i+1, MaxCommandPayload)
		}
		switch t.Action {
		case "command":
			if strings.TrimSpace(t.Payload) == "" {
				return fmt.Errorf("task %d: command payload is required", i+1)
			}
		case "backup", "start", "stop", "restart":
		default:
			return fmt.Errorf("task %d: supported actions: command, backup, start, stop, restart", i+1)
		}
	}
	return nil
}

type taskQuerier interface {
	Query(string, ...any) (*sql.Rows, error)
}

func readScheduleTasks(q taskQuerier, s *ServerSchedule) ([]ScheduleTask, error) {
	rows, err := q.Query(`SELECT id, task_order, action, payload, time_offset, continue_on_failure FROM schedule_tasks WHERE schedule_id = ? ORDER BY task_order`, s.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []ScheduleTask{}
	for rows.Next() {
		var t ScheduleTask
		if err := rows.Scan(&t.ID, &t.Order, &t.Action, &t.Payload, &t.TimeOffset, &t.ContinueOnFailure); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		tasks = append(tasks, ScheduleTask{Order: 1, Action: s.Action, Payload: s.Payload})
	}
	return tasks, nil
}

func (db *DB) ListScheduleTasks(s *ServerSchedule) ([]ScheduleTask, error) {
	return readScheduleTasks(db.conn, s)
}

// SaveScheduleWithTasks commits the schedule and its task array as one unit.
// nil tasks on update preserves all existing steps; legacy create needs no rows.
func (db *DB) SaveScheduleWithTasks(s *ServerSchedule, tasks []ScheduleTask, create bool) error {
	if _, err := ParseCron(s.CronExpr); err != nil {
		return err
	}
	if tasks != nil {
		if err := ValidateScheduleTasks(tasks); err != nil {
			return err
		}
		s.Action, s.Payload = tasks[0].Action, tasks[0].Payload
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	s.UpdatedAt = time.Now().UTC()
	if create {
		s.CreatedAt = s.UpdatedAt
		_, err = tx.Exec(`INSERT INTO server_schedules (id, server_id, name, cron_expr, action, payload, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, s.ID, s.ServerID, s.Name, s.CronExpr, s.Action, s.Payload, s.IsActive, s.CreatedAt, s.UpdatedAt)
	} else {
		res, e := tx.Exec(`UPDATE server_schedules SET name = ?, cron_expr = ?, action = ?, payload = ?, is_active = ?, updated_at = ? WHERE server_id = ? AND id = ?`, s.Name, s.CronExpr, s.Action, s.Payload, s.IsActive, s.UpdatedAt, s.ServerID, s.ID)
		err = affectedRow(res, e)
	}
	if err != nil {
		return err
	}
	if tasks != nil {
		if _, err := tx.Exec(`DELETE FROM schedule_tasks WHERE schedule_id = ?`, s.ID); err != nil {
			return err
		}
		for i, t := range tasks {
			if _, err := tx.Exec(`INSERT INTO schedule_tasks (schedule_id, task_order, action, payload, time_offset, continue_on_failure) VALUES (?, ?, ?, ?, ?, ?)`, s.ID, i+1, t.Action, t.Payload, t.TimeOffset, t.ContinueOnFailure); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// ClaimScheduleWithTasks snapshots both schedule and steps inside the claim
// transaction. No DB connection is held during waits or server operations.
func (db *DB) ClaimScheduleWithTasks(id string, now time.Time, scheduled bool) (*ServerSchedule, []ScheduleTask, bool, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return nil, nil, false, err
	}
	defer tx.Rollback()
	query := `UPDATE server_schedules SET last_status = 'started', last_error = '', last_run_at = ?, last_finished_at = NULL WHERE id = ? AND last_status NOT IN ('started', 'running')`
	args := []any{now.UTC(), id}
	if scheduled {
		query += ` AND is_active = 1 AND (last_run_at IS NULL OR last_run_at < ?)`
		args = append(args, now.UTC().Truncate(time.Minute))
	}
	res, err := tx.Exec(query, args...)
	if err != nil {
		return nil, nil, false, err
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return nil, nil, false, err
	}
	s, err := scanSchedule(tx.QueryRow(`SELECT `+scheduleColumns+` FROM server_schedules WHERE id = ?`, id))
	if err != nil {
		return nil, nil, false, err
	}
	tasks, err := readScheduleTasks(tx, s)
	if err != nil {
		return nil, nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, false, err
	}
	return s, tasks, true, nil
}
