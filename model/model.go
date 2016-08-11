package model

import (
	"github.com/cocotyty/gorm"
	"time"
)

type JobType int

const (
	Idempotence JobType = iota
	SampleJob
)

type Job struct {
	gorm.Model
	Cron          string
	Name          string
	Type          JobType
	ConfirmStatus int
	RetryTimes    int
	RetryInterval int
}

type Task struct {
	gorm.Model
	Job     *Job
	JobID   uint
	JobTime time.Time
}
