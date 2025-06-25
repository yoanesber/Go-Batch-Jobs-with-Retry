package scheduler

import (
	"github.com/go-co-op/gocron/v2"
)

func DailyJob(hour, minute, second uint, taskFunc func(string), taskType string) error {
	s, _ := gocron.NewScheduler()
	_, err := s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(hour, minute, second),
			),
		),
		gocron.NewTask(taskFunc, taskType),
	)

	s.Start()

	return err
}
