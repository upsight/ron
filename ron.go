package ron

import "flag"

// Run builds its internal commands and executes a
// matching command by parsing args and flags.
func Run(c *Commander, args []string) (int, error) {
	flagSet := flag.NewFlagSet(AppName, flag.ExitOnError)
	flagSet.Usage = func() { c.Usage(c.Stderr) }
	flagSet.Parse(args)
	if flagSet.NArg() < 1 {
		c.Usage(c.Stderr)
		return 1, nil
	}

	for _, cmd := range c.Commands {
		if _, ok := cmd.Aliases()[flagSet.Arg(0)]; ok {
			return cmd.Run(flagSet.Args()[1:])
		}
	}
	c.Usage(c.Stderr)
	return 1, nil
}
