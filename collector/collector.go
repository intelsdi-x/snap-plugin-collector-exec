// +build linux

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap-plugin-utilities/config"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/mitchellh/mapstructure"
)

const (
	//vendor namespace part
	vendor = "intel"

	//pluginName namespace part
	pluginName = "exec"

	// version of plugin
	version = 1

	//pluginType type of plugin
	pluginType = plugin.CollectorPluginType

	//nsLength allowed length of namespace
	nsLength = 3

	//setFileConfigVar configuration variable to define path to setfile
	setFileConfigVar = "setfile"

	//execTimeOutConfigVar configuration variable to define max time for command/program execution
	execTimeOutConfigVar = "execution_timeout"

	//metricExecMapKey key in setfile to mark path to executable file
	metricExecMapKey = "exec"

	//metricTypeMapKey key in setfile to mark type of returned value of metric
	metricTypeMapKey = "type"

	//argsMapKey key in setfile to mark arguments needed by executable file
	argsMapKey = "args"
)

//Plugin exec plugin struct which gathers plugin specific data
type Plugin struct {
	host    string
	metrics map[string]metric
	cmd     exeCmd
}

//Meta returns meta data for plugin
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		pluginName,
		version,
		pluginType,
		[]string{},
		[]string{plugin.SnapGOBContentType},
		plugin.ConcurrencyCount(1))
}

// New creates instance of exec collector plugin
func New() *Plugin {
	metrics := make(map[string]metric)
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	return &Plugin{host: host, metrics: metrics, cmd: executeCmd}
}

// GetMetricTypes returns list of available metric types
// It returns error in case retrieval was not successful
func (p *Plugin) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}
	item, err := config.GetConfigItem(cfg, setFileConfigVar)
	if err != nil {
		return mts, err
	}
	setFilePath, ok := item.(string)
	if !ok {
		return mts, serror.New(fmt.Errorf("Incorrect type of configuration variable, cannot parse value of %s to string", setFileConfigVar), nil)
	}

	serr := p.getMetricsFromConfig(setFilePath)
	if serr != nil {
		log.WithFields(serr.Fields()).Error(serr.Error())
		return mts, serr
	}

	for mtsName := range p.metrics {
		mts = append(mts, plugin.MetricType{
			Namespace_: core.NewNamespace(vendor, pluginName, mtsName)})
	}

	return mts, nil
}

// CollectMetrics returns list of requested metric values
// It returns error in case retrieval was not successful
func (p *Plugin) CollectMetrics(metrics []plugin.MetricType) ([]plugin.MetricType, error) {
	mts := []plugin.MetricType{}
	items, err := config.GetConfigItems(metrics[0], setFileConfigVar, execTimeOutConfigVar)

	if err != nil {
		return nil, err
	}
	setFilePath, ok := items[setFileConfigVar].(string)
	if !ok {
		return mts, serror.New(fmt.Errorf("Incorrect type of configuration variable, cannot parse value of %s to string", setFileConfigVar), nil)
	}

	execTimeout, ok := items[execTimeOutConfigVar].(int)
	if !ok {
		return mts, serror.New(fmt.Errorf("Incorrect type of configuration variable, cannot parse value of %s to int", execTimeOutConfigVar), nil)
	}
	execTimeoutSec := time.Second * time.Duration(execTimeout)

	serr := p.getMetricsFromConfig(setFilePath)
	if serr != nil {
		log.WithFields(serr.Fields()).Error(serr.Error())
		return nil, serr
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(metrics))

	for _, m := range metrics {

		go func(m plugin.MetricType) {
			defer wg.Done()
			logFields := map[string]interface{}{}

			ns := m.Namespace()
			logFields["namespace"] = m.Namespace().String()
			if len(ns) != nsLength {
				serr := serror.New(fmt.Errorf("Incorrect namespace length"), logFields)
				log.WithFields(serr.Fields()).Warn(serr.Error())
				return
			}
			//get metric name, it is the last element of namespace
			mtName := ns[nsLength-1].Value

			timer := time.Now()
			//execute command
			cmdOut, serr := p.cmd(p.metrics[mtName].Exec, p.metrics[mtName].Args)
			if serr != nil {
				serr.SetFields(logFields)
				log.WithFields(serr.Fields()).Warn(serr.Error())
				return
			}

			if time.Since(timer) > execTimeoutSec {
				//only notify that execution of command needs more time
				serr := serror.New(fmt.Errorf("Waiting for output more than %d seconds", execTimeoutSec), logFields)
				log.WithFields(serr.Fields()).Warn(serr.Error())
			}

			//convert output of command execution to type defined in setfile
			data, serr := convertMetricType(cmdOut, p.metrics[mtName].Type)
			if serr != nil {
				serr.SetFields(logFields)
				log.WithFields(serr.Fields()).Warn(serr.Error())
				return
			}

			mt := plugin.MetricType{
				Namespace_: m.Namespace(),
				Data_:      data,
				Timestamp_: time.Now(),
			}

			mu.Lock()
			mts = append(mts, mt)
			mu.Unlock()

		}(m)

	}
	wg.Wait()
	return mts, nil
}

