package m

import "github.com/spf13/pflag"

type ConfigCommandOptions struct {
	Flags         *pflag.FlagSet
	GetData       func() any
	ModuleName    string
	KeyFlag       string
	GetKey        func() string
	RequiredFlags []string
}
