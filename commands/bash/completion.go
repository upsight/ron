package bash

import (
	"fmt"
	"io"
)

const (
	ronComplete = `
_ron()
{
    local cur cmds topts copts tmopts sub_cmd
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    sub_cmd="${COMP_WORDS[1]}"
    cmds=$(ron help 2>&1 | grep -E '^    ' | cut -d ' ' -f5)
    targetopts=$(ron t -list_clean)
    topts="--debug --default --envs --list --verbose --yaml"
    copts="--loglevel --restart --wait --watch"
    tmopts="--debug --input --output"

    case ${COMP_CWORD} in
        1)
            COMPREPLY=($(compgen -W "${cmds}" -- ${cur}))
            ;;
        *)
            case ${sub_cmd} in
                t | target)
                    COMPREPLY=($(compgen -W "${topts} ${targetopts}" -- ${cur}))
                    ;;
                cmd)
                    COMPREPLY=($(compgen -W "${copts}" -- ${cur}))
                    ;;
                template)
                    COMPREPLY=($(compgen -W "${tmopts}" -- ${cur}))
                    ;;
            esac
            ;;
    esac
    return 0
}
complete -F _ron ron
`
)

// Command ...
type Command struct {
	Name string
	W    io.Writer
	WErr io.Writer
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	fmt.Fprintf(c.WErr, `Copy the following into your ~/.bashrc file or into /etc/bash_completion/ron`)
	fmt.Fprintf(c.W, ronComplete)
	return 0, nil
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"b":               struct{}{},
		"bash_completion": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Print the bash completion script."
}
