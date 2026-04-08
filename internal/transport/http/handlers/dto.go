package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// recurrenceDTO mirrors taskdomain.RecurrenceSettings for JSON transport.
type recurrenceDTO struct {
	Type       taskdomain.RecurrenceType  `json:"type"`
	Interval   *int                       `json:"interval,omitempty"`
	DayOfMonth *int                       `json:"day_of_month,omitempty"`
	Dates      []string                   `json:"dates,omitempty"`
	EvenOdd    *taskdomain.EvenOddValue   `json:"even_odd,omitempty"`
}

type taskMutationDTO struct {
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Status      taskdomain.Status         `json:"status"`
	Recurrence  *recurrenceDTO            `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	if task.Recurrence != nil {
		dto.Recurrence = &recurrenceDTO{
			Type:       task.Recurrence.Type,
			Interval:   task.Recurrence.Interval,
			DayOfMonth: task.Recurrence.DayOfMonth,
			Dates:      task.Recurrence.Dates,
			EvenOdd:    task.Recurrence.EvenOdd,
		}
	}

	return dto
}

func recurrenceToDomain(dto *recurrenceDTO) *taskdomain.RecurrenceSettings {
	if dto == nil {
		return nil
	}

	return &taskdomain.RecurrenceSettings{
		Type:       dto.Type,
		Interval:   dto.Interval,
		DayOfMonth: dto.DayOfMonth,
		Dates:      dto.Dates,
		EvenOdd:    dto.EvenOdd,
	}
}
