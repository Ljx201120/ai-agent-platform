package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Ljx201120/ai-agent-platform/task/internal/model"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id string) (*model.Task, error)
	UpdateStatus(ctx context.Context, id string, status model.TaskStatus, result string) error
	ListByUserID(ctx context.Context, userID string) ([]*model.Task, error)
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *model.Task) error {
	task.ID = uuid.New().String()
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).First(&task, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) UpdateStatus(ctx context.Context, id string, status model.TaskStatus, result string) error {
	return r.db.WithContext(ctx).Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status": status,
			"result": result,
		}).Error
}

func (r *taskRepository) ListByUserID(ctx context.Context, userID string) ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}
