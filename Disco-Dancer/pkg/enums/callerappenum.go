package enums

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type CallerApp int

const (
	NoneCallerApp CallerApp = iota
	ATLAS
	MCP
)

func (c CallerApp) String() string {
	return [...]string{"", "atlas", "mcp"}[c]
}

func (c CallerApp) EnumIndex() int {
	return int(c)
}

func GetCallerAppEnum(callerApp string) CallerApp {
	switch strings.ToLower(callerApp) {
	case "atlas":
		return ATLAS
	case "mcp":
		return MCP
	}
	return NoneCallerApp
}

var toCallerAppId = map[string]CallerApp{
	"atlas": ATLAS,
	"mcp":   MCP,
}

var toCallerAppString = map[CallerApp]string{
	ATLAS: "atlas",
	MCP:   "mcp",
}

func (c CallerApp) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toCallerAppString[c])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (c *CallerApp) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*c = toCallerAppId[strings.ToLower(j)]
	return nil
}

// Helps showing available enum values for ProductType on Error Message
func (CallerApp) GetAllowedEnumValues() string {
	values := make([]string, 0, len(toCallerAppString))
	for _, value := range toCallerAppString {
		values = append(values, value)
	}

	sort.Strings(values) // sort values in increasing order

	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	default:
		return strings.Join(values[:len(values)-1], ", ") + ", and " + values[len(values)-1]
	}
}

func (c *CallerApp) IsValidCallerApp() error {
	if *c == NoneCallerApp {
		return fmt.Errorf("caller app should be %s", c.GetAllowedEnumValues())
	}
	return nil
}
