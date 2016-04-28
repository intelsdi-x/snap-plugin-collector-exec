# snap plugin collector - exec

The plugin launches executable files and collects their outputs.

The plugin is a generic plugin. You can configure metrics names, type of collecting data, executables files, arguments necessary for executables files.

This plugin gives powerful and dangerous tool - user can run any program. Running this plugin with root privileges may be extremely dangerous and it is not advised. 

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
All OSs currently supported by snap:
* Linux/amd64

### Installation

#### Download exec plugin binary:

You can get the pre-built binaries for your OS and architecture at snap's [Github Releases](https://github.com/intelsdi-x/snap/releases) page.

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

* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started).

* Create configuration file (called as a setfile) in which will be defined metrics (metrics names, type of collecting data, executables files, arguments necessary for executables files), read setfile structure description available in [setfile structure](https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/README.md#setfile-structure) and see exemplary in [examples/setfiles/](https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/setfiles/).

* Create Global Config, see description in [snap's Global Config] (https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/README.md#snaps-global-config).
 
Notice that this plugin is a generic plugin, it cannot work without configuration, because there is no reasonable default behavior. Errors in configuration or in execution of program/command block collecting of metrics, errors are logged but plugin still work.

Setfile contains configuration which should be protected. It is advised to limit access for this configuration and regularly audit its content.

It is not advised to use dynamic query notation in task manifest for exec plugin. User should consciously make decision on starting program or executing command to avoid unexpected changes.

The executable file which is used to collect metrics must write its output (value of metric) to standard output (stdout) in plain form. User interaction cannot be needed by executable file. Some of commands or programs require special privileges to execute so user must check if configuration can be launch successfully.

The executable file will be launch per interval set in task manifest and it may have impact on system performance. New process is started per interval for each of metric and this process must end and clean resources itself.

## Documentation

### Collected Metrics
The plugin collects outputs of executable files.

Metrics are available in namespace: `/intel/exec/<metric_name>/`.

Metrics's names are defined in setfile.
Metrics can be collected in one of following data types: float64, float32, int64, int32, int16, int8, uint64, uint32, uint16, uint8, string.

### snap's Global Config
Global configuration files are described in [snap's documentation](https://github.com/intelsdi-x/snap/blob/master/docs/SNAPD_CONFIGURATION.md). You have to add section "exec" in "collector" section and then specify following options:
- `"setfile"` - path to exec plugin configuration file (path to setfile),
- `"execution_timeout"` -   max time for command/program execution in seconds (default value: 10 sec).

See example Global Config in [examples/cfg/] (https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/configs/).


### Setfile structure

Setfile contains JSON structure which is used to define metrics. 
Metric is defined as JSON object in following format:
```
  "<metric_name>": {
            "exec": "<executable_file>",
            "type": "<data_type>",
            "args": [ "<arg1>", "<arg1>", "<arg3>"]
    }
```
where:
- metric_name -  metric name which is used in metric's namespace (required),
- executable_file -  path to executable file which should by launch to collect metric (required),
- data_type -  metric data type (required)
- arg1, arg1, arg3 -  arguments needed by executable file which is used to collect metric (optional).

For example 'echo_metric' metric for  'echo' program which is available in '/bin' with argument '1.1' and  float64 data type should have following definition:
```
  "echo_metric": {
            "exec": "/bin/echo",
            "type": "float64",
            "args": [ "1.1"]
    }
```
The metric defined above has following namespace `/intel/exec/echo_metric`.

If program returns metric with additional information then it is needed to use another tool to extract value of metric, see example below which shows extraction of numeric data from output of echo command:
```
"echo_metric": {
				"exec": "/bin/sh",
				"type": "int64",
				"args": [ "-c", "echo  \"test:1775\" | awk -F':' '{printf $2}'"]
		}
```
As it is shown, `exec` in setfile could be defined as combination of commands.

### Examples
Example running snap-plugin-collector-exec plugin and writing data to a file.

Create configuration file (setfile) for exec plugin, see exemplary in [examples/setfiles/](https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/setfiles/).

Set path to configuration file as the field `setfile` and maximal time for command/program execution as the field `execution_timeout` in Global Config, see exemplary in [examples/configs/] (https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/configs/).

In one terminal window, open the snap daemon (in this case with logging set to 1,  trust disabled and global configuration saved in config.json ):
```
$ $SNAP_PATH/bin/snapd -l 1 -t 0 --config config.json
```
In another terminal window:

Load snap-plugin-collector-exec plugin
```
$ $SNAP_PATH/bin/snapctl plugin load snap-plugin-collector-exec
```
Load file plugin for publishing:
```
$ $SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-publisher-file
```
See available metrics for your system

```
$ $SNAP_PATH/bin/snapctl metric list
```

Create a task manifest file to use snap-plugin-collector-exec plugin (exemplary in [examples/tasks/] (https://github.com/intelsdi-x/snap-plugin-collector-exec/blob/master/examples/tasks/)):
```
{
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
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

Create a task:
```
$ $SNAP_PATH/bin/snapctl task create -t task.json
```

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-users/issues) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-users/pulls).

## Community Support
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap.

To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support) or visit [snap Gitter channel](https://gitter.im/intelsdi-x/snap).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Katarzyna Zabrocka](https://github.com/katarzyna-z)
