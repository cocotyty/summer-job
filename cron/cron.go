package cron

import (
	"strings"
	"strconv"
	"errors"
	"time"
	"github.com/emirpasic/gods/maps/treemap"
)

type empty struct {

}

var emp = empty{}

type Cron struct {
	Src          string
	MonthSet     *treemap.Map
	DaySet       *treemap.Map
	HourSet      *treemap.Map
	MinuteSet    *treemap.Map
	SecondSet    *treemap.Map
	DayOfWeekSet *treemap.Map
	Next         *time.Time
}

func (c *Cron)parse() (err error) {

	parts := strings.Split(c.Src, " ")
	if len(parts) != 5 {
		return errors.New("wrong cron format")
	}
	minute, hour, day, month, dayOfWeek := parts[0], parts[1], parts[2], parts[3], parts[4]
	// , - * /
	if c.MinuteSet, err = parse(minute, 0, 59); err != nil {
		return
	}
	if c.HourSet, err = parse(hour, 0, 23); err != nil {
		return
	}
	if c.DaySet, err = parse(day, 1, 30); err != nil {
		return
	}
	if c.MonthSet, err = parse(month, 1, 12); err != nil {
		return
	}
	if c.DayOfWeekSet, err = parse(dayOfWeek, 0, 6); err != nil {
		return
	}
	if c.DayOfWeekSet != nil && c.DaySet != nil {
		return errors.New("you can only define day of week or day of month")
	}
	return nil
}
func parse(str string, start, end int) (*treemap.Map, error) {
	set := treemap.NewWithIntComparator()
	if str == "*" {
		return nil, nil
	} else if strings.Contains(str, ",") {
		minutes := strings.Split(str, ",")
		for _, m := range minutes {
			m, err := strconv.Atoi(m)
			if err != nil {
				return nil, errors.New(str + " is not right")
			}
			if m < 0 {
				return nil, errors.New(str + " is not right, number must big than 0")
			}
			set.Put(m, emp)
		}
		return set, nil
	} else if strings.Contains(str, "-") {
		fromTo := strings.Split(str, "-")
		if len(fromTo) != 2 {
			return nil, errors.New(str + " is not right")
		}
		from, err := strconv.Atoi(fromTo[0])
		if err != nil || from < 0 {
			return nil, errors.New(str + " is not right,number not right")
		}
		to, err := strconv.Atoi(fromTo[1])
		if err != nil || from < 0 {
			return nil, errors.New(str + " is not right,number not right")
		}
		if to < from {
			index := from
			for ; index <= end; index++ {
				set.Put(index, emp)
			}
			index = start
			for ; index <= to; index++ {
				set.Put(index, emp)
			}
		} else {
			index := from
			for ; index <= to; index++ {
				set.Put(index, emp)
			}
		}
		return set, nil
	} else if strings.Contains(str, "/") {
		begin := 0
		if strings.HasPrefix(str, "*/") {
			begin = 2
		} else if strings.HasPrefix(str, "/") {
			begin = 1
		} else {
			return nil, errors.New(str + " is not right , '*/x' or '/x' is right")
		}
		every, err := strconv.Atoi(str[begin:])
		if err != nil || every < 0 {
			return nil, errors.New(str + " is not right")
		}
		for i := 0; i < end; i = i + every {
			set.Put(i, emp)
		}
		return set, nil
	} else {
		m, err := strconv.Atoi(str)
		if err != nil || m < 0 {
			return nil, errors.New("wrong number:" + str)
		}
		set.Put(m, emp)
		return set, nil
	}
}

