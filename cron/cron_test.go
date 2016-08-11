package cron

import (
	"testing"
	"time"
)

func TestCron_Next(t *testing.T) {
	c := &Cron{Src:"55 2 * * */4"}
	t.Log(c.parse())
	now := time.Date(2012, 2, 28, 3, 0, 0, 0, time.Local)
	c.GetNext(&now)
	t.Log(c.Next)
}
