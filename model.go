package proxy

// TaskID specifies task identifier.
type TaskID string

// TaskMode specifies task execution mode.
type TaskMode string

// TaskMode values.
const (
	Sequential TaskMode = "sequential"
	Parallel            = "parallel"
)

// TaskConfig specifies task parameters when creating new task.
type TaskConfig struct {
	ClientID    string   `json:"client_id"`
	Info        string   `json:"info"`
	Mode        TaskMode `json:"mode"`
	FailOnError bool     `json:"failonerror"`
}

// Status specifies remote command execution status.
type Status string

// Status values.
const (
	Pending Status = "pending"
	Running        = "running"
	Success        = "success"
	Failure        = "failure"
	Killed         = "killed"
)

// Result represents remote command execution result.
type Result struct {
	Addr   string `json:"addr"`
	Status Status `json:"status"`
	Msg    string `json:"message,omitempty"`
}

// TaskStatus represents overall task status.
type TaskStatus struct {
	Results []Result `json:"results"` // enforce copy when returning status
}
