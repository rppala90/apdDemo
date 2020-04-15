package pushback

import (
	"context"
	"errors"
	"fmt"

	"github.com/rppala90/apdDemo/cmd/common"
	"go.uber.org/cadence/activity"
)

// This is registration process where you register all your activity handlers.
func init() {
	fmt.Println("activity init()")
	activity.Register(startWorkflowActivity)
	activity.Register(approvalActivity)
	activity.Register(finishWorkflowActivity)
}

func startWorkflowActivity(ctx context.Context, requestID string) (int, error) {
	fmt.Println("Enter Number of approvals in workflow")
	var input int
	_, err := fmt.Scanln(&input)
	return input, err
}

func approvalActivity(ctx context.Context) error {
	fmt.Println("TaskToken is :" + string(activity.GetInfo(ctx).TaskToken))
	return activity.ErrResultPending
}

func finishWorkflowActivity(ctx context.Context) (string, error) {
	var h common.SampleHelper
	h.SetupServiceConfig()
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}
	var taskToken string
	for taskToken != "DONE" {
		fmt.Println("Enter taskToken for completeActivity or RETRY for RetryActivity or DONE to exit")
		_, err = fmt.Scanln(&taskToken)
		if taskToken == "RETRY" {
			fmt.Println("Enter taskToken to retry")
			_, err = fmt.Scanln(&taskToken)
			workflowClient.CompleteActivity(context.Background(), []byte(taskToken), "error", errors.New("Pushback"))
		} else if taskToken != "DONE" {
			err = workflowClient.CompleteActivity(context.Background(), []byte(taskToken), "completed", nil)
			if err != nil {
				return "error", err
			}
		}

	}
	fmt.Println("Return from Final Activity")
	return "completed", nil

}
