package utils

import (
	"strconv"
	"strings"

	. "github.com/ncpa0cpl/convenient-structures"
)

type ParsedArgs struct {
	Input       string
	NamedParams *Map[string, string]
}

func (args *ParsedArgs) GetParam(paramName string, defaultValue string) string {
	if v, ok := args.NamedParams.Get(paramName); ok {
		return v
	}
	return defaultValue
}

func (args *ParsedArgs) GetParamInt(paramName string, defaultValue int) int {
	if v, ok := args.NamedParams.Get(paramName); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func (args *ParsedArgs) GetParamUint64(paramName string, defaultValue uint64) uint64 {
	if v, ok := args.NamedParams.Get(paramName); ok {
		if i, err := strconv.ParseUint(v, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func removePrefix(argName string) string {
	if strings.HasPrefix(argName, "--") {
		return argName[2:]
	}
	return argName[1:]
}

func ParseArgs(args []string) ParsedArgs {
	results := ParsedArgs{
		Input:       "",
		NamedParams: NewMap(map[string]string{}),
	}

	for i := 0; i < len(args); i += 1 {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			vIdx := i + 1
			if vIdx < len(args) && !strings.HasPrefix(args[vIdx], "-") {
				results.NamedParams.Set(removePrefix(arg), args[vIdx])
				i += 1
			} else {
				results.NamedParams.Set(removePrefix(arg), "")
			}
		} else {
			if results.Input == "" {
				results.Input = arg
			}
		}

	}

	return results
}
