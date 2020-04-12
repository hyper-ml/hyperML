package schedules

import (
	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/qs"
)

// NbSchedulerKey : Notebook Scheduler Queue Database key
const NbSchedulerKey = "NBScheduler:Queues"

// qhandler : query handler for schedules package
type qHandler struct {
	*qs.QueryServer
}

func newQueryHandler(q *qs.QueryServer) *qHandler {
	return &qHandler{
		q,
	}
}

func (qh *qHandler) UpsertSchedulerQueues(nbs *NotebookScheduler) error {
	data, err := json.Marshal(nbs)
	if err != nil {
		return err
	}
	return qh.Upsert(NbSchedulerKey, data)
}

func (qh *qHandler) GetSchedulerQueues() (*NotebookScheduler, error) {
	data, err := qh.Get(NbSchedulerKey)
	if err != nil {
		return nil, err
	}
	nbs := NotebookScheduler{}
	err = json.Unmarshal(data, &nbs)
	if err != nil {
		return nil, err
	}

	return &nbs, nil
}
