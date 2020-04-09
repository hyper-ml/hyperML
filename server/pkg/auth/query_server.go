package auth

import (
	"github.com/hyper-ml/hyperml/server/pkg/qs"
)

const (

	// SessionIDSeqName :
	SessionIDSeqName = "SESSION_ID_S"
)

type sessionQueryServer struct {
	*qs.QueryServer
}

func newSessionQueryServer(q *qs.QueryServer) *sessionQueryServer {
	return &sessionQueryServer{
		q,
	}
}

func (sqs *sessionQueryServer) NewSessionID() (uint64, error) {
	return sqs.GetSequence(SessionIDSeqName, 1)
}
