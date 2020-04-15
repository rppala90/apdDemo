package main

import (
	"flag"
	"time"

	"github.com/pborman/uuid"
	"github.com/rppala90/apdDemo/cmd/common"
	"github.com/rppala90/apdDemo/cmd/pushback"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, pushback.ApplicationName, workerOptions)
}

func startWorkflow(h *common.SampleHelper, requestID string) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "businessobject_" + uuid.New(),
		TaskList:                        pushback.ApplicationName,
		ExecutionStartToCloseTimeout:    time.Hour,
		DecisionTaskStartToCloseTimeout: time.Hour,
	}
	h.StartWorkflow(workflowOptions, pushback.APDPushbackWorkflow, requestID)
}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":
		startWorkflow(&h, uuid.New())
	}
}
