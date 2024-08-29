package enums

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type ProductType int

const (
	NoneProductType ProductType = iota
	PPA
	Commercial
)

var toProductTypeId = map[string]ProductType{
	"ppa":        PPA,
	"commercial": Commercial,
}

var toProductTypeString = map[ProductType]string{
	PPA:        "ppa",
	Commercial: "commercial",
}

func (p ProductType) String() string {
	return toProductTypeString[p]
}

func (p ProductType) EnumIndex() int {
	return int(p)
}

func (p ProductType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toProductTypeString[p])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *ProductType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*p = toProductTypeId[strings.ToLower(j)]
	return nil
}

// Helps showing available enum values for ProductType on Error Message
func (p *ProductType) GetAllowedEnumValues() string {
	values := make([]string, 0, len(toProductTypeString))
	for _, value := range toProductTypeString {
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

func (p ProductType) IsValidProductType() error {
	if p == NoneProductType {
		return fmt.Errorf("product type should be %s", p.GetAllowedEnumValues())
	}
	return nil
}
