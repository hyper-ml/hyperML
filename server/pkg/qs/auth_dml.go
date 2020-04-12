package qs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyper-ml/hyperml/server/pkg/types"
)

const (
	// UserPrefix : Prefix used as database key for user
	UserPrefix = "user:"

	// UserAPIKeyPrefix : User and APIKey index
	UserAPIKeyPrefix = "userAPIKeyIndex:"

	// sep :
	sep = ":"

	// UserAlreadyExists : Error message when user already exists
	UserAlreadyExists = "user name already exists"

	// EmptyPasswordHash : Error Message when password is empty
	EmptyPasswordHash = "empty password hash"
)

// ErrEmptyPasswordHash : Error raised when password is empty
func ErrEmptyPasswordHash() error {
	return fmt.Errorf(EmptyPasswordHash)
}

// ErrUserNameAlreadyExists : Error raised when user name exists
func ErrUserNameAlreadyExists() error {
	return fmt.Errorf(UserAlreadyExists)
}

// InsertUserAttrs : Insert User Attributes
func (qs *QueryServer) InsertUserAttrs(name string, attrs *types.UserAttrs) error {

	if qs.CheckUserExists(name) {
		return ErrUserNameAlreadyExists()
	}

	userKey := userKey(name)
	// check if already exists

	if attrs.PasswordHash == "" {
		return ErrEmptyPasswordHash()
	}

	err := qs.Insert(userKey, attrs)
	if err != nil {
		return err
	}

	if attrs.APIKey != nullString {
		return qs.InsertIndex(userAPIKey(attrs.APIKey, attrs.Name))
	}

	return nil
}

func userAPIKey(apiKey, uname string) string {
	return UserAPIKeyPrefix + apiKey + ":" + uname
}

// UpdateUserAttrs : Update User Attributes
func (qs *QueryServer) UpdateUserAttrs(name string, attrs *types.UserAttrs) (*types.UserAttrs, error) {

	userKey := userKey(name)

	if attrs.PasswordHash == "" {
		return nil, ErrEmptyPasswordHash()
	}

	err := qs.Update(userKey, attrs)
	if err != nil {
		return attrs, err
	}

	if attrs.APIKey != nullString {
		return attrs, qs.InsertIndex(userAPIKey(attrs.APIKey, attrs.Name))
	}

	return attrs, nil
}

func userKey(name string) string {
	return UserPrefix + name
}

// CheckUserExists : chck if user exists in DB
func (qs *QueryServer) CheckUserExists(name string) bool {
	userKey := userKey(name)
	return qs.KeyExists(userKey)
}

// GetUserAttrs : Get user info from DB
func (qs *QueryServer) GetUserAttrs(name string) (*types.UserAttrs, error) {
	userKey := userKey(name)
	rawAttrs, err := qs.Get(userKey)
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

// GetUserByAPIKey : Get User by API Key
func (qs *QueryServer) GetUserByAPIKey(apiKey string) (*types.UserAttrs, error) {
	prefix := UserAPIKeyPrefix + apiKey + ":"

	keys, err := qs.ListKeys([]byte(prefix))
	if err != nil {
		return nil, fmt.Errorf("User not found")
	}

	var usernames []string
	for _, k := range keys {
		splitfunc := func(c rune) bool {
			return c == ':'
		}
		fields := strings.FieldsFunc(string(k), splitfunc)
		if fields != nil {
			usernames = append(usernames, fields[len(fields)-1])
		}
	}

	if len(usernames) > 1 {
		return nil, fmt.Errorf("You can not use with this API Key")
	}

	if len(usernames) == 0 {
		return nil, fmt.Errorf("User name against this API key does not exist")
	}

	return qs.GetUserAttrs(usernames[0])
}

func userDiskKey(name, diskName string) string {
	return "user_disk:" + name + sep + diskName
}

// CheckUserDiskExists : chck if user disk exists in DB
func (qs *QueryServer) CheckUserDiskExists(name, diskName string) bool {
	userDiskKey := userDiskKey(name, diskName)
	return qs.KeyExists(userDiskKey)
}

// InsertUserDisk :
func (qs *QueryServer) InsertUserDisk(pd *types.UserPersistentDisk) (*types.UserPersistentDisk, error) {
	dname := pd.Disk.Name
	size := pd.Disk.Size
	uname := pd.User.Name

	if uname == "" || !qs.CheckUserExists(uname) {
		return nil, ErrUserDoesntExist()
	}

	if dname == "" {
		return nil, ErrDiskNameNull()
	}

	if size == 0 {
		return nil, ErrDiskSizeZero()
	}

	if qs.CheckUserDiskExists(uname, dname) {
		return nil, ErrUserDiskAlreadyExists()
	}

	dKey := userDiskKey(uname, dname)
	err := qs.Insert(dKey, pd)
	if err != nil {
		return nil, err
	}

	return pd, nil

}

// GetUserDisk :
func (qs *QueryServer) GetUserDisk(uname, diskname string) (*types.UserPersistentDisk, error) {
	if uname == "" || !qs.CheckUserExists(uname) {
		return nil, ErrUserDoesntExist()
	}

	if diskname == "" {
		return nil, ErrDiskNameNull()
	}

	dKey := userDiskKey(uname, diskname)
	pdData, err := qs.Get(dKey)
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
