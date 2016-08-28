![ron](ron.jpg)

# Ron

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

http://godoc.org/github.com/upsight/ron

### Installation

	go get -u github.com/upsight/ron/cmd

or download from [releases](https://github.com/pkar/ron/releases)

### Upgrade

	LATEST_URL=https://github.com/upsight/ron/releases/download/v0.0.1/ron-linux-v0.0.1 ron upgrade 
