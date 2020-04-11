package workflow

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	ApplicationName = "apdGroup"
)

var ServerHostPort = "http://localhost:8099"

// This is registration process where you register all your workflow handlers.
func init() {
	fmt.Println("Workflow init()")
	workflow.Register(APDWorkflow)
}

// APDWorkflow execution
func APDWorkflow(ctx workflow.Context, requestID string) (result string, err error) {
	/*// step 1, create new Billing Object
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Hour,
		StartToCloseTimeout:    time.Hour,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	err = workflow.ExecuteActivity(ctx1, createBillingObjectActivity, requestID).Get(ctx1, nil)
	if err != nil {
		logger.Error("Failed to create report", zap.Error(err))
		return "", err
	}*/

	logger := workflow.GetLogger(ctx)

	// step 2, wait for the Billing Object to be approved (or rejected)
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Hour,
		StartToCloseTimeout:    10 * time.Hour,
	}
	ctx2 := workflow.WithActivityOptions(ctx, ao)
	// Notice that we set the timeout to be 10 minutes for this sample demo. If the expected time for the activity to
	// complete (waiting for human to approve the request) is longer, you should set the timeout accordingly so the
	// cadence system will wait accordingly. Otherwise, cadence system could mark the activity as failure by timeout.
	var status string
	err = workflow.ExecuteActivity(ctx2, waitForDecisionActivity, requestID).Get(ctx2, &status)
	if err != nil {
		return "", err
	}

	if status != "APPROVED" {
		logger.Info("Workflow completed.", zap.String("RequestStatus", status))
		return "", nil
	}

	// step 3, request
	err = workflow.ExecuteActivity(ctx2, approveBillingActivity, requestID).Get(ctx2, nil)
	if err != nil {
		logger.Info("Workflow completed with failed.", zap.Error(err))
		return "", err
	}

	logger.Info("Workflow completed with business object completed.")
	return "COMPLETED", nil
}
