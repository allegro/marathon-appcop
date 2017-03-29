package mgc

import (
	"time"
)

// MarathonDate parses and hold marathon date in marathon specific format
type MarathonDate struct {
	timeFormat string
	inTime     time.Time
}

func toMarathonDate(dateStr string) (*MarathonDate, error) {
	timeFormat := "2006-01-02T15:04:05.000Z"
	inTime, err := time.Parse(timeFormat, dateStr)
	if err != nil {
		return nil, err
	}
	return &MarathonDate{timeFormat: timeFormat, inTime: inTime}, nil
}

func (d *MarathonDate) elapsed() time.Duration {
	return time.Since(d.inTime)
}
