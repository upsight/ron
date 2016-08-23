package ron

import (
	"flag"
	"io"
	"os"
)

// Run builds its internal commands and executes a
// matching command by parsing args and flags.
func Run(stdOut io.Writer, stdErr io.Writer, args []string) (int, error) {
	if stdOut == nil {
		stdOut = os.Stdout
	}
	if stdErr == nil {
		stdErr = os.Stderr
	}

	c := NewCommander(stdOut, stdErr)

	flagSet := flag.NewFlagSet(AppName, flag.ExitOnError)
	flagSet.Usage = func() { c.Usage(stdErr) }
	flagSet.Parse(args)
	if flagSet.NArg() < 1 {
		c.Usage(stdErr)
		return 1, nil
	}

	for _, cmd := range c {
		if _, ok := cmd.Names()[flagSet.Arg(0)]; ok {
			return cmd.Run(flagSet.Args()[1:])
		}
	}
	c.Usage(stdErr)
	return 1, nil
}
