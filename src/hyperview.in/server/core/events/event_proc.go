package events


type EventType int


const (
  RecordCreated EventType = iota
  RecordUpdated 
  RecordDeleted 
)

type Event struct {
  Key string
  Value []byte

  // hold prior event value if available
  ParentKey string
  ParentValue []byte 
  Type EventType
}


type EventProc interface {
  Receive() <- chan *Event
  Watch() 
  Close()
}











