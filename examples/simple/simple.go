package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/godbus/dbus"
	"github.com/ldx/go-geoclue2"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

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
		t := rand.Intn(10)
		select {
		case <-ch:
			klog.Infof("stopping")
			gc2.Stop()
			klog.Infof("stopped")
			return
		case <-time.After(time.Duration(t) * time.Second):
			loc := gc2.GetLatestLocation()
			klog.Infof("latest location: %+v\n", loc)
			continue
		}
	}
}
