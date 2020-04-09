package requests

import (
	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/qs"
	types "github.com/hyper-ml/hyperml/server/pkg/types"
)

const (
	// UserPrefix :
	UserPrefix = "user:"

	// PodPrefix :
	PodPrefix = "pod:"

	// sep :
	sep = ":"
)

// UserDataStore : User data handler for DB storage
type UserDataStore struct {
	*qs.QueryServer
}

// NewUserDataStore : returns user data handler
func NewUserDataStore(q *qs.QueryServer) *UserDataStore {
	return &UserDataStore{
		q,
	}
}

func userKey(name string) string {
	return UserPrefix + name
}

// InsertUserAttrs : Insert user into DB
func (ds *UserDataStore) InsertUserAttrs(name string, userAttts *types.UserAttrs) error {
	if ds.CheckUserExists(name) {
		return ErrUserNameAlreadyExists()
	}

	userKey := userKey(name)
	// check if already exists

	if userAttts.PasswordHash == "" {
		return ErrEmptyPasswordHash()
	}

	err := ds.Insert(userKey, userAttts)
	if err != nil {
		return err
	}

	return nil
}

// CheckUserExists : chck if user exists in DB
func (ds *UserDataStore) CheckUserExists(name string) bool {
	userKey := userKey(name)
	return ds.KeyExists(userKey)
}

// GetUserAttrs : Get user info from DB
func (ds *UserDataStore) GetUserAttrs(name string) (*types.UserAttrs, error) {
	userKey := userKey(name)
	rawAttrs, err := ds.Get(userKey)
	if err != nil {
		return nil, err
	}

	userAttrs := types.UserAttrs{}
	err = json.Unmarshal(rawAttrs, &userAttrs)
	if err != nil {
		return nil, err
	}

	return &userAttrs, nil
}

func userDiskKey(name, diskName string) string {
	return "user_disk:" + name + sep + diskName
}

// CheckUserDiskExists : chck if user disk exists in DB
func (ds *UserDataStore) CheckUserDiskExists(name, diskName string) bool {
	userDiskKey := userDiskKey(name, diskName)
	return ds.KeyExists(userDiskKey)
}

// InsertUserDisk :
func (ds *UserDataStore) InsertUserDisk(pd *types.UserPersistentDisk) (*types.UserPersistentDisk, error) {
	dname := pd.Disk.Name
	size := pd.Disk.Size
	uname := pd.User.Name

	if uname == "" || !ds.CheckUserExists(uname) {
		return nil, ErrUserDoesntExist()
	}

	if dname == "" {
		return nil, ErrDiskNameNull()
	}

	if size == 0 {
		return nil, ErrDiskSizeZero()
	}

	if ds.CheckUserDiskExists(uname, dname) {
		return nil, ErrUserDiskAlreadyExists()
	}

	dKey := userDiskKey(uname, dname)
	err := ds.Insert(dKey, pd)
	if err != nil {
		return nil, err
	}

	return pd, nil

}

// GetUserDisk :
func (ds *UserDataStore) GetUserDisk(uname, diskname string) (*types.UserPersistentDisk, error) {
	if uname == "" || !ds.CheckUserExists(uname) {
		return nil, ErrUserDoesntExist()
	}

	if diskname == "" {
		return nil, ErrDiskNameNull()
	}

	dKey := userDiskKey(uname, diskname)
	pdData, err := ds.Get(dKey)
	if err != nil {
		return nil, err
	} else if pdData == nil {
		return nil, ErrUserDiskDoesntExist()
	}

	pDisk := types.UserPersistentDisk{}

	err = json.Unmarshal(pdData, &pDisk)
	if err != nil {
		return nil, err
	}

	return &pDisk, nil
}
