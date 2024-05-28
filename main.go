package main

import (
	"bufio"
	"fmt"
	"fswatch/pkg/fswatch"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	w := fswatch.New()
	reader := bufio.NewScanner(os.Stdin)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("received signal: %v\n", sig)
		w.Shutdown()
	}()

	reader.Scan()
	path := reader.Text()

	wg := sync.WaitGroup{}

	wg.Add(1)
	err := w.Watch(path, func(event fswatch.Event) {
		fmt.Printf("path: %s, type: %d\n", event.Path, event.Type)
		wg.Done()
	})

	if err != nil {
		panic(err)
	}

	wg.Wait()

	if err := w.Remove(path); err != nil {
		panic(err)
	}
}
