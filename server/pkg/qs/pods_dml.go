package qs

import (
	"encoding/json"
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/types"
	"strconv"
	"strings"
)

// NewPodID : Gets new database sequence
func (qs *QueryServer) NewPodID() (uint64, error) {
	return qs.GetSequence(PodSeqName, 1)
}

func podIDtoKey(id uint64) string {
	return PodPrefix + fmt.Sprintf("%d", id)
}

// GetPOD : Get POD from DB
func (qs *QueryServer) GetPOD(id uint64) (*types.POD, error) {
	podKey := podIDtoKey(id)
	podData, err := qs.Get(podKey)
	if err != nil {
		return nil, err
	}
	pod := types.POD{}
	err = json.Unmarshal(podData, &pod)
	if err != nil {
		return nil, err
	}

	return &pod, nil
}

// UpdateUserPOD : Updates POD and also raises a trigger for listeners
func (qs *QueryServer) UpdateUserPOD(pod *types.POD) (*types.POD, error) {
	podKey := podIDtoKey(pod.ID)
	var username string

	if pod.CreatedBy != nil {
		username = pod.CreatedBy.Name
	}

	err := qs.UpdateAndTrack(podKey, pod, getPodsTracker(), getUserTracker(username), getPodIDTracker(pod.ID))
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// UpsertPOD : Update or Insert POD
func (qs *QueryServer) UpsertPOD(pod *types.POD) (*types.POD, error) {
	podkey := podIDtoKey(pod.ID)

	if qs.KeyExists(podkey) {
		return qs.UpdateUserPOD(pod)
	}

	return qs.InsertPOD(pod)
}

// InsertPOD : Persist User POD info in database and raise triggers
func (qs *QueryServer) InsertPOD(pod *types.POD) (*types.POD, error) {
	var err error
	if pod.ID == 0 {
		pod.ID, err = qs.NewPodID()
		if err != nil {
			return nil, err
		}

		if pod.ID <= 0 {
			return nil, fmt.Errorf("Invalid seq 0")
		}
	}

	podKey := podIDtoKey(pod.ID)
	var username string

	if pod.CreatedBy != nil {
		username = pod.CreatedBy.Name
	}

	err = qs.InsertAndTrack(podKey, pod, getPodsTracker(), getUserTracker(username), getPodIDTracker(pod.ID))
	if err != nil {
		return nil, err
	}

	if err = qs.SaveUserPODIndex(pod); err != nil {
		_ = qs.SoftDelete(podKey)
		return nil, err
	}

	return pod, nil
}

// DeletePOD : Delete pod by given ID
func (qs *QueryServer) DeletePOD(ID uint64) error {
	podKey := podIDtoKey(ID)
	return qs.SoftDelete(podKey)
}

// userpodKey : User POD key
func userPodIndexKey(uname string, id uint64) string {
	return UserPodKeyPrefix + uname + ":" + fmt.Sprintf("%d", id)
}

// SaveUserPODIndex :
func (qs *QueryServer) SaveUserPODIndex(pod *types.POD) error {
	var err error
	if pod.ID <= 0 {
		return fmt.Errorf("Failed to save User POD: Invalid seq 0")
	}

	if pod.CreatedBy == nil {
		return fmt.Errorf("SaveUserPOD: Invalid User")
	}

	upodKey := userPodIndexKey(pod.CreatedBy.Name, pod.ID)

	err = qs.Insert(upodKey, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetPODByUser : Get POD only with user access perm
func (qs *QueryServer) GetPODByUser(user *types.User, id uint64) (*types.POD, error) {

	if user == nil {
		return nil, fmt.Errorf("Invalid User")
	}

	indexKey := userPodIndexKey(user.Name, id)

	if !qs.KeyExists(indexKey) {
		fmt.Println("User does not own the pod and key:", indexKey)
		//	return nil, fmt.Errorf("User does not have permission to delete this pod")
	}

	pod, err := qs.GetPOD(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to locate POD (%d): %v", id, err)
	}

	return pod, nil

}

// ListPODsByUser :
func (qs *QueryServer) ListPODsByUser(user *types.User) (pods []*types.POD, fnerr error) {

	prefix := UserPodKeyPrefix + user.Name + ":"
	keys, err := qs.ListKeys([]byte(prefix))
	if err != nil {
		return
	}

	var podIDs []string

	for _, k := range keys {
		splitfunc := func(c rune) bool {
			return c == ':'
		}
		fields := strings.FieldsFunc(string(k), splitfunc)
		if fields != nil {
			podIDs = append(podIDs, fields[len(fields)-1])
		}
	}

	for _, idstr := range podIDs {
		IDuint64, err := strconv.ParseUint(idstr, 10, 64)

		if err != nil {
			fnerr = err
		}

		p, err := qs.GetPOD(IDuint64)
		if err != nil {
			fnerr = err
		}
		if p != nil {
			pods = append(pods, p)
		}
	}
	return
}

// TrackPodByUser : Returns a channel to track insert/update database changes of a specific POD
func (qs *QueryServer) TrackPodByUser(username string, quit chan int, c chan interface{}) {
	lstnr := qs.GetListener()
	lstnr.RegisterObject(c, getUserTracker(username))
	return
}

// TrackPodChanges : Listen to POD Changes (insert/update) in database
func (qs *QueryServer) TrackPodChanges(quit chan int, c chan interface{}) {
	lstnr := qs.GetListener()
	lstnr.RegisterObject(c, getPodsTracker())
	return
}
