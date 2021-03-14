package api

import (
	"fmt"
	"sort"
	"time"
)

func ParseTime(input string) (time.Time, error) {
	//run through all of the supported formats and try to convert
	t, err := time.Parse("2006-01-02 15:04:05", input)
	if err != nil {
		t, err = time.Parse("2006-01-02", input)
	}
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", input)
	}
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", input)
	}
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z07:00", input)
	}
	return t, err
}

// ParseTimeToISO takes in a time string, massages it, and attempts to convert it to RFC3399/ISO8601 format. We expect the input to be in UTC already
func ParseTimeToISO(input string) (string, error) {
	t, err := ParseTime(input)

	if err != nil {
		return "", err
	}
	t = t.UTC()
	output := t.Format(time.RFC3339)
	return output, nil
}

// ParseTimeToDate parses an input to just its date component
func ParseTimeToDate(input string) (string, error) {
	t, err := ParseTime(input)

	if err != nil {
		return "", err
	}
	t = t.UTC()
	output := t.Format("2006-01-02")
	return output, nil
}

// ParseISOTimeToDBTime takes a ISO formatted string (such as from ParseTimeToISO) and converts it to MySQL's native DB format YYYY-MM-DD HH:mm:SS
func ParseISOTimeToDBTime(input string) (string, error) {
	t, err := ParseTime(input)
	if err != nil {
		return "", err
	}
	t = t.UTC()
	output := t.Format("2006-01-02 15:04:05")
	return output, nil
}

//CalculateDuration gets the duration from a certain date time to now
//timezone is not important for this one and we always calculate from now
//shamelessly adapted from icza at https://stackoverflow.com/questions/36530251/golang-time-since-with-months-and-years
func CalculateDuration(input time.Time) (year, month, day, hour, min, sec int) {
	now := time.Now()
	y1, M1, d1 := input.Date()
	y2, M2, d2 := now.Date()

	h1, m1, s1 := input.Clock()
	h2, m2, s2 := now.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

// IsValid24HourTime checks to see if a string is a valid 24h formatted hour
func IsValid24HourTime(inputFormat, input string) bool {
	_, err := time.Parse(inputFormat, input)
	if err != nil {
		return false
	}
	return true
}

// DatePoint is a day/count mapping for use in the reporting methods
type DatePoint struct {
	Count        int64  `json:"count" db:"count"`
	RunningTotal int64  `json:"runningTotal"`
	Day          string `json:"day" db:"day"`
}

// DatePointsProcessed is the total amount of data processed
type DatePointsProcessed struct {
	Data  []DatePoint `json:"data"`
	Total int64       `json:"total"`
}

// CreateDateRangesForReports takes in an array of DatePoints, fills in gaps for missing dates, and returns
// a new slice with a total
func CreateDateRangesForReports(input []DatePoint, startDate string, endDate string) (ret DatePointsProcessed, err error) {
	data := []DatePoint{}

	// add the start and end dates in the input with 0 values just to set the ret
	input = append(input, DatePoint{
		Day:          startDate[0:10],
		Count:        0,
		RunningTotal: 0,
	})
	if endDate != startDate {
		input = append(input, DatePoint{
			Day:          endDate[:10],
			Count:        0,
			RunningTotal: 0,
		})
	}

	// now we sort
	sort.Slice(input, func(i, j int) bool {
		return input[i].Day < input[j].Day
	})

	firstDay, err := time.Parse("2006-01-02", input[0].Day[0:10])
	if err != nil {
		return
	}
	lastDay, err := time.Parse("2006-01-02", input[len(input)-1].Day[0:10])
	if err != nil {
		return
	}
	if firstDay == lastDay {
		input[0].RunningTotal = input[0].Count
		return DatePointsProcessed{
			Data:  input,
			Total: input[0].Count,
		}, nil
	}

	// now we start the loop. here's the algorithm
	// each iteration, the current date is parsed to make sure it is a valid date (YYYY-MM-DD)
	// if it's the last index, we just add it
	// otherwise, we add it and get the next day
	// for each 24 hour difference (-1 since we don't want to add the blank for a date already there), we add a hole

	lastIndex := len(input) - 1
	total := int64(0)
	for index := range input {
		// sanity check on the date
		currentDate, err := time.Parse("2006-01-02", input[index].Day[0:10])
		if err != nil {
			return DatePointsProcessed{}, fmt.Errorf("unexpected date found: %s, error was %v", input[index].Day[0:10], err)
		}

		if index == lastIndex {
			// we are on the last one; make sure it wasn't duped
			if index != 0 && input[index].Day[0:10] == input[index-1].Day[0:10] {
				// should have been the last one in there
				total += input[index].Count
				data[len(data)-1].Count += input[index].Count
				data[len(data)-1].RunningTotal += input[index].Count
				break
			}
			total += input[index].Count
			input[index].RunningTotal = total
			data = append(data, input[index])
			break
		}

		// check to see if this was already added in
		if index != 0 && input[index].Day[0:10] == input[index-1].Day[0:10] {
			// should have been the last one in there
			total += input[index].Count
			data[len(data)-1].Count += input[index].Count
			data[len(data)-1].RunningTotal += input[index].Count

		} else {
			// now add it in
			total += input[index].Count
			input[index].RunningTotal = total
			data = append(data, input[index])
		}

		// see what the next date is
		nextDate, err := time.Parse("2006-01-02", input[index+1].Day[0:10])
		if err != nil {
			return DatePointsProcessed{}, fmt.Errorf("unexpected date found: %s, error was %v", input[index].Day[0:10], err)
		}

		// get the diff
		diff := nextDate.Sub(currentDate)
		if diff.Hours() > 24 {
			// how many days are missing?
			iterations := int(diff.Hours() / 24)
			for i := 1; i < iterations; i++ {
				// put in a hole
				data = append(data, DatePoint{
					Count:        0,
					Day:          currentDate.AddDate(0, 0, i).Format("2006-01-02"),
					RunningTotal: total,
				})
			}
		}

	}

	// now build the ret
	ret = DatePointsProcessed{
		Data:  data,
		Total: total,
	}

	return
}
