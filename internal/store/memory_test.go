package store

import (
	"github.com/ryoshimaru/MindMap/internal/domain"
	"testing"
)

func TestProfileIsValidatedAndSaved(t *testing.T) {
	store := NewMemoryStore()
	auth, err := store.RegisterUser(domain.RegisterRequest{Name: "Иван", Email: "ivan@example.com", Password: "password"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.SaveUserProfile(auth.User.ID, domain.UpsertUserProfileRequest{Age: 25, Occupation: "Разработчик", FreeHoursPerWeek: 10, AvailableBudget: 5000})
	if err != nil {
		t.Fatal(err)
	}
	profile, err := store.GetUserProfile(auth.User.ID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Occupation != "Разработчик" {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestDependencyBlocksCompletion(t *testing.T) {
	store := NewMemoryStore()
	auth, _ := store.RegisterUser(domain.RegisterRequest{Name: "Иван", Email: "steps@example.com", Password: "password"})
	goal, err := store.CreateGoal(auth.User.ID, domain.CreateGoalRequest{Title: "Go", Description: "Изучить Go", Category: "Обучение", Priority: domain.GoalPriorityHigh})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := store.CreateTask(auth.User.ID, goal.ID, domain.CreateTaskRequest{Title: "Основы", Priority: domain.GoalPriorityMedium})
	second, _ := store.CreateTask(auth.User.ID, goal.ID, domain.CreateTaskRequest{Title: "Практика", Priority: domain.GoalPriorityMedium})
	if _, err := store.CreateDependency(auth.User.ID, second.ID, domain.CreateTaskDependencyRequest{DependsOnTaskID: first.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpdateTaskStatus(auth.User.ID, second.ID, domain.TaskStatusDone); err != ErrInvalidState {
		t.Fatalf("expected dependency error, got %v", err)
	}
	if _, err := store.UpdateTaskStatus(auth.User.ID, first.ID, domain.TaskStatusDone); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpdateTaskStatus(auth.User.ID, second.ID, domain.TaskStatusDone); err != nil {
		t.Fatal(err)
	}
}
