# snap plugin collector - exec

A generic plugin to launch executable files and collects their outputs.

*WARNING*: This plugin gives the power to run any program on the server. Running this plugin with root privileges may be extremely dangerous and it is not advised. 

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Operating systems](#operating-systems)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [snap's Global Config](#snaps-global-config)
  * [Setfile structure](#setfile-structure)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

### System Requirements

* [golang 1.5+](https://golang.org/dl/) - needed only for building

### Operating systems
All OSs currently supported by Snap:
* Linux/amd64

### Installation

#### Download exec plugin binary:

You can get the pre-built binaries for your OS and architecture at Snap's [Github Releases](https://github.com/intelsdi-x/snap/releases) page.

#### To build the plugin binary:

Fork https://github.com/intelsdi-x/snap-plugin-collector-exec

Clone repo into `$GOPATH/src/github/intelsdi-x/`:
```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-exec
```
Build the plugin by running make in repo:
```
$ make
```
This builds the plugin in `/build/rootfs`.

### Configuration and Usage

* Set up the [Snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started).

* Create the configuration file (called `setfile`) using the examples below under [setfile structure](https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/README.md#setfile-structure) or by example in [`examples/setfiles/`](https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/setfiles/).

* Create a [Snap Global Config](https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/README.md#snaps-global-config), which is a requirement for this plugin.

#### Why Snap Global Config is required 
Notice that this plugin is a generic plugin. It cannot work without configuration, because there is no reasonable default behavior. Errors in configuration or in execution of the commands run by this plugin will block collecting of metrics and will show up in plugin log files.

#### A Note on Security
The Setfile contains configuration which should be protected. It is advised to limit access for this configuration and regularly audit its content. 

It is not advised to use dynamic query notation in task manifest for exec plugin. Users should consciously make decision on the executing command to avoid unexpected changes to the system.

#### Other important notes on usage

The executable file which is used to collect metrics must write its output (value of metric) to standard output (stdout) in plain form, without unnecessary white characters.

User interaction cannot be needed by executable file. Some commands or programs require special privileges to execute, so be aware of configuration to ensure successful runs.

The executable file will launch a process for each metric gathered and will be launch at the interval set in the Task Manifest (or through `snaptel`). Processes are expected to end and clean up any used resources between runs. This behavior may have impact on system performance.

## Documentation

### Collected Metrics
The plugin collects the outputs of executable files as metrics in the namespace `/intel/exec/<metric_name>/`.

Each metrics's name is defined in the Setfile. Metrics can be any of the following data types: float64, float32, int64, int32, int16, int8, uint64, uint32, uint16, uint8, string.

### Snap's Global Config
Global configuration files are described in [snap's documentation](https://github.com/intelsdi-x/snap/blob/master/docs/SNAPD_CONFIGURATION.md). A section is required, titled "exec" in "collector", with the following options:
- `"setfile"` - path to exec plugin configuration file (path to Setfile),
- `"execution_timeout"` -   max time for command/program execution in seconds (default value: 10 sec).

See example Global Config in [examples/cfg/] (https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/configs/).


### Setfile structure

Setfile contains a JSON structure which is used to define metrics. Each metric is defined as JSON object in following format:
```
  "<metric_name>": {
            "exec": "<executable_file>",
            "type": "<data_type>",
            "args": [ "<arg1>", "<arg2>", "<arg3>"]
    }
```
Where:
- `metric_name` -  metric name which is used in metric's namespace (required),
- `executable_file` -  path to executable file which should by launch to collect metric (required),
- `data_type` -  metric data type (required)
- `arg1`, `arg2`, `arg3` -  arguments needed by executable file which is used to collect metric (optional).

For example `'echo_metric'` metric for the `'echo'` program is available in `'/bin'` with arguments `'-n'`, `'1.1'` and results in a float64 data type should have the following definition:
```
  "echo_metric": {
            "exec": "/bin/echo",
            "type": "float64",
            "args": ["-n", "1.1"]
    }
```
The metric defined above has the following namespace `/intel/exec/echo_metric`.

If the running process returns metric with additional information, it will require another tool to extract the target values from the output. For example, the below example shows extraction of numeric data from output of `echo`:
```
"echo_metric": {
				"exec": "/bin/sh",
				"type": "int64",
				"args": [ "-c", "echo  \"test:1775\" | awk -F':' '{printf $2}'"]
		}
```
As you can see, `exec` in setfile could be defined as a combination of commands.

### Examples
To walk through a working example of snap-plugin-collector-exec, follow these steps:

1. Create a configuration file (setfile) or copy the example file at [`examples/setfiles/`](https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/setfiles/).

2. Copy the example Setfile and then set the correct path to the configuration file as the field `setfile` along with a max time for the process to execute as the field `execution_timeout` in Global Config ([`examples/configs/`] (https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/configs/)).

3. In one terminal window, start `snapteld`, the Snap daemon, (in this case with logging set to 1,  trust disabled and global configuration saved in config.json ):
```
$ snapteld -l 1 -t 0 --config config.json
```
4. In another terminal window:

Load snap-plugin-collector-exec plugin
```
$ snaptel plugin load snap-plugin-collector-exec
```
Load file plugin for publishing:
```
$ snaptel plugin load snap-plugin-publisher-file
```
See available metrics for your system

```
$ snaptel metric list
```

5. Write a Task Manifest (example in [`examples/tasks/`] (https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/tasks/)):
```
{
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "5s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/exec/metric0": {},
                "/intel/exec/metric1": {},
                "/intel/exec/metric2": {},
                "/intel/exec/metric3": {},
                "/intel/exec/metric4": {}
           },
            "config": {
            },
            "process": null,
            "publish": [
                {
                    "plugin_name": "file",
                    "config": {
                        "file": "/tmp/published_exec"
                    }
                }
            ]
        }
    }
}
```

6. Now create the task from the Task Manifest:
```
$ snapctl task create -t task.json
ID: ef720332-8f0f-4cd7-84f8-73219d403c35
Name: Task-ef720332-8f0f-4cd7-84f8-73219d403c35
State: Running
```

7. And watch the metrics populate: 
```
$ snapctl task watch ef720332-8f0f-4cd7-84f8-73219d403c35
```

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-exec/issues) and feel free to then submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-exec/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, the open telemetry framework. See the full project at http://github.com/intelsdi-x/snap. To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Katarzyna Kujawa](https://github.com/katarzyna-z)
