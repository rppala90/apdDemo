package workflow

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

// This is registration process where you register all your activity handlers.
func init() {
	fmt.Println("activity init()")
	//activity.Register(createBillingObjectActivity)
	activity.Register(waitForDecisionActivity)
	activity.Register(approveBillingActivity)
}

/*func createBillingObjectActivity(ctx context.Context, requestID string) error {
	if len(requestID) == 0 {
		return errors.New("request id is empty")
	}

	resp, err := http.Get(ServerHostPort + "/create?is_api_call=true&id=" + requestID)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("Billing Object created.", zap.String("requestID", requestID))
		return nil
	}

	return errors.New(string(body))
}
*/

// This activity will complete asynchronously. When this method returns error activity.ErrResultPending,
// the cadence client recognize this error, and won't mark this activity as failed or completed.
// The cadence server will wait until Client.CompleteActivity() is called or timeout happened
// whichever happen first.
// In this case, the CompleteActivity() method is called by apd server when the business object is approved.
func waitForDecisionActivity(ctx context.Context, requestID string) (string, error) {
	if len(requestID) == 0 {
		return "", errors.New("requestID is empty")
	}

	logger := activity.GetLogger(ctx)

	activityInfo := activity.GetInfo(ctx)
	formData := url.Values{}
	formData.Add("task_token", string(activityInfo.TaskToken))

	registerCallbackURL := ServerHostPort + "/registerCallback?id=" + requestID
	resp, err := http.PostForm(registerCallbackURL, formData)
	if err != nil {
		logger.Info("waitForDecisionActivity failed to register callback.", zap.Error(err))
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	status := string(body)
	if status == "SUCCEED" {
		// register callback succeed
		logger.Info("Successfully registered callback.", zap.String("RequestID", requestID))

		// ErrActivityResultPending is returned from activity's execution to indicate the activity is not completed when it returns.
		// activity will be completed asynchronously when Client.CompleteActivity() is called.
		return "", activity.ErrResultPending
	}

	logger.Warn("Register callback failed.", zap.String("RequestStatus", status))
	return "", fmt.Errorf("register callback failed status:%s", status)
}

func approveBillingActivity(ctx context.Context, requestID string) error {
	if len(requestID) == 0 {
		return errors.New("request id is empty")
	}

	resp, err := http.Get(ServerHostPort + "/action?is_api_call=true&type=finish&id=" + requestID)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("billing item creation succeed", zap.String("RequestID", requestID))
		return nil
	}

	return errors.New(string(body))
}
