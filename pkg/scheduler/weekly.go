package scheduler

import (
	"time"

	"github.com/go-co-op/gocron/v2"
)

func WeeklyJob(
	weekInterval uint,
	day time.Weekday,
	hour, minute, second uint,
	taskFunc func(string),
	taskType string,
) error {
	s, _ := gocron.NewScheduler()
	_, err := s.NewJob(
		gocron.WeeklyJob(
			weekInterval, // interval in weeks
			gocron.NewWeekdays(
				day, // day of the week to run the job
			),
			gocron.NewAtTimes(
				gocron.NewAtTime(
					hour,
					minute,
					second,
				),
			),
		),
		gocron.NewTask(taskFunc, taskType),
	)

	s.Start()

	return err
}
