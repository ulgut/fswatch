package main

import (
	"bufio"
	"fmt"
	"fswatch/pkg/fswatch"
	"os"
)

func main() {
	w := fswatch.New()
	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		path := reader.Text()

		err := w.Watch(path, func(event fswatch.Event) {
			fmt.Printf("path: %s, type: %d\n", event.Path, event.Type)
		})

		if err != nil {
			panic(err)
		}
	}
}
