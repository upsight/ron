![ron](ron.jpg)

# Ron [![GoDoc](https://godoc.org/github.com/upsight/ron?status.svg)](http://godoc.org/github.com/upsight/ron) [![Build Status](https://travis-ci.org/upsight/ron.svg?branch=master)](https://travis-ci.org/upsight/ron)

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

http://godoc.org/github.com/upsight/ron

### Installation

	$ go get -u github.com/upsight/ron/cmd/...

or download from [releases](https://github.com/upsight/ron/releases)

### Upgrade

	LATEST_URL=https://github.com/upsight/ron/releases/download/v1.1.0/ron-linux-v1.1.0 ron upgrade

### Testing

	$ ron t go:test
