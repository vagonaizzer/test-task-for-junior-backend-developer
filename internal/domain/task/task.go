package task

import (
	"slices"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Status      Status              `json:"status"`
	Recurrence  *RecurrenceSettings `json:"recurrence,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// RecurrenceType — тип периодичности задачи.
type RecurrenceType string

const (
	// RecurrenceTypeDaily — задача повторяется каждые n дней, отсчёт идёт от даты создания.
	RecurrenceTypeDaily RecurrenceType = "daily"
	// RecurrenceTypeMonthly — задача выпадает на фиксированное число каждого месяца (1–30).
	RecurrenceTypeMonthly RecurrenceType = "monthly"
	// RecurrenceTypeSpecificDates — задача создаётся только в конкретные указаные даты.
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	// RecurrenceTypeEvenOdd — задача выпадает только на чётные или только на нечётные числа месяца.
	RecurrenceTypeEvenOdd RecurrenceType = "even_odd"
)

// EvenOddValue — чётность дня месяца, на которую распространяется задача.
type EvenOddValue string

const (
	EvenOddEven EvenOddValue = "even"
	EvenOddOdd  EvenOddValue = "odd"
)

// RecurrenceSettings хранит настройки периодичности задачи.
//
// Нужно заполнять только те поля, которые относятся к выбраному типу:
//   - daily          → Interval (≥1, количество дней между повторениями)
//   - monthly        → DayOfMonth (1–30, число месяца)
//   - specific_dates → Dates (список дат в формате "YYYY-MM-DD")
//   - even_odd       → EvenOdd ("even" — чётные, "odd" — нечётные)
type RecurrenceSettings struct {
	Type       RecurrenceType `json:"type"`
	Interval   *int           `json:"interval,omitempty"`
	DayOfMonth *int           `json:"day_of_month,omitempty"`
	Dates      []string       `json:"dates,omitempty"`
	EvenOdd    *EvenOddValue  `json:"even_odd,omitempty"`
}

// IsApplicableFor проверяет, должна ли задача выполнятся в указаную дату.
// startDate — дата создания задачи, используется как точка отсчёта для ежедневного типа.
func (r *RecurrenceSettings) IsApplicableFor(date time.Time, startDate time.Time) bool {
	date = date.UTC().Truncate(24 * time.Hour)

	switch r.Type {
	case RecurrenceTypeDaily:
		if r.Interval == nil || *r.Interval <= 0 {
			return false
		}
		start := startDate.UTC().Truncate(24 * time.Hour)
		if date.Before(start) {
			return false
		}
		daysDiff := int(date.Sub(start).Hours() / 24)
		return daysDiff%*r.Interval == 0

	case RecurrenceTypeMonthly:
		if r.DayOfMonth == nil {
			return false
		}
		return date.Day() == *r.DayOfMonth

	case RecurrenceTypeSpecificDates:
		return slices.Contains(r.Dates, date.Format("2006-01-02"))

	case RecurrenceTypeEvenOdd:
		if r.EvenOdd == nil {
			return false
		}
		day := date.Day()
		if *r.EvenOdd == EvenOddEven {
			return day%2 == 0
		}
		return day%2 != 0
	}

	return false
}
