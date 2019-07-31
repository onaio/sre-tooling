package cli

type Command interface {
	Init(helpFlagName string, helpFlagDescription string)
	ParseArgs(args []string)
	GetName() string
	GetDescription() string
}
