package rest

import (
	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	meta "github.com/hyper-ml/hyperml/server/pkg/types"
	"github.com/hyper-ml/hyperml/worker/utils/structs"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (h *Handler) handleCreateContainerImage() error {
	var err error
	var imgInfo ContainerImageInfo

	if err := ValidateMethod(h.rq.Method, "POST"); err != nil {
		return err
	}

	rawInput, err := ioutil.ReadAll(h.getReqBody())
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to read JSON inputbytes")
	}

	if err := json.Unmarshal(rawInput, &imgInfo); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to unmarshal JSON input")
	}

	if errStr := validateContainerImageInfo(&imgInfo); errStr != "" {
		return base.HTTPErrorf(http.StatusPreconditionFailed, errStr)
	}

	cImage, err := h.server.qs.InsertContainerImage(imgInfo.ContainerImage)
	if err != nil {
		base.Error("Failed to Create container image: %v", err.Error())
		return raisePreconditionFailed("Failed to Create container image")
	}

	imgInfo = newContainerImageInfo(cImage)
	response := structs.Map(imgInfo)
	h.writeJSON(response)

	return nil
}

func (h *Handler) handleListContainerImage() error {
	var err error
	var cImages []meta.ContainerImage
	var response map[string]interface{}
	if err := ValidateMethod(h.rq.Method, "GET"); err != nil {
		return err
	}

	cImages, err = h.server.qs.ListContainerImages()
	if err != nil {
		base.Error("Failed to list container image: %v", err.Error())
		return raisePreconditionFailed("Failed to list container image")
	}

	response = map[string]interface{}{
		"ContainerImages": cImages,
	}

	h.writeJSON(response)
	return nil
}

func (h *Handler) handleDeleteContainerImage() error {
	var err error

	var response map[string]interface{}

	if err := ValidateMethod(h.rq.Method, "DELETE"); err != nil {
		return err
	}

	idStr, _ := h.getMandatoryURLParam("id")
	id, err := strconv.ParseUint(idStr, 0, 64)
	if err != nil {
		return raisePreconditionFailed("Invalid resource profile ID: " + idStr)
	}

	if err := h.server.qs.DeleteResourceProfile(id); err != nil {
		return raisePreconditionFailed("Failed to delete resource profile")
	}

	h.writeJSON(response)
	return nil
}
