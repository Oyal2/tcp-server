package executor

import (
	"context"

	"github.com/Oyal2/tcp-server/internal/model"
)

type TaskExecutor interface {
	ExecuteTask(ctx context.Context, taskRequest *model.TaskRequest) *model.TaskResult
}
