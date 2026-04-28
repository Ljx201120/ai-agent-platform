package service

import (
	"context"
	"errors"

	"github.com/Ljx201120/ai-agent-platform/task/internal/event"
	"github.com/Ljx201120/ai-agent-platform/task/internal/model"
	"github.com/Ljx201120/ai-agent-platform/task/internal/repository"
)

type TaskService interface {
	CreateTask(ctx context.Context, userID, prompt string) (*model.Task, error)
	GetTask(ctx context.Context, id string) (*model.Task, error)
	ListTasks(ctx context.Context, userID string) ([]*model.Task, error)
}

type taskService struct {
	repo      repository.TaskRepository
	publisher *event.Publisher
}

func NewTaskService(repo repository.TaskRepository, publisher *event.Publisher) TaskService {
	return &taskService{repo: repo, publisher: publisher}
}

func (s *taskService) CreateTask(ctx context.Context, userID, prompt string) (*model.Task, error) {
	if userID == "" {
		return nil, errors.New("user_id 不能为空")
	}
	if prompt == "" {
		return nil, errors.New("prompt 不能为空")
	}

	task := &model.Task{
		UserID: userID,
		Prompt: prompt,
		Status: model.StatusPending,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	// 发布事件到 Redis Stream
	if err := s.publisher.PublishTaskCreated(ctx, task.ID, task.Prompt); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id string) (*model.Task, error) {
	if id == "" {
		return nil, errors.New("id 不能为空")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *taskService) ListTasks(ctx context.Context, userID string) ([]*model.Task, error) {
	if userID == "" {
		return nil, errors.New("user_id 不能为空")
	}
	return s.repo.ListByUserID(ctx, userID)
}
