# A Go library for Geoclue2

[Geoclue2](https://developer.gnome.org/platform-overview/unstable/tech-geoclue2.html.en) is a DBus service for providing geolocation services. This is a Go library for making it easy to get location information via Geoclue2.

## Install

Use `go get github.com/ldx/go-geoclue2`.

## Usage

A simple workflow for starting the service, waiting for one location update, and then stopping the service:

	// Connect to DBus.
	conn, err := dbus.SystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
    // Start service and wait for location.
	gc2 := geoclue2.NewGeoClue2(conn, "")
	gc2.Start()
	loc, err := gc2.WaitForLocation(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("location update: %+v\n", j, loc)
    // Stop service.
	gc2.Stop()

There are more examples in `examples/`.
