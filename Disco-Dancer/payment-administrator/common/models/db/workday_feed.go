package db

import (
	"time"
)

type WorkdayFeed struct {
	PaymentId   int64
	Amount      float32
	PaymentDate time.Time
	Metadata    string
}
