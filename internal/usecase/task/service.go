package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Recurrence:  normalized.Recurrence,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

// ListByDate возвращает все периодические задачи, которые подходят под указаную дату.
func (s *Service) ListByDate(ctx context.Context, date time.Time) ([]taskdomain.Task, error) {
	recurring, err := s.repo.ListRecurring(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]taskdomain.Task, 0)
	for _, task := range recurring {
		if task.Recurrence.IsApplicableFor(date, task.CreatedAt) {
			result = append(result, task)
		}
	}

	return result, nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateRecurrence(input.Recurrence); err != nil {
		return CreateInput{}, err
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateRecurrence(input.Recurrence); err != nil {
		return UpdateInput{}, err
	}

	return input, nil
}

// validateRecurrence проверяет что настройки периодичности соответствуют выбраному типу.
// nil — это нормально, просто значит что задача без периодичности.
func validateRecurrence(r *taskdomain.RecurrenceSettings) error {
	if r == nil {
		return nil
	}

	switch r.Type {
	case taskdomain.RecurrenceTypeDaily:
		if r.Interval == nil {
			return fmt.Errorf("%w: recurrence.interval is required for type 'daily'", ErrInvalidInput)
		}
		if *r.Interval < 1 {
			return fmt.Errorf("%w: recurrence.interval must be at least 1", ErrInvalidInput)
		}

	case taskdomain.RecurrenceTypeMonthly:
		if r.DayOfMonth == nil {
			return fmt.Errorf("%w: recurrence.day_of_month is required for type 'monthly'", ErrInvalidInput)
		}
		if *r.DayOfMonth < 1 || *r.DayOfMonth > 30 {
			return fmt.Errorf("%w: recurrence.day_of_month must be between 1 and 30", ErrInvalidInput)
		}

	case taskdomain.RecurrenceTypeSpecificDates:
		if len(r.Dates) == 0 {
			return fmt.Errorf("%w: recurrence.dates must not be empty for type 'specific_dates'", ErrInvalidInput)
		}
		for _, d := range r.Dates {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return fmt.Errorf("%w: recurrence.dates contains invalid date %q (expected YYYY-MM-DD)", ErrInvalidInput, d)
			}
		}

	case taskdomain.RecurrenceTypeEvenOdd:
		if r.EvenOdd == nil {
			return fmt.Errorf("%w: recurrence.even_odd is required for type 'even_odd'", ErrInvalidInput)
		}
		if *r.EvenOdd != taskdomain.EvenOddEven && *r.EvenOdd != taskdomain.EvenOddOdd {
			return fmt.Errorf("%w: recurrence.even_odd must be 'even' or 'odd'", ErrInvalidInput)
		}

	default:
		return fmt.Errorf("%w: unknown recurrence type %q", ErrInvalidInput, r.Type)
	}

	return nil
}
