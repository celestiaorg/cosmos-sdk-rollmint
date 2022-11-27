package types

const (
	WriteOp  Operation = "write"
	ReadOp   Operation = "read"
	DeleteOp Operation = "delete"
)

type (
	// operation represents an IO operation
	Operation string

	// traceOperation implements a traced KVStore operation
	TraceOperation struct {
		Operation Operation `json:"operation"`
		Key       string    `json:"key"`
		Value     string    `json:"value"`
	}
)
