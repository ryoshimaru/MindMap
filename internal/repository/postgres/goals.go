package postgres

import (
	"context"

	"github.com/ryoshimaru/MindMap/internal/domain"
)

func (s *Store) ListGoals(ctx context.Context, userID string) ([]domain.Goal, error) {
	rows, err := s.db.QueryContext(ctx, `
		select id, title, description, category, priority, status, deadline,
		       available_hours_per_week, created_at, updated_at
		from goals
		where user_id = $1
		order by updated_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goals := make([]domain.Goal, 0)
	for rows.Next() {
		var goal domain.Goal
		if err := rows.Scan(
			&goal.ID,
			&goal.Title,
			&goal.Description,
			&goal.Category,
			&goal.Priority,
			&goal.Status,
			&goal.Deadline,
			&goal.AvailableHoursPerWeek,
			&goal.CreatedAt,
			&goal.UpdatedAt,
		); err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	return goals, rows.Err()
}
