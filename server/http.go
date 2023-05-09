package server

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"

	"github.com/colsephiroth/pollingworkermodel/common"
)

func (q *QueueServer[J, R]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(common.WorkerAuthHeader) != q.authorization {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	base := filepath.Base(r.URL.Path)

	if r.Method == http.MethodGet && base == "jobs" {
		q.getNewJobsHandlerFunc(w, r)
		return
	}

	if r.Method == http.MethodPost && base == "results" {
		q.postJobResultsHandlerFunc(w, r)
		return
	}

	http.NotFoundHandler().ServeHTTP(w, r)
}

func (q *QueueServer[J, R]) getNewJobsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(q.NewJobs()); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (q *QueueServer[J, R]) postJobResultsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var job *common.Job[J, R]

	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	q.UpdateJob(job)
}
