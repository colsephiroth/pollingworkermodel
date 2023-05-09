package common

type Status string

const (
	New      Status = "new"
	Pending         = "pending"
	Complete        = "complete"
	Error           = "error"
	NotExist        = "not-exist"
)

const WorkerAuthHeader = "X-Worker-Authorization"

type Job[J, R any] struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
	Error  string `json:"error"`
	Job    J      `json:"job"`
	Result R      `json:"result,omitempty"`
}
