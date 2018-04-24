package schedule

import (
	"fmt"
	"time"
)

type ScheduleSpec struct {
	StartTime WeekdayTime // Monday 8am
	EndTime   WeekdayTime // Friday 5pm
}

type WeekdayTime struct {
	Weekday   time.Weekday
	TimeOfDay TimeOfDay
}

type TimeOfDay struct {
	Hour   int
	Minute int
}

type Schedule struct {
	spec *ScheduleSpec
}

func New(schedSpec *ScheduleSpec) *Schedule {
	return &Schedule{spec: schedSpec}
}

func (s *Schedule) Contains(t time.Time) bool {
	startDay := int(s.spec.StartTime.Weekday)
	endDay := int(s.spec.EndTime.Weekday)
	inputDay := int(t.Weekday())

	if s.spec.StartTime.Weekday > s.spec.EndTime.Weekday {
		// Normalize weekdays.
		startDay = 0
		endDay = 7 + (int(s.spec.EndTime.Weekday) - int(s.spec.StartTime.Weekday))
		inputDay = (7 + (int(t.Weekday()) - int(s.spec.StartTime.Weekday))) % 7
	}

	nowHour := t.Hour()
	startHour := s.spec.StartTime.TimeOfDay.Hour
	endHour := s.spec.EndTime.TimeOfDay.Hour

	nowMinute := t.Minute()
	startMinute := s.spec.StartTime.TimeOfDay.Minute
	endMinute := s.spec.EndTime.TimeOfDay.Minute

	// If our day is in between the start and end day.
	if inputDay >= startDay && inputDay <= endDay {

		// Return early on start and end day when its outside of hours.
		if t.Weekday() == s.spec.StartTime.Weekday {
			if startHour > nowHour && startMinute < nowMinute {
				return false
			}
		}
		if t.Weekday() == s.spec.EndTime.Weekday {
			if endHour < nowHour && endMinute > nowMinute {
				return false
			}
		}

		return true
	}

	return false
}

func ConvertWeekday(day string) (time.Weekday, error) {
	weekdays := map[string]time.Weekday{
		"Monday":    time.Weekday(1),
		"monday":    time.Weekday(1),
		"Tuesday":   time.Weekday(2),
		"tuesday":   time.Weekday(2),
		"Wednesday": time.Weekday(3),
		"wednesday": time.Weekday(3),
		"Thursday":  time.Weekday(4),
		"thursday":  time.Weekday(4),
		"Friday":    time.Weekday(5),
		"friday":    time.Weekday(5),
		"Saturday":  time.Weekday(6),
		"saturday":  time.Weekday(6),
		"Sunday":    time.Weekday(0),
		"sunday":    time.Weekday(0),
	}
	value, ok := weekdays[day]
	if !ok {
		return 0, fmt.Errorf("Wrong weekday.")
	}
	return value, nil
}
