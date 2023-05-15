package client

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/colsephiroth/pollingworkermodel/common"
)

type QueueWorker[J, R any] struct {
	queueURL        *url.URL
	authorization   string
	getJobsTickRate time.Duration
	jobChannel      chan common.Job[J, R]
	resultChannel   chan common.Job[J, R]
	httpClient      *http.Client
}

func NewQueueWorker[J, R any](queueURL, authorization string, options *Options) (*QueueWorker[J, R], error) {
	qurl, err := url.Parse(queueURL)
	if err != nil {
		return nil, err
	}

	var opts *Options
	if options != nil {
		opts = options
	} else {
		opts = NewOptions()
	}

	worker := &QueueWorker[J, R]{
		queueURL:      qurl,
		authorization: authorization,

		getJobsTickRate: opts.GetJobsTickRate,
		jobChannel:      make(chan common.Job[J, R], opts.JobChannelBufferSize),
		resultChannel:   make(chan common.Job[J, R], opts.ResultChannelBufferSize),
		httpClient:      opts.HttpClient,
	}

	go worker.getNewJobs()
	go worker.postJobResults()

	return worker, nil
}

func (w *QueueWorker[J, R]) ProcessJobs(f func(J) (R, error)) {
	for {
		job := w.GetNewJob()

		go func(j common.Job[J, R]) {
			result, err := f(j.Job)
			if err != nil {
				j.Status = common.Error
				j.Error = err.Error()
			} else {
				j.Status = common.Complete
				j.Result = result
			}

			w.PostJobResult(j)
		}(job)
	}
}

func (w *QueueWorker[J, R]) GetNewJob() common.Job[J, R] {
	return <-w.jobChannel
}

func (w *QueueWorker[J, R]) PostJobResult(job common.Job[J, R]) {
	w.resultChannel <- job
}

func (w *QueueWorker[J, R]) getNewJobs() {
	ticker := time.NewTicker(w.getJobsTickRate)

	for {
		select {
		case <-ticker.C:
			var jobs []common.Job[J, R]

			req, err := http.NewRequest(http.MethodGet, w.queueURL.String()+"/jobs", nil)
			if err != nil {
				log.Println(err)
				continue
			}

			req.Header.Set(common.WorkerAuthHeader, w.authorization)

			res, err := w.httpClient.Do(req)
			if err != nil {
				log.Println(err)
				continue
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				log.Println(err)
				continue
			}

			if err := res.Body.Close(); err != nil {
				log.Println(err)
			}

			if err := json.Unmarshal(b, &jobs); err != nil {
				log.Printf("%s: %s", err, b)
				continue
			}

			for _, job := range jobs {
				w.jobChannel <- job
			}
		}
	}
}

func (w *QueueWorker[J, R]) postJobResults() {
	for {
		select {
		case job := <-w.resultChannel:
			b, err := json.Marshal(job)
			if err != nil {
				log.Println(err)
				continue
			}

			req, err := http.NewRequest(http.MethodPost, w.queueURL.String()+"/results", bytes.NewReader(b))
			if err != nil {
				log.Println(err)
				continue
			}

			req.Header.Set(common.WorkerAuthHeader, w.authorization)

			_, err = w.httpClient.Do(req)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
