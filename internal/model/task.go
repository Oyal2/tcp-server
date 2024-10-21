package model

type TaskRequest struct {
	Command []string `json:"command"`
	Timeout int      `json:"timeout,omitempty"`
}

type TaskResult struct {
	Command    []string `json:"command"`
	ExecutedAt int64    `json:"executed_at"`
	DurationMs float64  `json:"duration_ms"`
	ExitCode   int      `json:"exit_code"`
	Output     string   `json:"output,omitempty"`
	Error      string   `json:"error,omitempty"`
}
