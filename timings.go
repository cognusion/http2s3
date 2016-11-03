package main

import (
	"log"
	"time"
)

// This doesn't belong here, but it needed a subpackage so I could grab it
// other subpackages
type Change struct {
	Time    time.Time
	Changed string
}

// If you want to track how long a function takes, make the first line
// defer Track("func_name", time.Now(), logger).
func Track(name string, start time.Time, l *log.Logger) {
	elapsed := time.Since(start)
	l.Printf("%s took %s", name, elapsed)
}

// A helper function in case "time" isn't loaded elsewhere
func Now() time.Time {
	return time.Now()
}

// A helper function in case "time" isn't loaded elsewhere
func Sleep(dur string) {
	d, err := time.ParseDuration(dur)
	if err == nil {
		time.Sleep(d)
	}
}
