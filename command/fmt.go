package command

import (
	"fmt"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/posener/complete"
)

const (
	stdinArg = "-"
)

type FormatCommand struct {
	Meta
	list      bool
	write     bool
	diff      bool
	check     bool
	recursive bool
}

func (*FormatCommand) Help() string {
	helpText := `
Usage: nomad fmt [options] [target...]

  Rewrites all Nomad configuration files (.nomad or .hcl) to a canonical format.

  By default, fmt scans the current directory for configuration files. If you
  provide a directory for the target argument, then fmt will scan that directory
  instead. If you provide a file, then fmt will process just that file. If you
  provide a single dash ("-"), then fmt will read from standard input (STDIN).

  The content must be in the HCL2 language native syntax; JSON is not supported.

General Options:

  ` + generalOptionsUsage(usageOptsDefault) + `

Format Options:

  -list=false
    Don't list files whose formatting differs (always disabled if using STDIN)

  -check
    Check if the input is formatted. Exit status will be 0 if all input is properly
    formatted and non-zero otherwise.

  -diff
    Display diffs of formatting change

  -write=false
    Don't write to source files (always disabled if using -check)

  -recursive
    Also process files in subdirectories. By default, only the given directory (or
    current directory) is processed.
`

	return strings.TrimSpace(helpText)
}

func (*FormatCommand) Synopsis() string {
	return "Rewrites HCL2 config files to canonical format"
}

func (*FormatCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictOr(
		complete.PredictDirs("*"),
		complete.PredictFiles("*.nomad"),
		complete.PredictFiles("*.hcl"),
	)
}

func (*FormatCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-list":      complete.PredictNothing,
		"-check":     complete.PredictNothing,
		"-diff":      complete.PredictNothing,
		"-write":     complete.PredictNothing,
		"-recursive": complete.PredictNothing,
	}
}

func (c *FormatCommand) Run(args []string) int {
	flagSet := c.Meta.FlagSet("fmt", FlagSetClient)
	flagSet.Usage = func() { c.Ui.Output(c.Help()) }
	flagSet.BoolVar(&c.list, "list", true, "")
	flagSet.BoolVar(&c.check, "check", false, "")
	flagSet.BoolVar(&c.diff, "diff", false, "")
	flagSet.BoolVar(&c.write, "write", true, "")
	flagSet.BoolVar(&c.recursive, "recursive", false, "")

	if err := flagSet.Parse(args); err != nil {
		return 255
	}

	args = flagSet.Args()

	var paths []string
	if len(args) == 0 {
		paths = []string{"."}
	} else if args[0] == stdinArg {
		c.list = false
		c.write = false
	} else {
		paths = args
	}

	c.Ui.Output(fmt.Sprintf("args: %+v", c))
	c.Ui.Output(fmt.Sprintf("file paths: %+v", paths))

	diags := c.fmt(paths)
	if diags.HasErrors() {
		mErr := multierror.Append(nil, diags.Errs()...)
		c.Ui.Error(fmt.Sprintf("Error formatting input: %+v", mErr))
		return 2
	}

	// if check, set exit code and exit.
	// if list, output modified file names.
	return 0
}

func (c *FormatCommand) fmt(paths []string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// do work

	return diags
}
