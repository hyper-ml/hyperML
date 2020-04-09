package qs

import (
	"encoding/json"
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

func resProfileIDToKey(rpid uint64) string {
	return ResourceProfilePrefix + fmt.Sprintf("%d", rpid)
}

// InsertResourceProfile :
func (qs *QueryServer) InsertResourceProfile(rp *types.ResourceProfile) (*types.ResourceProfile, error) {
	rpSeq, err := qs.newResourceProfileID()
	if err != nil {
		return nil, err
	}
	if rpSeq <= 0 {
		return nil, fmt.Errorf("Invalid seq 0")
	}

	rpKey := resProfileIDToKey(rpSeq)
	rp.ID = rpSeq

	fmt.Println("rpKey:", rpKey)

	err = qs.Insert(rpKey, rp)
	if err != nil {
		return nil, err
	}
	return rp, nil

}

// GetResourceProfile  : Get Resource Profile from DB
func (qs *QueryServer) GetResourceProfile(ID uint64) (*types.ResourceProfile, error) {
	profkey := resProfileIDToKey(ID)
	rawData, err := qs.Get(profkey)

	if err != nil {
		return nil, err
	}

	rprof := types.ResourceProfile{}
	err = json.Unmarshal(rawData, &rprof)
	if err != nil {
		return nil, err
	}

	return &rprof, nil
}

// ListResourceProfiles : List of Resource Profiles
func (qs *QueryServer) ListResourceProfiles() (rProfiles []types.ResourceProfile, fnerr error) {
	data, err := qs.List([]byte(ResourceProfilePrefix))
	if err != nil {
		fnerr = err
	}

	for _, rawbytes := range data {
		rp := types.ResourceProfile{}
		if err := json.Unmarshal(rawbytes, &rp); err != nil {
			base.Errorf(err.Error())
		}
		rProfiles = append(rProfiles, rp)
	}

	return
}

// DeleteResourceProfile : Delete resource profile with given name
func (qs *QueryServer) DeleteResourceProfile(id uint64) error {
	rpKey := resProfileIDToKey(id)
	err := qs.SoftDelete(rpKey)
	return err
}

// DisableResourceProfile  : Disable Resource Profile in DB
func (qs *QueryServer) DisableResourceProfile(ID uint64) (*types.ResourceProfile, error) {
	profkey := resProfileIDToKey(ID)
	rawData, err := qs.Get(profkey)

	if err != nil {
		return nil, err
	}

	rprof := types.ResourceProfile{}
	err = json.Unmarshal(rawData, &rprof)
	if err != nil {
		return nil, err
	}

	rprof.Disabled = true
	err = qs.Update(profkey, rprof)
	if err != nil {
		return nil, err
	}

	return &rprof, nil
}

// UpsertResourceProfile : Merges the prof record
func (qs *QueryServer) UpsertResourceProfile(input *types.ResourceProfile) (*types.ResourceProfile, error) {

	if !types.IsValidResourceProfile(input) {
		return nil, fmt.Errorf("Invalid Profile Params")
	}

	if input.ID != 0 {
		profKey := resProfileIDToKey(input.ID)
		return input, qs.Upsert(profKey, input)
	}

	return qs.InsertResourceProfile(input)
}
