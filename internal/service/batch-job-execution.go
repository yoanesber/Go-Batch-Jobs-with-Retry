package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/yoanesber/go-batch-jobs-with-retry/config/database"
	"github.com/yoanesber/go-batch-jobs-with-retry/internal/entity"
	"github.com/yoanesber/go-batch-jobs-with-retry/internal/repository"
	"github.com/yoanesber/go-batch-jobs-with-retry/pkg/logger"
)

const (
	batchSize        = 10
	workerNamePrefix = "worker-"
	workerCount      = 5
	maxRetries       = 3
	maxElapsedTime   = 3 * time.Second // Maximum elapsed time for retries
	maxInterval      = 1 * time.Second // Maximum interval between retries
)

type BatchJobExecutionService interface {
	ProcessLargeBatch(dataType string)
}

type batchJobExecutionService struct {
	txRepo  repository.TransactionRepository
	bjeRepo repository.BatchJobExecutionRepository
}

func NewBatchJobExecutionService(txRepo repository.TransactionRepository,
	bjeRepo repository.BatchJobExecutionRepository) BatchJobExecutionService {
	return &batchJobExecutionService{txRepo: txRepo, bjeRepo: bjeRepo}
}

func (s *batchJobExecutionService) ProcessLargeBatch(jobType string) {
	db := database.GetPostgres()
	if db == nil {
		logger.Error("Database connection is nil", nil)
		return
	}

	// Create a batch job execution ID
	// This ID will be used to track the batch job execution
	batchId := fmt.Sprintf("batch-job-execution-%s-%s", jobType, time.Now().Format("20060102150405"))
	logger.Info(fmt.Sprintf("Starting batch job execution with ID: %s", batchId), nil)

	// Create a new batch job execution record
	batchJobExecution := entity.BatchJobExecution{
		ID:          batchId,
		JobType:     jobType,
		StartTime:   time.Now(),
		Status:      entity.StatusInProgress,
		LastUpdated: time.Now(),
	}
	batchJobExecution, err := s.bjeRepo.CreateBatchJobExecution(db, batchJobExecution)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create batch job execution for batch ID %s due to error: %v", batchId, err), nil)
		return
	}

	var offset int
	var numOfCompleted, numOfFailed int

	// Set default values
	status := entity.StatusCompleted
	exitCode := entity.ExitCodeSuccess
	exitMessage := "Batch job completed successfully"
	for {
		var batchTransactions []entity.Transaction
		batchTransactions, err := s.txRepo.GetAllTransactions(db, offset, batchSize)
		if err != nil {
			status = entity.StatusFailed
			exitCode = entity.ExitCodeFailure
			exitMessage = err.Error()
			break
		}

		// If no more transactions are found, exit the loop
		if len(batchTransactions) == 0 {
			logger.Info("No more transactions to process", nil)
			break
		}

		// Process each transaction in the batch
		completed, failed := processBatchConcurrently(s, batchId, batchTransactions)
		numOfCompleted += completed
		numOfFailed += failed

		offset += batchSize
	}

	// Update the batch job execution status to completed
	now := time.Now()
	batchJobExecution.EndTime = &now
	batchJobExecution.Status = status
	batchJobExecution.ExitCode = exitCode
	batchJobExecution.ExitMessage = exitMessage
	batchJobExecution.LastUpdated = now
	batchJobExecution.NumOfCompleted = numOfCompleted
	batchJobExecution.NumOfFailed = numOfFailed
	_, err = s.bjeRepo.UpdateBatchJobExecution(db, batchJobExecution)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update batch job execution for batch ID %s due to error: %v", batchId, err), nil)
		return
	}
}

func processBatchConcurrently(s *batchJobExecutionService, batchId string, batch []entity.Transaction) (int, int) {
	db := database.GetPostgres()
	if db == nil {
		logger.Error("Database connection is nil", nil)
		return 0, 0
	}

	// Create a channel to send transactions to workers
	// This will allow us to send transactions to multiple workers concurrently
	ch := make(chan entity.Transaction, len(batch))

	// Create a wait group to wait for all goroutines to finish
	// This will allow us to wait for all workers to complete their tasks
	wg := sync.WaitGroup{}

	numOfCompleted := 0
	numOfFailed := 0
	for i := 0; i < workerCount; i++ {
		wg.Add(1) // Increment the wait group counter for each worker
		go func() {
			defer wg.Done() // Ensure the goroutine signals completion
			for tx := range ch {
				// workerName := fmt.Sprintf("%s%d", workerNamePrefix, i+1)

				// Process the transaction with retry logic
				err := retryWithBackoff(func() error {
					return handleTransaction(tx)
				})
				if err != nil {
					numOfFailed++

					// Log the error and create a batch job failure detail
					batchJobFailureDetail := entity.BatchJobFailureDetail{
						BatchID:       batchId,
						DataID:        tx.ID.String(),
						ErrorMessage:  err.Error(),
						LastAttemptAt: time.Now(),
					}

					// Create a batch job failure detail record
					_, err = s.bjeRepo.CreateBatchJobFailureDetail(db, batchJobFailureDetail)
					if err != nil {
						logger.Error(fmt.Sprintf("Failed to create batch job failure detail for transaction %s due to error: %v", tx.ID.String(), err), nil)
					}

					continue
				}

				numOfCompleted++
			}
		}()
	}

	// Send transactions to the channel for processing
	// This will block until all transactions are sent
	for _, tx := range batch {
		ch <- tx
	}
	close(ch)
	wg.Wait() // Wait for all workers to finish processing

	return numOfCompleted, numOfFailed
}

// Retryable worker logic
func retryWithBackoff(operation func() error) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = maxElapsedTime // Set maximum elapsed time for retries
	expBackoff.MaxInterval = maxInterval       // Set maximum interval between retries
	retryable := backoff.WithMaxRetries(expBackoff, maxRetries)

	retryAttempts := 0
	notify := func(err error, t time.Duration) {
		retryAttempts++
		if retryAttempts >= maxRetries {
			logger.Error(fmt.Sprintf("Max retries reached (%d), giving up on operation", retryAttempts), nil)
		}
	}

	return backoff.RetryNotify(operation, retryable, notify)
}

// handleTransaction simulates processing a transaction.
func handleTransaction(tx entity.Transaction) error {
	// time.Sleep(1 * time.Second) // Simulate processing logic
	if tx.Status == "FAILED" {
		return fmt.Errorf("simulated error processing transaction %s", tx.ID.String())
	}
	return nil
}
