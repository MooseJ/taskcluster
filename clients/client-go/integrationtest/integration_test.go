package integrationtest

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/taskcluster/slugid-go/slugid"
	tcclient "github.com/taskcluster/taskcluster/v25/clients/client-go"
	"github.com/taskcluster/taskcluster/v25/clients/client-go/tcindex"
	"github.com/taskcluster/taskcluster/v25/clients/client-go/tcqueue"
	"github.com/taskcluster/taskcluster/v25/internal/jsontest"
)

// This is a silly test that looks for the garbage index namespace, to somewhat
// validate that the index client was generated correctly.
//
// Note, no credentials are needed.
func TestGarbageNamespaces(t *testing.T) {
	Index := tcindex.NewFromEnv()
	n, err := Index.ListNamespaces("garbage", "", "")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	t.Logf("Namespaces: %#v", n.Namespaces)
}

func permaCreds(t *testing.T) *tcclient.Credentials {
	permaCreds := tcclient.CredentialsFromEnvVars()
	if permaCreds.ClientID == "" || permaCreds.AccessToken == "" {
		t.Skip("Skipping test since TASKCLUSTER_CLIENT_ID and/or TASKCLUSTER_ACCESS_TOKEN env vars not set")
	}
	if permaCreds.Certificate != "" {
		t.Skip("Skipping test since temporary credentials are in use, and permanent credentials are needed")
	}
	return permaCreds
}

// Tests whether it is possible to define a task against the production Queue.
func TestDefineTask(t *testing.T) {
	rootURL := tcclient.RootURLFromEnvVars()
	if rootURL == "" {
		t.Skip("Cannot run test, neither TASKCLUSTER_PROXY_URL nor TASKCLUSTER_ROOT_URL are set to non-empty strings")
	}
	permaCreds := permaCreds(t)
	myQueue := tcqueue.New(permaCreds, rootURL)

	taskID := slugid.Nice()
	taskGroupID := slugid.Nice()
	created := time.Now()
	deadline := created.AddDate(0, 0, 1)
	expires := deadline

	td := &tcqueue.TaskDefinitionRequest{
		Created:  tcclient.Time(created),
		Deadline: tcclient.Time(deadline),
		Expires:  tcclient.Time(expires),
		Extra:    json.RawMessage(`{"index":{"rank":12345}}`),
		Metadata: tcqueue.TaskMetadata{
			Description: "Stuff",
			Name:        "[TC] Pete",
			Owner:       "pmoore@mozilla.com",
			Source:      "http://everywhere.com/",
		},
		Payload:       json.RawMessage(`{"features":{"relengApiProxy":true}}`),
		ProvisionerID: "win-provisioner",
		Retries:       5,
		Routes: []string{
			"garbage.tc-client-go.test",
		},
		SchedulerID: "go-test-test-scheduler",
		Scopes: []string{
			"queue:task-priority:high",
		},
		Tags:        map[string]string{"createdForUser": "cbook@mozilla.com"},
		Priority:    "high",
		TaskGroupID: taskGroupID,
		WorkerType:  "win2008-worker",
	}

	tsr, err := myQueue.CreateTask(taskID, td)

	//////////////////////////////////
	// And now validate results.... //
	//////////////////////////////////

	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("Task %v created successfully in %v", taskID, rootURL)

	if provisionerID := tsr.Status.ProvisionerID; provisionerID != "win-provisioner" {
		t.Errorf("provisionerId 'win-provisioner' expected but got %s", provisionerID)
	}
	if schedulerID := tsr.Status.SchedulerID; schedulerID != "go-test-test-scheduler" {
		t.Errorf("schedulerId 'go-test-test-scheduler' expected but got %s", schedulerID)
	}
	if retriesLeft := tsr.Status.RetriesLeft; retriesLeft != 5 {
		t.Errorf("Expected 'retriesLeft' to be 5, but got %v", retriesLeft)
	}
	if state := tsr.Status.State; state != "pending" {
		t.Errorf("Expected 'state' to be 'pending', but got %s", state)
	}

	taskDef, err := myQueue.Task(taskID)

	if err != nil {
		t.Fatalf("%s", err)
	}

	submittedPayload, err := json.Marshal(taskDef)

	if err != nil {
		t.Fatalf("%s", err)
	}

	// only the contents is relevant below - the formatting and order of properties does not matter
	// since a json comparison is done, not a string comparison...
	expectedJSON := []byte(`
	{
	  "created":  "` + created.UTC().Format("2006-01-02T15:04:05.000Z") + `",
	  "deadline": "` + deadline.UTC().Format("2006-01-02T15:04:05.000Z") + `",
	  "expires":  "` + expires.UTC().Format("2006-01-02T15:04:05.000Z") + `",

	  "taskGroupId": "` + taskGroupID + `",
	  "workerType":  "win2008-worker",
	  "schedulerId": "go-test-test-scheduler",

	  "payload": {
	    "features": {
	      "relengApiProxy":true
	    }
	  },

      "dependencies": [],
      "requires": "all-completed",

	  "priority":      "high",
	  "provisionerId": "win-provisioner",
	  "retries":       5,

	  "routes": [
	    "garbage.tc-client-go.test"
	  ],

	  "scopes": [
	  	"queue:task-priority:high"
	  ],

	  "tags": {
	    "createdForUser": "cbook@mozilla.com"
	  },

	  "extra": {
	    "index": {
	      "rank": 12345
	    }
	  },

	  "metadata": {
	    "description": "Stuff",
	    "name":        "[TC] Pete",
	    "owner":       "pmoore@mozilla.com",
	    "source":      "http://everywhere.com/"
	  }
	}
	`)

	jsonCorrect, formattedExpected, formattedActual, err := jsontest.JsonEqual(expectedJSON, submittedPayload)
	if err != nil {
		t.Fatalf("Exception thrown formatting json data!\n%s\n\nStruggled to format either:\n%s\n\nor:\n\n%s", err, string(expectedJSON), submittedPayload)
	}

	if !jsonCorrect {
		t.Log("Anticipated json not generated. Expected:")
		t.Logf("%s", formattedExpected)
		t.Log("Actual:")
		t.Errorf("%s", formattedActual)
	}

	// check it is possible to cancel the unscheduled task using **temporary credentials**
	tempCreds, err := permaCreds.CreateTemporaryCredentials(30*time.Second, "queue:cancel-task:"+td.SchedulerID+"/"+td.TaskGroupID+"/"+taskID)
	if err != nil {
		t.Fatalf("Exception thrown generating temporary credentials!\n\n%s\n\n", err)
	}
	myQueue = tcqueue.New(tempCreds, rootURL)
	_, err = myQueue.CancelTask(taskID)
	if err != nil {
		t.Fatalf("Exception thrown cancelling task with temporary credentials!\n\n%s\n\n", err)
	}
}
