// Package timer handles running functions for specified amount of time.
package timer

import (
	"fmt"
	"log"
	"time"
)

// RunFor runs a function for a specified amount of time.
// Then it notifies the function that it has to stop.
func RunFor(f func(<-chan struct{}), durationStr string) {
	stopCh := make(chan struct{})
	defer close(stopCh)

	if durationStr != "" {
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			log.Fatal(fmt.Sprintf("couldn't parse duration '%s'", duration))
		}

		log.Println(fmt.Sprintf("running for %s", duration.String()))
		go f(stopCh)

		time.Sleep(duration)
		stopCh <- struct{}{}
	} else {
		log.Println("running indefinitely")
		f(stopCh)
	}
}