// GetConfigPolicy returns config policy
// It returns error in case retrieval was not successful
func (p *Plugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule(setFileConfigVar, true)
	if err != nil {
		return cp, err
	}
	r1.Description = "Configuration file"
	config.Add(r1)

	r2, err := cpolicy.NewIntegerRule(execTimeOutConfigVar, false, 10)
	if err != nil {
		return cp, err
	}
	r2.Description = "Execution timeout"
	config.Add(r2)

	return cp, nil
}

// getMetricsFromConfig extracts metrics configuration from setfile
func (p *Plugin) getMetricsFromConfig(setFilePath string) serror.SnapError {
	logFields := map[string]interface{}{}
	logFields["setFilePath"] = setFilePath

	setFileContent, err := ioutil.ReadFile(setFilePath)
	logFields["setFileContent"] = setFileContent
	if err != nil {
		return serror.New(err, logFields)
	}

	if len(setFileContent) == 0 {
		return serror.New(fmt.Errorf("Settings file is empty"), logFields)
	}

	var setFileUnmarshalled map[string]interface{}
	err = json.Unmarshal(setFileContent, &setFileUnmarshalled)
	if err != nil {
		return serror.New(fmt.Errorf("Settings file cannot be unmarshalled"), logFields)
	}

	err = mapstructure.Decode(setFileUnmarshalled, &p.metrics)
	if err != nil {
		return serror.New(fmt.Errorf("Settings file cannot be decoded"), logFields)
	}

	//validate if structure contains necessary fields
	for k, m := range p.metrics {
		if m.Type == "" {
			return serror.New(fmt.Errorf("Incorrect structure of settings file, missing metric type for %s .", k), logFields)
		}
		if m.Exec == "" {
			return serror.New(fmt.Errorf("Incorrect structure of settings file, missing metric exec for %s .", k), logFields)
		}
	}

	return nil
}

//convertMetricType converts metric value to type defined in setfile
func convertMetricType(data []byte, dataType string) (interface{}, serror.SnapError) {
	var err error
	var converted interface{}

	dataStr := string(data)
	logFields := map[string]interface{}{"date": data, "dataType": dataType}

	switch dataType {
	case "float64":
		converted, err = strconv.ParseFloat(dataStr, 64)
	case "float32":
		converted, err = strconv.ParseFloat(dataStr, 32)
	case "int64":
		converted, err = strconv.ParseInt(dataStr, 10, 64)
	case "int32":
		converted, err = strconv.ParseInt(dataStr, 10, 32)
	case "int16":
		converted, err = strconv.ParseInt(dataStr, 10, 16)
	case "int8":
		converted, err = strconv.ParseInt(dataStr, 10, 8)
	case "uint64":
		converted, err = strconv.ParseUint(dataStr, 10, 64)
	case "uint32":
		converted, err = strconv.ParseInt(dataStr, 10, 32)
	case "uint16":
		converted, err = strconv.ParseInt(dataStr, 10, 16)
	case "uint8":
		converted, err = strconv.ParseInt(dataStr, 10, 16)
	case "string":
		converted = dataStr
	default:
		converted = dataStr
		serr := serror.New(fmt.Errorf("Unsupported data type, metric saved as string"), logFields)
		log.WithFields(serr.Fields()).Warn(serr.Error())
	}

	if err != nil {
		return converted, serror.New(err, logFields)
	}
	return converted, nil
}

type exeCmd func(executableFilePath string, args []string) ([]byte, serror.SnapError)

func executeCmd(executableFilePath string, args []string) ([]byte, serror.SnapError) {
	cmdOut, err := exec.Command(executableFilePath, args...).Output()
	if err != nil {
		return cmdOut, serror.New(err)
	}
	return cmdOut, nil
}

type metric struct {
	Exec string
	Type string
	Args []string
}
