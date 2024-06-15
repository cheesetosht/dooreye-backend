package utility

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// TODO: implement graceful stop
func GracefulStop() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.Printf("> caught signal %v, shutting down application.", <-sigChan)

	//First become unhealthy and wait 5 seconds to drain.
	// healthcheck.IsHealthy = false
	// time.Sleep(5 * time.Second)

	//Close fiber connections
	// router.App.Shutdown()

	//Close All DB connections
	// db.CloseConn()
}
