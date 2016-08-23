/*
Package ron provides a command line interface to common build tasks.
ron is a process runner/container builder/yaml unicorn maker/zombie process reaper.

	$ ron
	Usage: ron <command>

	Available commands are:
			b, bash_completion    Print the bash completion script.
			burgundy              Stay classy.
			cmd                   Run a command with optional restart and watch for changes to restart.
			replace               Find and replace in text.
			t, target             Execute a configured target.
			template              Render a Go template using environment variables.
			upgrade               Upgrade the ron binary.
			version               Print the version.

Documentation

You can find the documentation here http://godoc.org/github.com/upsight/ron.

To install the latest ron run go get

	$ go get -u github.com/upsight/ron

cmd

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


target

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
	If no -default or -yaml is provided and in the current working directory there
	exists a ron.yaml, then those will be used as the -yaml option.

	The yaml config should contain a list of "envs" and a
	hash of "targets".

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

template

	$ ron template
	Usage: ron template
		-input string
				Path or URL to template file.
		-output string
				Path to output file. Defaults to stdout.

upgrade

	$ ron upgrade

version

		$ ron version
		ron 0.0.1 0.0.1

Testing

To start install the latest ron then

There are a few ways to test

	$ ron target test

Generate and show coverage

	$ ron target cover
*/
package ron
