package enums

import (
	"bytes"
	"encoding/json"
	"strings"
)

type ACHAccountType int

const (
	NoneACHAccountType ACHAccountType = iota
	Checking
	Savings
)

func (a ACHAccountType) String() string {
	return [...]string{"", "checking", "saving"}[a]
}

func (a ACHAccountType) EnumIndex() int {
	return int(a)
}

func GetACHAccountTypeEnum(accountType string) ACHAccountType {
	switch strings.ToLower(accountType) {
	case "checking":
		return Checking
	case "savings":
		return Savings
	}
	return NoneACHAccountType
}

var toACHAccountTypeId = map[string]ACHAccountType{
	"checking": Checking,
	"savings":  Savings,
}

var toACHAccountTypeString = map[ACHAccountType]string{
	Checking: "checking",
	Savings:  "savings",
}

func (c ACHAccountType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toACHAccountTypeString[c])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (a *ACHAccountType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*a = toACHAccountTypeId[j]
	return nil
}
