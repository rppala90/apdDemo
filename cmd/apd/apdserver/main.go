package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/pborman/uuid"
	"github.com/rppala90/apdDemo/cmd/common"
	"github.com/rppala90/apdDemo/cmd/workflow"
	"go.uber.org/cadence/client"
)

type requestState string

const (
	created   requestState = "CREATED"
	approved               = "APPROVED"
	rejected               = "REJECTED"
	completed              = "COMPLETED"
)

var allRequests = make(map[string]requestState)

var tokenMap = make(map[string][]byte)

var workflowClient client.Client

func main() {
	var h common.SampleHelper
	h.SetupServiceConfig()
	var err error
	workflowClient, err = h.Builder.BuildCadenceClient()
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting apd server...")
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/action", actionHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/registerCallback", callbackHandler)
	http.ListenAndServe(":8099", nil)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>APD SYSTEM</h1>"+"<a href=\"/list\">HOME</a>"+
		"<h3>All requests:</h3><table border=1><tr><th>Request ID</th><th>Status</th><th>Action</th>")
	keys := []string{}
	for k := range allRequests {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, id := range keys {
		state := allRequests[id]
		actionLink := ""
		if state == created {
			actionLink = fmt.Sprintf("<a href=\"/action?type=approve&id=%s\">"+
				"<button style=\"background-color:#4CAF50;\">APPROVE</button></a>"+
				"&nbsp;&nbsp;<a href=\"/action?type=reject&id=%s\">"+
				"<button style=\"background-color:#f44336;\">REJECT</button></a>", id, id)
		}
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>", id, state, actionLink)
	}
	fmt.Fprint(w, "</table>")
}

func actionHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	oldState, ok := allRequests[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	if oldState == "COMPLETED" {
		fmt.Fprint(w, "Already Completed")
		return
	}
	actionType := r.URL.Query().Get("type")
	switch actionType {
	case "approve":
		allRequests[id] = approved
	case "reject":
		allRequests[id] = rejected
	case "finish":
		allRequests[id] = completed
	}

	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		listHandler(w, r)
	}

	if oldState == created && (allRequests[id] == approved || allRequests[id] == rejected) {
		// report state change
		notifyRequestStateChange(id, string(allRequests[id]))
	}

	fmt.Printf("Set state for %s from %s to %s.\n", id, oldState, allRequests[id])
	return
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "businessobject_" + uuid.New(),
		TaskList:                        workflow.ApplicationName,
		ExecutionStartToCloseTimeout:    time.Hour,
		DecisionTaskStartToCloseTimeout: time.Hour,
	}
	workflowClient.StartWorkflow(context.Background(), workflowOptions, workflow.APDWorkflow, id)
	allRequests[id] = created
	fmt.Fprintf(w, "<h1>This is like create new request in Epic # 2.<h1>"+"<h4> ID:%s<h4>", id)
	return
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	state, ok := allRequests[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}

	fmt.Fprint(w, state)
	fmt.Printf("Checking status for %s: %s\n", id, state)
	return
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	currState, ok := allRequests[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	if currState != created {
		fmt.Fprint(w, "ERROR:INVALID_STATE")
		return
	}

	err := r.ParseForm()
	if err != nil {
		// Handle error here via logging and then return
		fmt.Fprint(w, "ERROR:INVALID_FORM_DATA")
		return
	}

	taskToken := r.PostFormValue("task_token")
	fmt.Printf("Registered callback for ID=%s, token=%s\n", id, taskToken)
	tokenMap[id] = []byte(taskToken)
	fmt.Fprint(w, "SUCCEED")
}

func notifyRequestStateChange(id, state string) {
	token, ok := tokenMap[id]
	if !ok {
		fmt.Printf("Invalid id:%s\n", id)
		return
	}
	err := workflowClient.CompleteActivity(context.Background(), token, state, nil)
	if err != nil {
		fmt.Printf("Failed to complete activity with error: %+v\n", err)
	} else {
		fmt.Printf("Successfully complete activity: %s\n", token)
	}
}