func (c *Cron)Before(c2 *Cron, t *time.Time) {

}
func lastDay(now *time.Time) int {
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	return lastOfMonth.Day()
}
func (c *Cron)GetNext(t *time.Time) {
	weekday := int(t.Weekday())
	year, month, day := t.Date()
	hour := t.Hour()
	minute := t.Minute()
	lastDayOfMonth := lastDay(t)
	nt := &nextTime{lastDayOfMonth, year, int(month), day, hour, minute, t.Location()}
	if c.MinuteSet != nil {
		k, _ := c.MinuteSet.Find(func(key interface{}, value interface{}) bool {
			return key.(int) > nt.minute
		})
		if k == nil {
			k, _ = c.MinuteSet.Min()
			nt.HourNext()
			hour = nt.hour
		}
		nt.minute = k.(int)
	} else {
		nt.MinuteNext()
	}

	if c.HourSet != nil {
		k, _ := c.HourSet.Find(func(key interface{}, value interface{}) bool {
			return key.(int) >= nt.hour
		})
		if k == nil {
			k, _ = c.HourSet.Min()
			nt.DayNext()
			day = nt.day
		}
		nt.hour = k.(int)
		if nt.hour != hour {
			nt.ResetMinute(c.MinuteSet)
		}
	}

	if c.DaySet != nil {
		k, _ := c.DaySet.Find(func(key interface{}, value interface{}) bool {
			return key.(int) >= nt.day
		})
		if k == nil {
			k, _ = c.DaySet.Min()
			nt.MonthNext()
			month = time.Month(nt.month)
		}
		nt.day = k.(int)
		if nt.day != day {
			if c.HourSet == nil {
				nt.ResetHour(c.HourSet)
			}
			if c.MinuteSet == nil {
				nt.ResetMinute(c.MinuteSet)
			}
		}
	} else if c.DayOfWeekSet != nil {
		k, _ := c.DayOfWeekSet.Find(func(key interface{}, value interface{}) bool {
			return key.(int) >= weekday
		})
		var blank int
		if k == nil {
			k, _ = c.DayOfWeekSet.Min()
			blank = k.(int) + (6 - weekday)
		} else {
			blank = k.(int) - weekday
		}
		nt.AddDay(blank)
		month = time.Month(nt.month)
		if nt.day != day {
			nt.ResetHour(c.HourSet)
			nt.ResetMinute(c.MinuteSet)
		}
	}
	if c.MonthSet != nil {
		k, _ := c.MonthSet.Find(func(key interface{}, value interface{}) bool {
			return key.(int) >= nt.month
		})
		if k == nil {
			k, _ = c.MonthSet.Min()
			nt.YearNext()
		}
		nt.month = k.(int)
		if nt.month != int(month) {
			nt.ResetHour(c.HourSet)
			nt.ResetMinute(c.MinuteSet)
			if c.DaySet != nil {
				nt.ResetDay(c.DaySet)
			} else if c.DayOfWeekSet != nil {
				nt.ResetWeekDay(c.DayOfWeekSet)
			} else {
				nt.day = 1
			}
		}
	}
	c.Next = nt.GetTime()
}

type nextTime struct {
	lastDayOfMonth int
	year           int
	month          int
	day            int
	hour           int
	minute         int
	location       *time.Location
}

func (nt *nextTime) YearNext() {
	nt.year++
}
func (nt *nextTime) MonthNext() {
	if nt.month == 12 {
		nt.month = 1
		nt.YearNext()
	} else {
		nt.month++
	}
}
func (nt *nextTime) DayNext() {
	if nt.day == nt.lastDayOfMonth {
		nt.day = 1
		nt.MonthNext()
	} else {
		nt.day++
	}
}
func (nt *nextTime) AddDay(n int) {
	if nt.day <= nt.lastDayOfMonth - n {
		nt.day += n
	} else {
		nt.day = n - (nt.lastDayOfMonth - nt.day)
		nt.MonthNext()
	}
}
func (nt *nextTime) HourNext() {
	if nt.hour == 23 {
		nt.hour = 1
		nt.DayNext()
	} else {
		nt.hour++
	}
}
func (nt *nextTime) MinuteNext() {
	if nt.minute == 59 {
		nt.minute = 0
		nt.HourNext()
	} else {
		nt.minute++
	}
}
func (nt *nextTime) ResetMinute(minuteSet *treemap.Map) {
	if minuteSet != nil {
		k, _ := minuteSet.Min()
		nt.minute = k.(int)
	} else {
		nt.minute = 0
	}
}
func (nt *nextTime) ResetHour(hourSet *treemap.Map) {
	if hourSet != nil {
		k, _ := hourSet.Min()
		nt.hour = k.(int)
	} else {
		nt.hour = 0
	}
}
func (nt *nextTime) ResetDay(daySet *treemap.Map) {
	if daySet != nil {
		k, _ := daySet.Min()
		nt.day = k.(int)
	} else {
		nt.day = 1
	}
}
func (nt *nextTime) ResetWeekDay(weekdaySet *treemap.Map) {
	k, _ := weekdaySet.Min()
	weekday := k.(int)
	firstWeekDay := int(time.Date(nt.year, time.Month(nt.month), 1, 0, 0, 0, 0, nt.location).Weekday())
	if weekday == firstWeekDay {
		nt.day = 1
	} else if weekday > firstWeekDay {
		nt.day = weekday - firstWeekDay
	} else {
		nt.day = (6 - firstWeekDay + nt.day)
	}
}
func (nt *nextTime)GetTime() *time.Time {
	t := time.Date(nt.year, time.Month(nt.month), nt.day, nt.hour, nt.minute, 0, 0, nt.location)
	return &t
}

type Action interface {
	Call()
	Cron() Cron
}