# Polling-Worker Model
Go module for setting up a generic's based job queue with a polling worker. Utilizes https://github.com/alphadose/haxmap as its base for sync.Map functionality.

# Concepts
- (Server) Generics based job queue
- (Client) Generics based worker which polls server for new jobs, does whatever it needs to do, then posts the results back to the server
- Client and Server both initialized with the same generic parameters [J, R]
	- where J is the user defined type of the job and R is the user defined type of the result

# Use Case
I use this to dynamically access internal resources from externally hosted web applications which do not have direct access to the internal network. Maybe you'll find other uses.

# Example Usage

## Server
```golang
package main

import (
	"encoding/json"
	"net/http"
	
	"github.com/fillingc/pollingworkermodel/common"
	"github.com/fillingc/pollingworkermodel/server"
	"github.com/go-chi/chi"
)

// Custom job information needed by the client to do whatever it needs to do,
// this needs to be the same for the server and client, easily done with a shared
// module or just copying the struct to both places.
type CustomJob struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func main() {
	// initialize the QueueServer with the generic job type, generic result type, 
	// and the worker authorization key.
	queue := server.NewQueueServer[*CustomJob, string]("<worker-authorization-key>")

	// QueueServer[J, R any] implements ServeHTTP, I'm sure most http servers would work 
	// but I'm using go-chi here.
	router := chi.NewRouter()

	// mount handler which processes:
	// 
	// GET /jobs
	// - responds with json serialized []*common.Job[J, R any]
	//
	// POST /results
	// - expects json serialized *common.Job[J, R any]
	//
	// both requests check for the validity of the worker authorization key
	router.Mount("/queue/", queue)

	// example request handler which creates a job and waits for the result before returning it as json
	router.Get("/do-something", func(w http.ResponseWriter, r *http.Request) {
		req := &CustomJob{
			Field1: "whatever",
			Field2: "other stuff",
		}

		job := queue.AddWaitJob(req)

		if job.Status == common.Error {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(job); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe("localhost:8888", router)
}
```

## Client
```golang
package main

import (
	"fmt"
	"github.com/fillingc/pollingworkermodel/client"
	"github.com/fillingc/pollingworkermodel/common"
)

// Custom job information needed by the client to do whatever it needs to do,
// this needs to be the same for the server and client, easily done with a shared
// module or just copying the struct to both places.
type CustomJob struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func main() {
	// initialize the worker with the generic job type, generic result type,
	// queueURL, the worker authorization key, and the optional QueueWorkerOptions.
	worker, err := client.NewQueueWorker[*CustomJob, string](
		"http://localhost:8888/queue",
		"<worker-authorization-key>",
		nil,
	)
	if err != nil {
		panic(err)
	}

	for {
		// get queued up job from the job channel
		job := worker.GetNewJob()
		
		go func(j *common.Job[*CustomJob, string]) {
			result, err := DoSomething(j.Job)
			if err != nil {
				j.Status = common.Error
				j.Result = err.Error()
			} else {
				j.Status = common.Complete
				// result should be of type R defined in client.NewQueueWorker[J, R]()
				j.Result = result
			}

			// queue up completed job to be sent back to the server
			worker.PostJobResult(j)
		}(job)
	}
}

// do something with the job recieved from the server
func DoSomething(custom *CustomJob) (string, error) {
	if custom.Field1 == "" || custom.Field2 == "" {
		return "", fmt.Errorf("one or more fields empty")
	}
	return custom.Field1 + custom.Field2, nil
}
```
