package command

import (
	"fmt"
	"io"
	"os"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/posener/complete"
)

const (
	stdinArg = "-"
)

type FmtCommand struct {
	Meta
	list      bool
	write     bool
	diff      bool
	check     bool
	recursive bool
}

func (*FmtCommand) Help() string {
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

func (*FmtCommand) Synopsis() string {
	return "Rewrites HCL2 config files to canonical format"
}

func (*FmtCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictOr(
		complete.PredictDirs("*"),
		complete.PredictFiles("*.nomad"),
		complete.PredictFiles("*.hcl"),
	)
}

func (*FmtCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-list":      complete.PredictNothing,
		"-check":     complete.PredictNothing,
		"-diff":      complete.PredictNothing,
		"-write":     complete.PredictNothing,
		"-recursive": complete.PredictNothing,
	}
}

func (c *FmtCommand) Run(args []string) int {
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

func (c *FmtCommand) fmt(paths []string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, path := range paths {
		// path could be stdin ("-"), or a file, or a directory
		if path == stdinArg {
			c.Ui.Info("processing stdin")
			diags.Extend(c.processFile(path, os.Stdin, os.Stdout))
			// TODO

		} else {
			info, _ := os.Stat(path)
			switch {
			case info.IsDir():
				c.Ui.Info("processing directory")
				// TODO

			default:
				c.Ui.Info("processing physical file")

				f, err := os.Open(path)
				if err != nil {
					diags = diags.Append(&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("Failed to open %s", path),
						Detail:   fmt.Sprintf("%+v", err),
					})
				}

				defer f.Close()
				diags.Extend(c.processFile(path, f, os.Stdout))
			}
		}
	}

	return diags
}

func (c *FmtCommand) processFile(path string, r io.Reader, w io.Writer) hcl.Diagnostics {
	var diags hcl.Diagnostics

	src, err := io.ReadAll(r)
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed to read %s", path),
			Detail:   fmt.Sprintf("%+v", err),
		})
		return diags
	}

	// File must be parseable as HCL native syntax before we'll try to format
	// it. If not, the formatter is likely to make drastic changes that would
	// be hard for the user to undo.
	_, syntaxDiags := hclsyntax.ParseConfig(src, path, hcl.Pos{Line: 1, Column: 1})
	if syntaxDiags.HasErrors() {
		diags = diags.Extend(syntaxDiags)
		return diags
	}

	out := hclwrite.Format(src)
	w.Write(out)

	return diags
}
