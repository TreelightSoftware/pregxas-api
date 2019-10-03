package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateParsing(t *testing.T) {
	desiredOutput := "2017-11-28T10:02:03Z"
	desiredSQLOutput := "2017-11-28 10:02:03"
	desiredJustDate := "2017-11-28T00:00:00Z"
	input1 := "2017-11-28 10:02:03"
	input2 := "2017-11-28T10:02:03"
	input3 := "2017-11-28T10:02:03Z"
	input4 := "2017-11-28T11:02:03+01:00"
	input5 := "2017-11-28T05:02:03-05:00"
	input6 := "2017-11-28"
	input7 := "2017-11-28T10:02:03Z"

	_, err := ParseTimeToISO("badtime")
	assert.NotNil(t, err)
	output1, err := ParseTimeToISO(input1)
	assert.Nil(t, err)
	assert.Equal(t, desiredOutput, output1)

	output2, err := ParseTimeToISO(input2)
	assert.Nil(t, err)
	assert.Equal(t, desiredOutput, output2)

	output3, err := ParseTimeToISO(input3)
	assert.Nil(t, err)
	assert.Equal(t, desiredOutput, output3)

	output4, err := ParseTimeToISO(input4)
	assert.Nil(t, err)
	assert.Equal(t, desiredOutput, output4)

	output5, err := ParseTimeToISO(input5)
	assert.Nil(t, err)
	assert.Equal(t, desiredOutput, output5)

	sql1, err := ParseISOTimeToDBTime(output1)
	assert.Nil(t, err)
	assert.Equal(t, desiredSQLOutput, sql1)

	output6, err := ParseTimeToISO(input6)
	assert.Nil(t, err)
	assert.Equal(t, desiredJustDate, output6)

	sql2, err := ParseISOTimeToDBTime(input7)
	assert.Nil(t, err)
	assert.Equal(t, desiredSQLOutput, sql2)
}

func Test24HourValid(t *testing.T) {
	assert.True(t, IsValid24HourTime("15:04", "00:00"))
	assert.True(t, IsValid24HourTime("15:04", "03:23"))
	assert.True(t, IsValid24HourTime("15:04", "10:20"))
	assert.True(t, IsValid24HourTime("15:04", "23:59"))
	assert.False(t, IsValid24HourTime("15:04", "25:00"))
	assert.False(t, IsValid24HourTime("15:04", "16:6"))
}

func TestDataRanges(t *testing.T) {
	// note that this is constructed poorly on purpose; the method should sort
	// and combine
	start := []DatePoint{
		DatePoint{
			Day:   "2018-06-30",
			Count: 8,
		},
		DatePoint{
			Day:   "2018-07-05",
			Count: 6,
		},
		DatePoint{
			Day:   "2018-07-01 15:32:04",
			Count: 9,
		},
		DatePoint{
			Day:   "2018-07-01",
			Count: 3,
		},
		DatePoint{
			Day:   "2018-07-02",
			Count: 3,
		},
		DatePoint{
			Day:   "2018-06-28",
			Count: 1,
		},
	}

	expected := []DatePoint{
		DatePoint{
			Day:          "2018-06-27",
			Count:        0,
			RunningTotal: 0,
		},
		DatePoint{
			Day:          "2018-06-28",
			Count:        1,
			RunningTotal: 1,
		},
		DatePoint{
			Day:          "2018-06-29",
			Count:        0,
			RunningTotal: 1,
		},
		DatePoint{
			Day:          "2018-06-30",
			Count:        8,
			RunningTotal: 9,
		},
		DatePoint{
			Day:          "2018-07-01",
			Count:        12,
			RunningTotal: 21,
		},
		DatePoint{
			Day:          "2018-07-02",
			Count:        3,
			RunningTotal: 24,
		},
		DatePoint{
			Day:          "2018-07-03",
			Count:        0,
			RunningTotal: 24,
		},
		DatePoint{
			Day:          "2018-07-04",
			Count:        0,
			RunningTotal: 24,
		},
		DatePoint{
			Day:          "2018-07-05",
			Count:        6,
			RunningTotal: 30,
		},
	}

	simple, err := CreateDateRangesForReports([]DatePoint{}, "2018-06-27", "2018-06-27")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(simple.Data))

	processed, err := CreateDateRangesForReports(start, "2018-06-27 00:00:00", "2018-07-05")
	assert.Nil(t, err)

	// check for expected results
	for index := range processed.Data {
		assert.Equal(t, expected[index].Day, processed.Data[index].Day)
		assert.Equal(t, expected[index].Count, processed.Data[index].Count)
		assert.Equal(t, expected[index].RunningTotal, processed.Data[index].RunningTotal)
	}
	assert.Equal(t, int64(30), processed.Total)
}

func TestDurations(t *testing.T) {
	ConfigSetup()

	awhileAgo := time.Now().AddDate(-1, -1, -1)
	year, month, day, hour, min, sec := CalculateDuration(awhileAgo)
	assert.Equal(t, 1, year)
	assert.Equal(t, 1, month)
	assert.Equal(t, 1, day)
	assert.Equal(t, 0, hour)
	assert.Equal(t, 0, min)
	assert.Equal(t, 0, sec)

	sixThirty10Ago := time.Now().Add(time.Hour * -6).Add(time.Minute * -30).Add(time.Second * -10)
	year, month, day, hour, min, sec = CalculateDuration(sixThirty10Ago)
	assert.Equal(t, 0, year)
	assert.Equal(t, 0, month)
	assert.Equal(t, 0, day)
	assert.Equal(t, 6, hour)
	assert.Equal(t, 30, min)
	assert.Equal(t, 10, sec)
}
