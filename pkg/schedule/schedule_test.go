package schedule

import (
	"testing"
	"time"
)

func TestContains(t *testing.T) {
	examples := []struct {
		start time.Weekday
		end   time.Weekday
		input time.Time
		res   bool
	}{
		{time.Friday, time.Monday, time.Date(2017, 12, 18, 11, 0, 0, 0, time.UTC), false},
		{time.Friday, time.Monday, time.Date(2017, 12, 19, 0, 0, 0, 0, time.UTC), false},
		{time.Friday, time.Monday, time.Date(2017, 12, 20, 0, 0, 0, 0, time.UTC), false},
		{time.Friday, time.Monday, time.Date(2017, 12, 21, 0, 0, 0, 0, time.UTC), false},
		{time.Friday, time.Monday, time.Date(2017, 12, 22, 21, 0, 0, 0, time.UTC), true}, // Friday
		{time.Friday, time.Monday, time.Date(2017, 12, 23, 0, 0, 0, 0, time.UTC), true},  // Saturday
		{time.Friday, time.Monday, time.Date(2017, 12, 24, 0, 0, 0, 0, time.UTC), true},
	}

	for _, example := range examples {
		s := New(&ScheduleSpec{
			StartTime: WeekdayTime{
				Weekday:   example.start,
				TimeOfDay: TimeOfDay{Hour: 20, Minute: 10},
			},
			EndTime: WeekdayTime{
				Weekday:   example.end,
				TimeOfDay: TimeOfDay{Hour: 10, Minute: 10},
			},
		})

		res := s.Contains(example.input)
		if res != example.res {
			t.Errorf("Contains incorrect, got: %v, expected: %v for example: %s.", res, example.res, example.input.Weekday())
		}
	}
}
