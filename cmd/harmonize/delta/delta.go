package delta

import (
	"fmt"

	"go.uber.org/zap"
)

type Delta interface {
	Apply(*zap.Logger, Discord) error
}

type DeltaRoleCreate struct {
	RoleName    string
	Color       int
	Permissions int64
}

func (delta DeltaRoleCreate) Apply(logger *zap.Logger, discord Discord) error {
	logger.Info("creating role",
		zap.String("name", delta.RoleName),
		zap.String("color", fmt.Sprintf("%06x", delta.Color)),
		zap.Int64("permissions", delta.Permissions))

	return discord.CreateRole(delta)
}

type DeltaRoleEdit struct {
	RoleID      string
	RoleName    string
	Color       int
	Permissions int64
}

func (delta DeltaRoleEdit) Apply(logger *zap.Logger, discord Discord) error {
	logger.Info("editing role",
		zap.String("id", delta.RoleID),
		zap.String("name", delta.RoleName),
		zap.String("color", fmt.Sprintf("%06x", delta.Color)),
		zap.Int64("permissions", delta.Permissions))

	return discord.EditRole(delta)
}

type DeltaRolePositions []string

func (delta DeltaRolePositions) Apply(logger *zap.Logger, discord Discord) error {
	logger.Info("updating role positions",
		zap.Strings("order", delta))

	return discord.SetRolePositions(delta)
}

type DeltaUserAddRole struct {
	UserID   string
	RoleName string
}

func (delta DeltaUserAddRole) Apply(logger *zap.Logger, discord Discord) error {
	logger.Info("adding user role",
		zap.String("user-id", delta.UserID),
		zap.String("role", delta.RoleName))

	return discord.AddUserRole(delta)
}

type DeltaUserRemoveRole struct {
	UserID   string
	RoleName string
}

func (delta DeltaUserRemoveRole) Apply(logger *zap.Logger, discord Discord) error {
	logger.Info("removing user role",
		zap.String("user-id", delta.UserID),
		zap.String("role", delta.RoleName))

	return discord.RemoveUserRole(delta)
}
