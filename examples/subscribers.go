package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/godbus/dbus"
	"github.com/ldx/go-geoclue2"
	"k8s.io/klog"
)

func main() {
	workers := flag.Int("workers", 3, "Number of workers receiving updates")
	klog.InitFlags(nil)
	flag.Parse()

	conn, err := dbus.SystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	gc2 := geoclue2.NewGeoClue2(conn, "")
	gc2.Start()

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < *workers; i++ {
		go func(j int) {
			wg.Add(1)
			defer wg.Done()
			loc, err := gc2.WaitForLocation(ctx)
			if err != nil {
				klog.Infof("waiting for location (client %d): %v\n", j, err)
			}
			klog.Infof("location update (client %d): %+v\n", j, loc)
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
			cancel()
			gc2.Stop()
			return
		case <-done:
			gc2.Stop()
			return
		}
	}
}
