package repository

import (
	"gorm.io/gorm" // Import GORM for ORM functionalities

	"github.com/yoanesber/go-batch-jobs-with-retry/internal/entity"
)

type BatchJobExecutionRepository interface {
	CreateBatchJobExecution(tx *gorm.DB, jobExecution entity.BatchJobExecution) (entity.BatchJobExecution, error)
	UpdateBatchJobExecution(tx *gorm.DB, jobExecution entity.BatchJobExecution) (entity.BatchJobExecution, error)
	GetBatchJobExecutionByID(tx *gorm.DB, jobExecutionID string) (entity.BatchJobExecution, error)
	CreateBatchJobFailureDetail(tx *gorm.DB, failureDetail entity.BatchJobFailureDetail) (entity.BatchJobFailureDetail, error)
}

type batchJobExecutionRepository struct{}

func NewBatchJobExecutionRepository() BatchJobExecutionRepository {
	return &batchJobExecutionRepository{}
}

func (r *batchJobExecutionRepository) CreateBatchJobExecution(tx *gorm.DB, jobExecution entity.BatchJobExecution) (entity.BatchJobExecution, error) {
	if err := tx.Create(&jobExecution).Error; err != nil {
		return entity.BatchJobExecution{}, err
	}
	return jobExecution, nil
}

func (r *batchJobExecutionRepository) UpdateBatchJobExecution(tx *gorm.DB, jobExecution entity.BatchJobExecution) (entity.BatchJobExecution, error) {
	if err := tx.Save(&jobExecution).Error; err != nil {
		return entity.BatchJobExecution{}, err
	}
	return jobExecution, nil
}

func (r *batchJobExecutionRepository) GetBatchJobExecutionByID(tx *gorm.DB, id string) (entity.BatchJobExecution, error) {
	var jobExecution entity.BatchJobExecution
	if err := tx.Where("id = ?", id).First(&jobExecution).Error; err != nil {
		return entity.BatchJobExecution{}, err
	}
	return jobExecution, nil
}

func (r *batchJobExecutionRepository) CreateBatchJobFailureDetail(tx *gorm.DB, failureDetail entity.BatchJobFailureDetail) (entity.BatchJobFailureDetail, error) {
	if err := tx.Create(&failureDetail).Error; err != nil {
		return entity.BatchJobFailureDetail{}, err
	}
	return failureDetail, nil
}
