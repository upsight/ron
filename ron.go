package ron

import "flag"

// Run builds its internal commands and executes a
// matching command by parsing args and flags.
func Run(c *Commander, args []string) (int, error) {
	f := flag.NewFlagSet(AppName, flag.ExitOnError)
	f.Usage = func() { c.Usage(c.Stderr) }
	var list bool
	f.BoolVar(&list, "list", false, "List commands")
	f.Parse(args)

	if list {
		c.List(c.Stdout)
		return 0, nil
	}

	if f.NArg() < 1 {
		c.Usage(c.Stderr)
		return 1, nil
	}

	for _, cmd := range c.Commands {
		if _, ok := cmd.Aliases()[f.Arg(0)]; ok {
			return cmd.Run(f.Args()[1:])
		}
	}
	c.Usage(c.Stderr)
	return 1, nil
}
