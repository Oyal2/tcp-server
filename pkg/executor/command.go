package executor

import (
	"context"
	"os/exec"
	"time"

	"github.com/Oyal2/tcp-server/internal/constant"
	"github.com/Oyal2/tcp-server/internal/model"
)

type CommandExecutor struct{}

func NewCommandExecutor() *CommandExecutor {
	ce := CommandExecutor{}
	return &ce
}

func (ce *CommandExecutor) ExecuteTask(ctx context.Context, request *model.TaskRequest) *model.TaskResult {
	// Build the Task Result
	result := &model.TaskResult{
		Command:    request.Command,
		ExecutedAt: time.Now().Unix(),
	}

	// Safe check if the command coming in is set
	if request.Command == nil {
		result.ExitCode = -1
		result.Error = constant.TaskResultCommandNilError
		return result
	}

	// Setup the command that we will be running with the context.
	cmd := exec.CommandContext(ctx, request.Command[0], request.Command[1:]...)
	// Run it and collect the outputs
	output, err := cmd.CombinedOutput()
	// Populate the ouput to our result
	result.Output = string(output)

	// If there was a context deadline then we set the result to timeout exceeded
	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		result.Error = constant.TaskResultTimeoutError
	} else if err != nil {
		// If there was an error during the execution, then output that error
		result.ExitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = err.Error()
	}

	// Get the duration
	result.DurationMs = calculateDuration(result.ExecutedAt)

	return result
}

func calculateDuration(executedAt int64) float64 {
	return float64(time.Since(time.Unix(executedAt, 0))) / float64(time.Millisecond)
}
