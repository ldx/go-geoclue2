package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	for {
		t := rand.Intn(100)
		select {
		case <-ch:
			fmt.Printf("stopping\n")
			gc2.Stop()
			fmt.Printf("stopped\n")
			return
		case <-time.After(time.Duration(t) * time.Second):
			loc := gc2.GetLatestLocation()
			fmt.Printf("latest location: %+v\n", loc)
			continue
		}
	}
}
