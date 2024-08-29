package enums

import "strings"

type WorkdayFeed int

const (
	NoneWorkDayFeed WorkdayFeed = iota
	ProcessAll
	ProcessNotSent
)

func (t WorkdayFeed) String() string {
	return [...]string{"", "processall", "processnotsent"}[t]
}

func (t WorkdayFeed) EnumIndex() int {
	return int(t)
}

func GetWorkDayFeedEnum(workdayFeed string) WorkdayFeed {
	switch strings.ToLower(workdayFeed) {
	case "processall":
		return ProcessAll
	case "processnotsent":
		return ProcessNotSent
	}
	return NoneWorkDayFeed
}
