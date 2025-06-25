package scheduler

import (
	"time"

	"github.com/go-co-op/gocron/v2"
)

func DurationJob(hour, minute, second uint, taskFunc func(string), taskType string) error {
	s, _ := gocron.NewScheduler()
	_, err := s.NewJob(
		gocron.DurationJob(
			time.Duration(hour)*time.Hour+
				time.Duration(minute)*time.Minute+
				time.Duration(second)*time.Second,
		),
		gocron.NewTask(taskFunc, taskType),
	)

	s.Start()

	return err
}
