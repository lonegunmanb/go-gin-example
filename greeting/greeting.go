package greeting

import (
	"fmt"
	"time"
)

func Greeting() string {
	hour := time.Now().Hour()
	var greeting string
	if hour >= 6 && hour <= 12 {
		greeting = "Morning"
	} else if hour > 12 && hour <= 17 {
		greeting = "Afternoon"
	} else {
		greeting = "Evening"
	}
	return fmt.Sprintf("Good %s", greeting)
}
