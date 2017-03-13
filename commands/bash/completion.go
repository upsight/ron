package bash

import (
	"fmt"
	"io"
)

const (
	ronComplete = `
_ron()
{
    local cur sub_cmd
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    sub_cmd="${COMP_WORDS[1]}"

    case ${COMP_CWORD} in
        1)
            local cmds=$(ron -list)
            COMPREPLY=($(compgen -W "${cmds}" -- ${cur}))
            ;;
        *)
            case ${sub_cmd} in
                cmd)
                    local command_opts="-loglevel -restart -wait -watch"
                    COMPREPLY=($(compgen -W "${replace_opts}" -- ${cur}))
                    ;;
                replace)
                    local replace_opts="-debug -input -output"
                    COMPREPLY=($(compgen -W "${command_opts}" -- ${cur}))
                    ;;
                t | target)
                    local target_opts="-debug -default -envs -list -list_remotes -remotes -verbose -yaml"
                    local target_list_opts=$(ron t -list_clean)
                    COMPREPLY=($(compgen -W "${target_opts} ${target_list_opts}" -- ${cur}))
                    ;;
                template)
                    local template_opts="-debug -input -output"
                    COMPREPLY=($(compgen -W "${template_opts}" -- ${cur}))
                    ;;
            esac
            ;;
    esac
    return 0
}

complete -F _ron ron
# See https://tiswww.case.edu/php/chet/bash/FAQ
# The current set of completion word break characters is available in bash as
# the value of the COMP_WORDBREAKS variable. Removing ':' from that value is
# enough to make the colon not special to completion
COMP_WORDBREAKS=${COMP_WORDBREAKS//:}
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
		"bash_completion": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "Print the bash completion script."
}
