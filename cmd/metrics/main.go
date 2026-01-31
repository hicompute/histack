package main

import (
	"context"
	"net/http"

	"github.com/hicompute/histack/pkg/metrics"
	"github.com/hicompute/histack/pkg/ovs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	metrics.RegisterOVSMetrics()

	agent, err := ovs.CreateOVSagent()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	agent.Start(ctx)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9476", nil)
}
