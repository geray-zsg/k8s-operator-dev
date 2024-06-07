package registerhandlers

import (
	"net/http"

	"github.com/golang/glog"
)

func StartHealthCheckServer(healthzPort string) {
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	healthMux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	healthServer := &http.Server{
		Addr:    healthzPort,
		Handler: healthMux,
	}

	glog.Info("Starting health check server on port ", healthzPort)
	if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		glog.Fatalf("Failed to start health check server: %v", err)
	}
}
