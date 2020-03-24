## Golang External Agent Helper

This is a library that can help you write a [Choria External Agent](https://choria.io/docs/development/mcorpc/externalagents/) in Go.

External Agents are agents that live outside of the main Choria Server binary and so does not require recompiling the server to include it.

They are distributable over the Puppet Forge or any other method that can put files on disk.

## Status

Today this library can be used to build standalone agents that you compile and deliver to your nodes, in the next version of Choria you will be able to deliver just the source for agents that rely only on Go stdlib and they will be compiled and run in place.

It's a new library and a new feature in Choria, feedback and bug fixes appreciated.

This is heavily inspired by the Python [py-mco-agent](https://github.com/optiz0r/py-mco-agent) written by Ben Roberts.

## Example

We will write a basic agent called `parrot` that receives a message on its `echo` action and sends it back.

This agent can be invoked from the CLI as `choria req parrot echo message="hello world!"` and likewise be callable from any other API client like playbooks, Ruby or Go.

## Code

The most basic agent that receives a request, parses it and sends back a reply can be seen here.

```golang
package main

import (
	"github.com/choria-io/go-external/agent"
)

type echoRequest struct {
	Message string `json:"message"`
}

func echoAction(request *agent.Request, reply *agent.Reply, config map[string]string) {
	req := &echoRequest{}

	// parse the received request, sets appropriate errors on reply on failure
	if !request.ParseRequestData(req, reply) {
		return
	}

	reply.Data = map[string]string{
		"message": req.Message,
	}
}

func main() {
	parrot := agent.NewAgent("parrot")
	defer parrot.ProcessRequest()

	// action will be invoked on demand
	parrot.MustRegisterAction("echo", echoAction)
}
```

### Activation

In some cases your agent might have dependencies that the node need to satisfy before it can activate. Without an activator - like the above code - the agent will be active on any node.

```golang
// checks if a specific dependency exist, here we just check some file is on the node
func shouldActivate(agent string, config map[string]string) (bool, error) {
    _, err := os.Stat("/etc/dependency.txt")
    if os.IsNotExist(err) {
        // logs as info level in the choria server log
        agent.Infof("The /etc/dependency.txt file could not be found")
        return false, nil
    } else {
        // logs at error level in the choria server log
        agent.Errorf("Could not check if /etc/dependency.txt exist: %s", err)
        return false, err
    }

    return true, nil
}

func main() {
	parrot := external.NewAgent("parrot")
	defer parrot.ProcessRequest()

	// shouldActivate will be called on agent startup
	parrot.RegisterActivator(shouldActivate)
	parrot.MustRegisterAction("echo", echoAction)
}
```

### Configuration

The action and activator both receive a config map, this is a parsed version of the contents of - for example - `/etc/choria/plugin.d/parrot.cfg`. 

It's a simple file in the format:

```
setting = value
```

### Logging

The above example shows to logging examples, external agents can only log at level `info` and `error`. Any `STDOUT` output would be `info` level and `STDERR` output is logged as error.

### DDL

Choria needs 2 files that describe the features and behavior of the agent.  These are called DDL files and can be generated using `choria tool generate ddl parrot.json parrot.ddl`.  This wizard will guide you though creating these files.

You should fill in details as the wizard suggests, add 1 action - `echo` - and a `string` input and output called `message`.

### Facts

At the time of invoking your action the server will write a JSON file holding a snapshot of it's facts at the time. You can access this using `external.Facts()` or a path to the file in `external.FactsPath()`. This requires Choria Server version 0.14.0 or newer.

Handling Random JSON data in Go is a bit complicated, I suggest a look at [gjson](https://github.com/tidwall/gjson) to dig into the data.

## Packaging

Plugins can be [packaged and distributed on the forge](https://choria.io/docs/development/mcorpc/packaging/). Create a directory layout as below:

```
parrot
└── agent
    ├── parrot
    ├── parrot.ddl
    └── parrot.json
```

Here the `parrot` file is the compiled version of the binary produced above. In a future release of Choria Server we will support storing Go Code in the file without compiling it.

Execute the command `mco plugin package --vendor yourco` in the `parrot` directory and it will create a plugin in a file similar to `yourco-mcollective_agent_parrot-1.0.0.tar.gz`
