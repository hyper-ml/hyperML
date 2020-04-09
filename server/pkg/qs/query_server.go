package qs

import (
	"encoding/json"
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	db_pkg "github.com/hyper-ml/hyperml/server/pkg/db"
	"github.com/hyper-ml/hyperml/server/pkg/types"
	"strconv"
	"strings"
)

// QueryServer : Global query server for all metadata objects
type QueryServer struct {
	db_pkg.DatabaseContext
}

// NewQueryServer : A generic query handler for metadata
func NewQueryServer(db db_pkg.DatabaseContext) *QueryServer {
	qs := &QueryServer{}
	qs.DatabaseContext = db
	return qs
}

func (qs *QueryServer) newResourceGroupID() (id uint64, fnerr error) {
	return qs.GetSequence(ResourceGroupSeqName, 1)
}

func (qs *QueryServer) newResourceProfileID() (uint64, error) {
	return qs.GetSequence(ResourceProfileSeqName, 1)
}

func (qs *QueryServer) newContainerImageID() (uint64, error) {
	return qs.GetSequence(ContainerImageSeqName, 1)
}

func (qs *QueryServer) resourceGroupKey(rgid uint64) string {
	return ResourceGroupPrefix + fmt.Sprintf("%d", rgid)
}

// InsertResourceGroup :
func (qs *QueryServer) InsertResourceGroup(rg *types.ResourceGroup) (*types.ResourceGroup, error) {
	rgSeq, err := qs.newResourceGroupID()
	if err != nil {
		return nil, err
	}
	if rgSeq <= 0 {
		return nil, fmt.Errorf("Invalid seq 0")
	}

	rgKey := qs.resourceGroupKey(rgSeq)
	rg.ID = rgSeq
	err = qs.Insert(rgKey, rg)
	if err != nil {
		return nil, err
	}

	return rg, nil

}

func (qs *QueryServer) containerImageKey(rpid uint64) string {
	return ContainerImagePrefix + fmt.Sprintf("%d", rpid)
}

func (qs *QueryServer) containerImgNameIndexKey(name string, rpid uint64) string {
	return ContainerImageIndexPrefix + ":" + name + ":" + fmt.Sprintf("%d", rpid)
}

// InsertContainerImage :
func (qs *QueryServer) InsertContainerImage(ci *types.ContainerImage) (*types.ContainerImage, error) {

	imgSeq, err := qs.newContainerImageID()
	if err != nil {
		return nil, err
	}

	if imgSeq <= 0 {
		return nil, fmt.Errorf("Invalid seq 0")
	}

	imgKey := qs.containerImageKey(imgSeq)
	ci.ID = imgSeq

	err = qs.Insert(imgKey, ci)
	if err != nil {
		return nil, err
	}

	// create name index
	indexKey := qs.containerImgNameIndexKey(ci.Name, ci.ID)
	err = qs.Insert(indexKey, nil)
	if err != nil {
		_ = qs.Delete(imgKey)
		return nil, err
	}

	return ci, nil
}

// GetOrCreateContainerImage : Gets existing record or creates new
func (qs *QueryServer) GetOrCreateContainerImage(param *types.ContainerImage) (*types.ContainerImage, error) {
	if param.Name == "" {
		return nil, fmt.Errorf("Image Name can not be empty")
	}

	ci, err := qs.GetContainerImageByName(param.Name)
	if err != nil {
		return nil, err
	}

	if ci != nil {
		return ci, err
	}

	return qs.InsertContainerImage(param)
}

// GetContainerImageByName  : Get Container Image by Name from DB Index
func (qs *QueryServer) GetContainerImageByName(name string) (*types.ContainerImage, error) {
	if name == "" {
		return nil, fmt.Errorf("Image Name can not be empty")
	}

	nameIndexPrefix := ContainerImageIndexPrefix + ":" + name + ":"
	keys, err := qs.ListKeys([]byte(nameIndexPrefix))
	if err != nil {
		return nil, err
	}

	// Container Image with given name does not exist
	if keys == nil {
		return nil, nil
	}

	var containerImageid uint64

	for _, k := range keys {
		fmt.Println("keys:", string(k))
		splitfunc := func(c rune) bool {
			return c == ':'
		}

		fields := strings.FieldsFunc(string(k), splitfunc)
		if fields != nil {
			containerImageid, _ = strconv.ParseUint(fields[len(fields)-1], 10, 64)
		}
	}

	if containerImageid != 0 {
		return qs.GetContainerImage(containerImageid)
	}

	return nil, nil
}

// GetContainerImage  : Get Container Image from DB
func (qs *QueryServer) GetContainerImage(ID uint64) (*types.ContainerImage, error) {
	imgkey := qs.containerImageKey(ID)
	rawData, err := qs.Get(imgkey)

	if err != nil {
		return nil, err
	}

	cimg := types.ContainerImage{}
	err = json.Unmarshal(rawData, &cimg)
	if err != nil {
		return nil, err
	}

	return &cimg, nil
}

// ListContainerImages : List of Container Images
func (qs *QueryServer) ListContainerImages() (cImages []types.ContainerImage, fnerr error) {
	data, err := qs.List([]byte(ContainerImagePrefix))
	if err != nil {
		fnerr = err
	}

	for _, rawbytes := range data {
		ci := types.ContainerImage{}
		if err := json.Unmarshal(rawbytes, &ci); err != nil {
			base.Errorf(err.Error())
		}
		cImages = append(cImages, ci)
	}

	return
}

// DeleteContainerImage : Delete container image with given id
func (qs *QueryServer) DeleteContainerImage(id uint64) error {
	ciKey := qs.containerImageKey(id)
	err := qs.SoftDelete(ciKey)
	return err
}
