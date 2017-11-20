/*
Package ron provides a command line interface to common build tasks.

	$ ron
	Usage: ron <command>

	Available commands are:
	    b, bash_completion    Print the bash completion script.
	    burgundy              Stay classy.
	    cmd                   Run a command with optional restart and watch for changes to restart.
	    hs, httpstat          HTTP trace timings
	    replace               Find and replace in text.
	    t, target             Execute a configured target.
	    template              Render a Go template using environment variables.
	    upgrade               Upgrade the ron binary.
	    version               Print the version.

Installation

ron can be used as a standalone binary, or imported as a library.

To install the latest ron use go get

	$ go get -u github.com/upsight/ron/cmd/ron


Bash Completion

The bash completion command (b,bash_completion) can be added to bashrc or bash_profile by
appending the output or manually by copying it's output. Tab completion is intended for use
the target command to get a list of targets.

	$ ron bash_completion >> ~/.bashrc
	$ source ~/.bashrc

	$ ron t
	--debug             --list              default:_hello      ron:build           ron:install         ron:run             ron:vet
	--default           --verbose           default:burgundy    ron:build_all       ron:lint            ron:test
	--envs              --yaml              ron:_vendor_update  ron:cover           ron:prep            ron:testv

	$ ron t ron
	ron:_vendor_update  ron:build_all       ron:install         ron:prep            ron:test            ron:vet
	ron:build           ron:cover           ron:lint            ron:run             ron:testv

	$ ron t defa
	default:_hello    default:burgundy


Http Stat

The http stat (hs,httpstat) command will give response times broken up by task.

	$ ron httpstat http://google.com
	Connected to 74.125.199.106:80

	HTTP/1.1 200 OK
	Server: gws
	Cache-Control: private, max-age=0
	Content-Type: text/html; charset=ISO-8859-1
	Date: Tue, 28 Feb 2017 20:27:43 GMT
	Expires: -1
	P3p: CP="This is not a P3P policy! See https://www.google.com/support/accounts/answer/151657?hl=en for more info."
	Set-Cookie: NID=98=TKCxQ2pGglNzt6RfHysWo6m-KmiUtHV0UWIuIJa2SLqnLIx2G9mRwuDTntyLvOIZD6bDVVw_jnbHuqUwiiAUqE_xIaiyNpcBtIjpkoCRkuGQu1Pb2Y1rBaLlVBLWrj008GArnUIe5lmshObT; expires=Wed, 30-Aug-2017 20:27:43 GMT; path=/; domain=.google.com; HttpOnly
	X-Frame-Options: SAMEORIGIN
	X-Xss-Protection: 1; mode=block

	Body discarded

	   DNS Lookup   TCP Connection   Server Processing   Content Transfer
	[      14ms  |          19ms  |             56ms  |             0ms  ]
	             |                |                   |                  |
	    namelookup:14       ms      |                   |                  |
	                        connect:33       ms         |                  |
	                                      starttransfer:89       ms        |
	                                                                 total:90       ms
Cmd

The cmd (cmd) command allows for watching file changes and restarting or executing commands.

	$ ron cmd
	Usage: ron cmd --wait --watch <path> --restart <command>
		-ignore string
				a comma seperated list of regex patterns to ignore *optional (default ".*\\.git.*,.*\\.DS_Store$,.*\\.pyc$")
		-restart
				Restart the command if it dies.
		-wait
				With watch wait for file changes before running the command.
		-watch string
				Path to directory or file to watch.

	$ ron cmd --restart ls
	2015/08/03 17:02:38 runit.go:69: running ls
	2015/08/03 17:02:38 runit.go:104: running ls
	Dockerfile	README.md	_vendor		bin		config		make.sh		src
	2015/08/03 17:02:38 cmd.go:40: captured child exited continue...
	2015/08/03 17:02:39 runit.go:69: running ls
	2015/08/03 17:02:39 runit.go:104: running ls
	Dockerfile	README.md	_vendor		bin		config		make.sh		src
	2015/08/03 17:02:39 cmd.go:40: captured child exited continue...
	^C2015/08/03 17:02:40 cmd.go:37: captured interrupt

	$ ron cmd --wait --watch . ls
	2015/08/03 17:03:06 runit.go:104: running ls
	Dockerfile	README.md	_vendor		bin		config		make.sh		src
	2015/08/03 17:03:06 cmd.go:40: captured child exited continue...
	2015/08/03 17:03:11 watch.go:28: event:  "foo": CREATE
	2015/08/03 17:03:11 watch.go:47: Detected new file foo
	2015/08/03 17:03:11 runit.go:87: restart event
	2015/08/03 17:03:11 runit.go:141: restarting
	2015/08/03 17:03:11 runit.go:123: killing subprocess
	2015/08/03 17:03:11 runit.go:104: running ls
	Dockerfile	README.md	_vendor		bin		config		foo		make.sh		src
	2015/08/03 17:03:11 cmd.go:40: captured child exited continue...
	2015/08/03 17:03:11 watch.go:28: event:  "foo": CHMOD
	2015/08/03 17:03:11 watch.go:28: event:  "foo": CHMOD

Replace

The replace (replace) command will replace text in a file or directories. If a file is input
only that file will replace text, if a directory is given it will recurse all files
and do the replace.

	$ ron replace
	Usage: ron replace [-debug] [path] [replace] [replacewith]
	  -debug
	    	Debug the replace run

Target

The target (t,target) command allows for specifying bash scripts within yaml files for execution.
Executed targets can be specified by giving the filename without extension, and the
target name seperated by colon.

	$ ron target
	Usage: ron target <target> <target>
	  -debug
	    	Debug the target command being run
	  -default string
	    	Path to the default yaml config file, local or http.
	  -envs
	    	List the initialized environment variables.
	  -l	List the available targets.
	  -list
	    	List the available targets.
	  -v	Be verbose.
	  -verbose
	    	Be verbose.
	  -yaml string
	    	Path to override yaml file, can be local or http.

		ron contains a default set of envs and targets that can be inspected with the
		flag options listed above. Those can also be overidden with another yaml file.
		If no -default or -yaml is provided and in the current or parent working directory there
		exists a ron.yaml, then those will be used as the -yaml option.

		The yaml config should contain "remotes" (optional), "envs", and a hash of "targets".

		remotes should be defined as a map with any environment name and a list of server values. It's only
		necessary to define them once so they could be globally set for example in ~/.ron/remotes.yaml
		You can then reference it with -remote=remotes:some_other_env

			remotes:
				staging:
					-
						host: example1.com
						port: 22
						user: test
					-
						host: example2.com
						port: 22
						user: test
				some_other_env:
					-
						host: exampleprod.com
						port: 22
						user: test
						proxy_host: bastionserver.com
						proxy_port: 22
						proxy_user: bastion_user
						identity_file: /optional/path/to/identityfile

		If no identity file is provided, the users local ssh agent will be attempted. You can add
		keys with ssh-add.

		env values prefixed with a +(subject to change) will be executed and set to the os environment
		prior to target execution.

			envs:
				- APP: ron
				- UNAME: +uname | tr '[:upper:]' '[:lower:]'

		targets can contain a before/after hash which is a list of other targets to
		execute. Each target should contain a cmd which can contain any valid bash
		scripting and can use previously defined envs

			targets:
				prep:
					cmd: |
						echo prep
				install:
					before:
						- prep
					after:
						- prep
					cmd: |
						echo $APP

In order to execute a target you can either run it with the yaml file prefix without extension
or if you leave that off it will find the first available target, with the default targets executing
last.

	$ ron target default:prep
	$ ron target ron:prep
	$ ron target prep

ron will attempt to find any additional files in the following order:

	1. A -yaml flag filepath config file
	2. A local ./ron.yaml or in a parent directory
	3. Any files in the folder ./.ron/*.yaml
	4. Any ~/.ron/*yaml files
	5. A -default configuration or any binary built in targets

Template

The template (template) will render a Go template file to the given output file.

	$ ron template
	Usage: ron template
		-input string
				Path or URL to template file.
		-output string
				Path to output file. Defaults to stdout.

Upgrade

Ron can be upgraded if you already have it installed. The easiest way
is to just provide a LATEST_URL to the upgrade command:

	$ LATEST_URL=https://github.com/upsight/ron/releases/download/v1.1.3/ron-linux-v1.1.3 ron upgrade

Version

To print the current tag and git commit run:

	$ ron version
	ron 0.0.1 53a7de4612c36b4cf36a9059b5dfa66fbc2639f9

*/
package ron
