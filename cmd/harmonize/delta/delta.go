package delta

type Delta interface {
	Apply(Discord) error
}

type DeltaRoleCreate struct {
	RoleName    string
	Color       int
	Permissions int64
}

func (delta DeltaRoleCreate) Apply(Discord) error {
	return nil
}

type DeltaRoleEdit struct {
	RoleID      string
	Color       int
	Permissions int64
}

func (delta DeltaRoleEdit) Apply(Discord) error {
	return nil
}

type DeltaRolesReorder []string

func (delta DeltaRolesReorder) Apply(Discord) error {
	return nil
}

type DeltaUserAddRole struct {
	UserID   string
	RoleName string
}

func (delta DeltaUserAddRole) Apply(Discord) error {
	return nil
}

type DeltaUserRemoveRole struct {
	UserID   string
	RoleName string
}

func (delta DeltaUserRemoveRole) Apply(Discord) error {
	return nil
}
