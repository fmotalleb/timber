package config

import (
	"errors"
	"reflect"
	"strings"
)

// User represents a user with their credentials and access rights.
type User struct {
	Name       string   `mapstructure:"name"`
	Password   string   `mapstructure:"password"`
	AccessList []string `mapstructure:"access"`
}

// Decode is a custom decoder for the User type to handle string format.
// The format is <user>:<pass>@<access-1>,<access-2>
func (u *User) Decode(from reflect.Type, val interface{}) (any, error) {
	if from.Kind() != reflect.String {
		return val, nil
	}
	raw, ok := val.(string)
	if !ok {
		return val, errors.New("expected string value for user object")
	}

	// <user>:<pass>@<access-1>,<access-2>
	atIdx := strings.IndexByte(raw, '@')
	if atIdx == -1 {
		return nil, errors.New("invalid user format: missing '@'")
	}

	credPart := raw[:atIdx]
	accessPart := raw[atIdx+1:]

	colIdx := strings.IndexByte(credPart, ':')
	if colIdx == -1 {
		return nil, errors.New("invalid user format: missing ':'")
	}

	username := credPart[:colIdx]
	password := credPart[colIdx+1:]

	if username == "" || password == "" {
		return nil, errors.New("username and password must not be empty")
	}

	var access []string
	if accessPart != "" {
		for _, a := range strings.Split(accessPart, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				access = append(access, a)
			}
		}
	}
	u.Name = username
	u.Password = password
	u.AccessList = access
	return u, nil
}
