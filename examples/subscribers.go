package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/godbus/dbus"
	"github.com/ldx/go-geoclue2"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	gc2 := geoclue2.NewGeoClue2(conn, "")
	gc2.Start()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		go func(j int) {
			wg.Add(1)
			defer wg.Done()
			loc, err := gc2.WaitForLocation(context.TODO())
			if err != nil {
				fmt.Printf("waiting for location (client %d): %v\n", j, err)
			}
			fmt.Printf("location update (client %d): %+v\n", j, loc)
		}(i)
	}

	done := make(chan interface{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-sig:
			gc2.Stop()
			return
		case <-done:
			gc2.Stop()
			return
		}
	}
}
