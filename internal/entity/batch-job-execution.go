package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	// StatusInProgress represents the in-progress status of a batch job execution.
	StatusInProgress = "IN_PROGRESS"
	// StatusCompleted represents the completed status of a batch job execution.
	StatusCompleted = "COMPLETED"
	// StatusFailed represents the failed status of a batch job execution.
	StatusFailed = "FAILED"

	// ExitCodeSuccess represents a successful exit code for batch job execution.
	ExitCodeSuccess = "SUCCESS"
	// ExitCodeFailure represents a failure exit code for batch job execution.
	ExitCodeFailure = "FAILURE"
	// ExitCodeTimeout represents a timeout exit code for batch job execution.
	ExitCodeTimeout = "TIMEOUT"

	// JobTypeTransaction represents the job type for transaction processing.
	JobTypeTransaction = "TRANSACTION"
)

// BatchJobExecution represents the batch job execution entity in the database.
type BatchJobExecution struct {
	ID             string     `gorm:"type:varchar(100);primaryKey"`
	JobType        string     `gorm:"type:varchar(20);not null"`
	StartTime      time.Time  `gorm:"type:timestamptz"`
	EndTime        *time.Time `gorm:"type:timestamptz"`
	Status         string     `gorm:"type:varchar(20);not null"`
	ExitCode       string     `gorm:"type:varchar(20)"`
	ExitMessage    string     `gorm:"type:text"`
	LastUpdated    time.Time  `gorm:"type:timestamptz"`
	NumOfCompleted int        `gorm:"type:int;default:0"`
	NumOfFailed    int        `gorm:"type:int;default:0"`
}

type BatchJobFailureDetail struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BatchID       string    `gorm:"type:varchar(50);not null"`
	DataID        string    `gorm:"type:varchar(50);not null"`
	ErrorMessage  string    `gorm:"type:text"`
	LastAttemptAt time.Time `gorm:"autoUpdateTime"`

	// Optional: Relation
	BatchJobExecution BatchJobExecution `gorm:"foreignKey:BatchID;references:ID;constraint:OnDelete:CASCADE"`
}
