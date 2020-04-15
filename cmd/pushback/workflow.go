package pushback

import (
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	ApplicationName = "apdGroup"
)

// This is registration process where you register all your workflow handlers.
func init() {
	fmt.Println("Workflow init()")
	workflow.Register(APDPushbackWorkflow)
}

// APDPushbackWorkflow execution
func APDPushbackWorkflow(ctx workflow.Context, requestID string) (result string, err error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * 24 * time.Hour,
		StartToCloseTimeout:    5 * 24 * time.Hour,
	}
	workflowCTX := workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	var num int
	err = workflow.ExecuteActivity(workflowCTX, startWorkflowActivity, requestID).Get(workflowCTX, &num)
	if err != nil {
		logger.Error("Failed to start workflow", zap.Error(err))
		return "ERROR", err
	}
	fmt.Println("Completed startWorkflowActivity activity")
	for i := 1; i <= num; i++ {
		retryPolicy := &cadence.RetryPolicy{
			InitialInterval: time.Second,
			MaximumAttempts: 2,
		}
		aoi := workflow.ActivityOptions{
			ScheduleToStartTimeout: 5 * 24 * time.Hour,
			StartToCloseTimeout:    5 * 24 * time.Hour,
			RetryPolicy:            retryPolicy, // Enable retry.
		}
		workflowCTX = workflow.WithActivityOptions(ctx, aoi)
		workflow.ExecuteActivity(workflowCTX, approvalActivity)
	}
	fmt.Println("Running finishWorkflowActivity")
	var status string
	err = workflow.ExecuteActivity(workflowCTX, finishWorkflowActivity).Get(workflowCTX, &status)
	fmt.Println(status)
	if err != nil {
		logger.Error("Workflow completed with error")
		return "ERROR", err
	}
	logger.Info("Workflow completed with business object completed.")
	return "COMPLETED", nil
}
