package model

type Roles struct {
	RoleID   int    `json:"roleId" db:"role_id"`
	RoleName string `json:"roleName" db:"role_name"`
}
