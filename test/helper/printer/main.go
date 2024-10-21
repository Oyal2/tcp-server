package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	message := flag.String("message", "test", "The message to print")
	repeat := flag.Int("repeat", 1, "Number of times to repeat the message")
	sleep := flag.Int("sleep", 0, "Time to sleep for in milliseconds")
	flag.Parse()

	time.Sleep(time.Duration(*sleep) * time.Millisecond)

	for i := 0; i < *repeat; i++ {
		_, err := fmt.Println(*message)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error printing to STDOUT: %v\n", err)
			os.Exit(1)
		}
	}
}
