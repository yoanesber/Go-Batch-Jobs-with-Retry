package scheduler

import (
	"github.com/go-co-op/gocron/v2"
)

func MonthlyJob(
	monthInterval uint,
	dayOfMonth int,
	hour, minute, second uint,
	taskFunc func(string),
	taskType string,
) error {
	s, _ := gocron.NewScheduler()
	_, err := s.NewJob(
		gocron.MonthlyJob(
			monthInterval, // interval in weeks
			gocron.NewDaysOfTheMonth(
				dayOfMonth, // day of the month to run the job
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
