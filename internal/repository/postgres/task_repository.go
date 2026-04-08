package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, recurrence, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, recurrence, created_at, updated_at
	`

	recurrenceJSON, err := marshalRecurrence(task.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, recurrenceJSON,
		task.CreatedAt, task.UpdatedAt,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			recurrence = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, recurrence, created_at, updated_at
	`

	recurrenceJSON, err := marshalRecurrence(task.Recurrence)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, recurrenceJSON, task.UpdatedAt, task.ID,
	)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	return r.queryTasks(ctx, query)
}

// ListRecurring достаёт из базы все задачи у которых прописана периодичность.
func (r *Repository) ListRecurring(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, recurrence, created_at, updated_at
		FROM tasks
		WHERE recurrence IS NOT NULL
		ORDER BY id DESC
	`

	return r.queryTasks(ctx, query)
}

func (r *Repository) queryTasks(ctx context.Context, query string, args ...any) ([]taskdomain.Task, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task          taskdomain.Task
		status        string
		recurrenceRaw []byte
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&recurrenceRaw,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	if recurrenceRaw != nil {
		var r taskdomain.RecurrenceSettings
		if err := json.Unmarshal(recurrenceRaw, &r); err != nil {
			return nil, fmt.Errorf("unmarshal recurrence: %w", err)
		}
		task.Recurrence = &r
	}

	return &task, nil
}

// marshalRecurrence сериализует настройки периодичности в JSON для сохранения в базу.
// Если r == nil, возвращает nil — в базе ляжет NULL.
func marshalRecurrence(r *taskdomain.RecurrenceSettings) ([]byte, error) {
	if r == nil {
		return nil, nil
	}

	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshal recurrence: %w", err)
	}

	return data, nil
}
