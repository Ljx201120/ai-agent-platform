package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ljx201120/ai-agent-platform/task/internal/service"
)

type TaskHandler struct {
	svc service.TaskService
}

func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/api/tasks")
	g.POST("", h.CreateTask)
	g.GET("/:id", h.GetTask)
	g.GET("", h.ListTasks)
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Prompt string `json:"prompt"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.svc.CreateTask(c.Request.Context(), req.UserID, req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	task, err := h.svc.GetTask(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id 不能为空"})
		return
	}

	tasks, err := h.svc.ListTasks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}
