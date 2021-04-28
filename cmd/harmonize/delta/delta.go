package delta

type Delta interface {
	Apply(Discord) error
}

type DeltaRoleCreate struct {
	RoleName    string
	Color       int
	Permissions int64
}

func (delta DeltaRoleCreate) Apply(discord Discord) error {
	return discord.CreateRole(delta)
}

type DeltaRoleEdit struct {
	RoleID      string
	RoleName    string
	Color       int
	Permissions int64
}

func (delta DeltaRoleEdit) Apply(discord Discord) error {
	return discord.EditRole(delta)
}

type DeltaRolePositions []string

func (delta DeltaRolePositions) Apply(discord Discord) error {
	return discord.SetRolePositions(delta)
}

type DeltaUserAddRole struct {
	UserID   string
	RoleName string
}

func (delta DeltaUserAddRole) Apply(discord Discord) error {
	return discord.AddUserRole(delta)
}

type DeltaUserRemoveRole struct {
	UserID   string
	RoleName string
}

func (delta DeltaUserRemoveRole) Apply(discord Discord) error {
	return discord.RemoveUserRole(delta)
}
