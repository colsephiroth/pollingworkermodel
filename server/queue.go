package server

import (
	"github.com/alphadose/haxmap"
	"github.com/colsephiroth/pollingworkermodel/common"
	"github.com/teris-io/shortid"
)

type QueueServer[J, R any] struct {
	queue         *haxmap.Map[string, *common.Job[J, R]]
	authorization string
}

func NewQueueServer[T, R any](authorization string) *QueueServer[T, R] {
	return &QueueServer[T, R]{
		queue:         haxmap.New[string, *common.Job[T, R]](),
		authorization: authorization,
	}
}

func (q *QueueServer[J, R]) NewJob(job J) string {
	var result R

	newJob := &common.Job[J, R]{
		ID:     shortid.MustGenerate(),
		Status: common.New,
		Job:    job,
		Result: result,
	}

	q.AddJob(newJob)

	return newJob.ID
}

func (q *QueueServer[J, R]) AddJob(job *common.Job[J, R]) {
	q.queue.Set(job.ID, job)
}

func (q *QueueServer[J, R]) RemoveJob(job *common.Job[J, R]) {
	q.queue.Del(job.ID)
}

func (q *QueueServer[J, R]) GetJob(id string) (*common.Job[J, R], bool) {
	return q.queue.Get(id)
}

func (q *QueueServer[J, R]) UpdateJob(job *common.Job[J, R]) {
	q.queue.Set(job.ID, job)
}

func (q *QueueServer[J, R]) CheckJob(id string) common.Status {
	job, ok := q.GetJob(id)
	if !ok {
		return common.NotExist
	}

	return job.Status
}

func (q *QueueServer[J, R]) WaitJob(id string) (*common.Job[J, R], bool) {
	for {
		status := q.CheckJob(id)

		if status == common.NotExist {
			return nil, false
		}

		if status == common.Complete || status == common.Error {
			return q.GetJob(id)
		}
	}
}

func (q *QueueServer[J, R]) AddWaitJob(content J) *common.Job[J, R] {
	job, _ := q.WaitJob(q.NewJob(content))

	return job
}

func (q *QueueServer[J, R]) NewJobs() []*common.Job[J, R] {
	var jobs []*common.Job[J, R]

	q.queue.ForEach(func(k string, j *common.Job[J, R]) bool {
		if j.Status == common.New {
			jobs = append(jobs, j)
			j.Status = common.Pending
			q.UpdateJob(j)
		}

		return true
	})

	return jobs
}
