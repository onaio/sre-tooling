package flags

import "strings"

type StringArray []string

func (stringArray *StringArray) String() string {
	return strings.Join(*stringArray, ",")
}

func (stringArray *StringArray) Set(value string) error {
	*stringArray = append(*stringArray, value)
	return nil
}
