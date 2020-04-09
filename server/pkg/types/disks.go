package types

// PersistentDisk : Represents Persistent Disks / Volumes
type PersistentDisk struct {
	Name     string
	HostPath string
	Size     uint64
}
