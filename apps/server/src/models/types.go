package models

// Types represent custom enums in postgres

import "database/sql/driver"

// The name of the type as it's stored in postgres
const PG_USER_TYPE_NAME = "user_type"

type UserType string

const (
	PROVIDER  UserType = "PROVIDER"
	REQUESTER UserType = "REQUESTER"
)

func (ct *UserType) Scan(value interface{}) error {
	*ct = UserType(value.(string))
	return nil
}

func (ct UserType) Value() (driver.Value, error) {
	return string(ct), nil
}
