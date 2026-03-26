package utils

import (
	"time"
	_ "time/tzdata" // Ensure timezone data is available even if not provided by OS
)

const JakartaTimezone = "Asia/Jakarta"

// GetJakartaTime returns the current time in Asia/Jakarta timezone
func GetJakartaTime() time.Time {
	loc, err := time.LoadLocation(JakartaTimezone)
	if err != nil {
		// Fallback to UTC+7 offset if location loading fails
		return time.Now().UTC().Add(7 * time.Hour)
	}
	return time.Now().In(loc)
}

// GetJakartaDateString returns the current date in YYYY-MM-DD format (Jakarta time)
func GetJakartaDateString() string {
	return GetJakartaTime().Format("2006-01-02")
}

// GetIndonesianDayName returns the day name in Indonesian for a given time
func GetIndonesianDayName(t time.Time) string {
	days := map[time.Weekday]string{
		time.Sunday:    "Minggu",
		time.Monday:    "Senin",
		time.Tuesday:   "Selasa",
		time.Wednesday: "Rabu",
		time.Thursday:  "Kamis",
		time.Friday:    "Jumat",
		time.Saturday:  "Sabtu",
	}
	return days[t.Weekday()]
}
