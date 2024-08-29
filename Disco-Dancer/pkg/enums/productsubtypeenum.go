package enums

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type ProductSubType int

const (
	NoneProductSubType ProductSubType = iota
	Auto
	Cycle
	Umbrella
)

var toProductSubTypeId = map[string]ProductSubType{
	"auto":     Auto,
	"cycle":    Cycle,
	"umbrella": Umbrella,
}

var toProductSubTypeString = map[ProductSubType]string{
	Auto:     "auto",
	Cycle:    "cycle",
	Umbrella: "umbrella",
}

func (p ProductSubType) String() string {
	return toProductSubTypeString[p]
}

func (p ProductSubType) EnumIndex() int {
	return int(p)
}

func (p ProductSubType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toProductSubTypeString[p])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *ProductSubType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*p = toProductSubTypeId[strings.ToLower(j)]
	return nil
}

// Helps showing available enum values for ProductSubType on Error Message
func (p *ProductSubType) GetAllowedEnumValues() string {
	values := make([]string, 0, len(toProductSubTypeString))
	for _, value := range toProductSubTypeString {
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

func (p ProductSubType) IsValidProductSubType() error {
	if p == NoneProductSubType {
		return fmt.Errorf("product sub type should be %s", p.GetAllowedEnumValues())
	}
	return nil
}
