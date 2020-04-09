package rest

import (
	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	meta "github.com/hyper-ml/hyperml/server/pkg/types"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) handleCreateResourceProfile() error {
	var err error
	var rProfInfo ResourceProfInfo

	if err := ValidateMethod(h.rq.Method, "POST"); err != nil {
		return err
	}

	rawInput, err := ioutil.ReadAll(h.getReqBody())
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to read JSON inputbytes")
	}

	if err := json.Unmarshal(rawInput, &rProfInfo); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to unmarshal JSON input")
	}

	if errStr := validateResourceProfInfo(&rProfInfo); errStr != "" {
		return base.HTTPErrorf(http.StatusPreconditionFailed, errStr)
	}

	rprof, err := h.server.qs.InsertResourceProfile(rProfInfo.ResourceProfile)
	if err != nil {
		base.Error("Failed to Create Resource Profile: %v", err.Error())
		return raisePreconditionFailed("Failed to Create Resource Profile")
	}

	profInfo := newResourceProfInfo(rprof)
	h.writeJSON(profInfo)

	return nil
}

func (h *Handler) handleUpdateResourceProfiles() error {
	var profs ResourceProfsInfo
	var errors []string

	if err := ValidateMethod(h.rq.Method, "PUT"); err != nil {
		return err
	}

	rawInput, err := ioutil.ReadAll(h.getReqBody())
	if err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to read JSON inputbytes")
	}

	if err := json.Unmarshal(rawInput, &profs); err != nil {
		return base.HTTPErrorf(http.StatusBadRequest, "Failed to unmarshal JSON input")
	}

	for _, prof := range profs {
		if prof.Name != nullString {
			_, err := h.server.qs.UpsertResourceProfile(&prof)
			if err != nil {
				errors = append(errors, err.Error())
			}
		} else {
			errors = append(errors, "Resource Profile with no name \n")
		}

	}

	if len(errors) > 0 {
		return raisePreconditionFailed("Some requests failed: " + strings.Join(errors, ";"))
	}

	return nil
}

func (h *Handler) handleListResourceProfile() error {
	var err error
	var rProfiles []meta.ResourceProfile
	var response map[string]interface{}
	if err := ValidateMethod(h.rq.Method, "GET"); err != nil {
		return err
	}

	rProfiles, err = h.server.qs.ListResourceProfiles()
	if err != nil {
		base.Error("Failed to list Resource Profile: %v", err.Error())
		return raisePreconditionFailed("Failed to list Resource Profile")
	}

	response = map[string]interface{}{
		"ResourceProfiles": rProfiles,
	}

	h.writeJSON(response)
	return nil
}

func (h *Handler) handleDeleteResourceProfile() error {
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

func (h *Handler) handleDisableResourceProfile() error {
	var err error
	if err := ValidateMethod(h.rq.Method, "PUT"); err != nil {
		return err
	}

	idStr, _ := h.getMandatoryURLParam("id")
	id, err := strconv.ParseUint(idStr, 0, 64)
	if err != nil {
		return raisePreconditionFailed("Invalid resource profile ID: " + idStr + ":" + err.Error())
	}

	updated, err := h.server.qs.DisableResourceProfile(id)
	if err != nil {
		return raisePreconditionFailed("Failed to delete resource profile")
	}

	profInfo := newResourceProfInfo(updated)

	h.writeJSON(profInfo)
	return nil
}
