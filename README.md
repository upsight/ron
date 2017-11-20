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

	LATEST_URL=https://github.com/upsight/ron/releases/download/v1.1.3/ron-linux-v1.1.3 ron upgrade

### Testing

	$ ron t go:test

### Extending

To extend or customize ron, you can either fork the repo, or vendor it and overwrite the target/default.yaml
file and add custom commands.

The vendoring approach may look something like this:

1. Add vendor/github.com/upsight/ron to your project with whatever package manager you choose (dep, glide, etc.)

2. Create a new folder for your custom commands ./commands/commandname/commandname.go

```go
package commandname

import (
	"fmt"
	"io"
)

// Command ...
type Command struct {
	Name       string
	W          io.Writer
	WErr       io.Writer
	AppName    string
	AppVersion string
	GitCommit  string
}

// Run ...
func (c *Command) Run(args []string) (int, error) {
	fmt.Fprintln(c.W, c.AppName, c.AppVersion, c.GitCommit)
	return 0, nil
}

// Key returns the commands name for sorting.
func (c *Command) Key() string {
	return c.Name
}

// Aliases are the aliases and name for the command. For instance
// a command can have a long form and short form.
func (c *Command) Aliases() map[string]struct{} {
	return map[string]struct{}{
		"someothername": struct{}{},
	}
}

// Description is what is printed in Usage.
func (c *Command) Description() string {
	return "My command."
}
```

3. Create a custom cmd/ron/main.go and add all of your commands.

```golang
package main

import (
	"log"
	"os"

	"mygitrepo.com/ron/commands/commandname"

	"github.com/upsight/ron"
	"github.com/upsight/ron/color"
)

func main() {
	c := ron.NewDefaultCommander(os.Stdout, os.Stderr)
	c.Add(&commandname.Command{AppName: ron.AppName, Name: "project", W: os.Stdout, WErr: os.Stderr})
	status, err := ron.Run(c, os.Args[1:])
	if err != nil {
		hostname, _ := os.Hostname()
		log.Println(hostname, color.Red(err.Error()))
	}
	os.Exit(status)
}
```

4. Optionally overwrite ./vendor/github.com/upsight/ron/target/default.yaml to have custom targets.

This can be done in a target or manually.

```bash
touch default.yaml
cp default.yaml vendor/$RONREPO/target/default.yaml
```

5. Create a ./ron.yaml file with instructions on how to put it all together.

```yaml
{{$path := "export PATH=$GOPATH/bin:$PATH"}}
envs:
  - APP: ron
  - ARCH: amd64
  - PACKAGE_VERSION: +echo `git describe --tags`.`git rev-parse HEAD`
  - REPO: myrepo.com/ron
  - RONREPO: github.com/upsight/ron
  - UNAME: +uname | tr '[:upper:]' '[:lower:]'
  - VERSION: 1.4.0
  - TAG: v${VERSION}
targets:
  prep:
    description: Compile the default yaml asset to a go file.
    cmd: |
      go install ./vendor/github.com/jteeuwen/go-bindata/go-bindata || go get -u github.com/jteeuwen/go-bindata/...
      cp default.yaml vendor/$RONREPO/target/default.yaml
      cd vendor/$RONREPO/target
      go-bindata -o default.go -pkg=target default.yaml
  build:
    description: Compile a binary to ./bin/${UNAME}_${ARCH}
    before:
      - go:prep
    cmd: |
      {{$path}}
      mkdir -p bin/${UNAME}_${ARCH}
      GOARCH=$ARCH GOOS=$UNAME go build -o bin/${UNAME}_${ARCH}/${APP}-${UNAME}-${TAG} -ldflags "-X $REPO/vendor/$RONREPO.GitCommit=$PACKAGE_VERSION -X $REPO/vendor/$RONREPO.AppVersion=$TAG -X $REPO/vendor/$RONREPO.AppName=$APP" cmd/ron/*.go
```

6. Run the build

```bash
ron t ron:build && ls bin/
```
