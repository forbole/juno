package telemetry

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RunAdditionalOperations runs the module additional operations
func RunAdditionalOperations(cfg *Config) error {
	err := checkConfig(cfg)
	if err != nil {
		return err
	}

	go startPrometheus(cfg)

	return nil
}

// checkConfig checks if the given config is valid
func checkConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("no telemetry config found")
	}

	return nil
}

// startPrometheus starts a Prometheus server using the given configuration
func startPrometheus(cfg *Config) {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	// Create a new server
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
