package seed

import (
	"encoding/json"
	"fmt"
	authpkg "github.com/hyper-ml/hyperml/server/pkg/auth"
	"github.com/hyper-ml/hyperml/server/pkg/base"
	"github.com/hyper-ml/hyperml/server/pkg/config"
	"github.com/hyper-ml/hyperml/server/pkg/db"
	qspkg "github.com/hyper-ml/hyperml/server/pkg/qs"
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

// Do : Initiates metadata after fresh install
func Do(conf *config.Config) error {
	dbc, err := db.NewDatabaseContext(conf.DB)
	if err != nil {
		base.Error("Failed to create db context: ", err)
		return err
	}

	qs := qspkg.NewQueryServer(dbc)

	var failures []error
	if err := initAdminUser(qs, conf.NoAuth); err != nil {
		failures = append(failures, err)
	}

	if err := initResourceProfiles(qs); err != nil {
		failures = append(failures, err)
	}

	if err := insertContainerImages(qs); err != nil {
		failures = append(failures, err)
	}

	if len(failures) > 0 {
		for _, f := range failures {
			base.Error(f.Error())
		}
		return fmt.Errorf("some failures occurred while seeding data")
	}
	return nil
}

func initAdminUser(qs *qspkg.QueryServer, noauth bool) error {
	authServer := authpkg.NewAuthServer(qs, noauth)
	user, err := authServer.CreateTypedUser(types.AdminUser, "admin", "admin@email.com", "admin")
	if err != nil {
		return fmt.Errorf("Failed to create admin user:" + err.Error())
	}
	userj, _ := json.Marshal(user)
	base.Out("User " + user.Name + " created")
	base.Out(string(userj))
	base.Out("*****")
	return nil
}

func initResourceProfiles(qs *qspkg.QueryServer) error {

	unlimited := &types.ResourceProfile{
		Name:      "Unlimited",
		Subtitle:  "Unlimited Resource Plan",
		ShortDesc: "This resource plan places no limits. Requests may wait if server does not have sufficient resources.",
		LongDesc:  "This resource plan places no limits. Requests may wait if server does not have sufficient resources.",
		CPU:       "0",
		GPU:       "0",
		GPURam:    "0",
		RAM:       "0",
		Disk:      "0",
	}

	ulPlan, err := qs.InsertResourceProfile(unlimited)
	if err != nil {
		return fmt.Errorf("Failed to create resoure profile:" + err.Error())
	}
	planJSON, _ := json.Marshal(ulPlan)
	base.Out("Resource Profile " + ulPlan.Name + " created")
	base.Out(string(planJSON))
	base.Out("*****")
	return nil
}

func insertContainerImages(qs *qspkg.QueryServer) error {

	images := []*types.ContainerImage{
		&types.ContainerImage{Name: "jupyter/minimal-notebook", DescText: "Jupyter Minimal Notebook"},
	}
	for _, image := range images {
		containerImage, err := qs.InsertContainerImage(image)
		if err != nil {
			base.Error("Failed to create container image (" + image.Name + ") :" + err.Error())
		}
		imageJSON, _ := json.Marshal(containerImage)
		base.Out("Container Image " + containerImage.Name + " created")

		base.Out(string(imageJSON))
		base.Out("*****")
	}
	return nil
}
