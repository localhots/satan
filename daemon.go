package uberdaemon

// Daemon is the interface that contains a set of methods required to be
// implemented in order to be treated as a daemon.
type Daemon interface {
	// Startup implementation should...
	//
	// func (d *DaemonName) Startup() {
	//     // Set up a panic handler:
	//     b.HandlePanics(func() {
	//         log.Error("Oh, crap!")
	//     })
	//
	//     // If the daemon is also a consumer we need to subscribe for topics
	//     // that would be consumed by the daemon.
	//     b.Subscribe("ProductPriceUpdates", func(p PriceUpdate) {
	// 	       log.Printf("Price for %q is now $%.2f", p.Product, p.Amount)
	//     })
	//
	//     // If the daemon is doing some IO it is a good idea to limit the rate
	//     // of its execution.
	//     b.SetRateLimit(10, 1 * time.Second)
	// }
	Startup()

	// Shutdown implementation should clean up all daemon related stuff:
	// close channels, process the last batch of items, etc.
	Shutdown()

	// base is a (hack) function that allows the Daemon interface to reference
	// underlying BaseDaemon structure.
	base() *BaseDaemon
}

