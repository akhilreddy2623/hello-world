package enums

import (
	"strings"
)

type ComponentType int

const (
	NoneComponentType ComponentType = iota
	Administrator
	Executor
	Vault
)

func (c ComponentType) String() string {
	return [...]string{
		"",
		"administrator",
		"executor",
		"vault"}[c]
}

func (c ComponentType) EnumIndex() int {
	return int(c)
}

func GetComponentTypeEnum(componentType string) ComponentType {
	switch strings.ToLower(componentType) {
	case "administrator":
		return Administrator
	case "executor":
		return Executor
	case "vault":
		return Vault
	}
	return NoneComponentType
}
