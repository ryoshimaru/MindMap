package service

import (
	"github.com/ryoshimaru/MindMap/internal/domain"
	"github.com/ryoshimaru/MindMap/internal/store"
)

type TaskService struct {
	store store.Store
}

func NewTaskService(store store.Store) *TaskService {
	return &TaskService{store: store}
}

func (s *TaskService) ListByGoal(userID, goalID string) ([]domain.TaskTreeNode, error) {
	return s.store.ListTasksByGoal(userID, goalID)
}

func (s *TaskService) Create(userID, goalID string, req domain.CreateTaskRequest) (domain.Task, error) {
	return s.store.CreateTask(userID, goalID, req)
}

func (s *TaskService) Get(userID, taskID string) (domain.Task, error) {
	return s.store.GetTask(userID, taskID)
}

func (s *TaskService) Update(userID, taskID string, req domain.UpdateTaskRequest) (domain.Task, error) {
	return s.store.UpdateTask(userID, taskID, req)
}

func (s *TaskService) Delete(userID, taskID string) error {
	return s.store.DeleteTask(userID, taskID)
}

func (s *TaskService) UpdateStatus(userID, taskID string, status domain.TaskStatus) (domain.Task, error) {
	return s.store.UpdateTaskStatus(userID, taskID, status)
}

func (s *TaskService) ListDependencies(userID, taskID string) ([]domain.TaskDependency, error) {
	return s.store.ListDependencies(userID, taskID)
}

func (s *TaskService) CreateDependency(userID, taskID string, req domain.CreateTaskDependencyRequest) (domain.TaskDependency, error) {
	return s.store.CreateDependency(userID, taskID, req)
}

func (s *TaskService) DeleteDependency(userID, taskID, dependencyID string) error {
	return s.store.DeleteDependency(userID, taskID, dependencyID)
}
