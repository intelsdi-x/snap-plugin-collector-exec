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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMeta(t *testing.T) {
	Convey("Calling Meta function", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, pluginName)
		So(meta.Version, ShouldResemble, version)
		So(meta.Type, ShouldResemble, plugin.CollectorPluginType)
	})
}

func TestNew(t *testing.T) {
	Convey("Creating new plugin", t, func() {
		plugin := New()
		So(plugin, ShouldNotBeNil)
		So(plugin.host, ShouldNotBeNil)
		So(plugin.metrics, ShouldNotBeNil)
	})
}

func TestGetConfigPolicy(t *testing.T) {
	plugin := New()

	Convey("Getting config policy", t, func() {
		So(func() { plugin.GetConfigPolicy() }, ShouldNotPanic)
		configPolicy, err := plugin.GetConfigPolicy()
		So(err, ShouldBeNil)
		So(configPolicy, ShouldNotBeNil)
	})
}

func TestGetMetricTypes(t *testing.T) {
	Convey("Getting exposed metric types", t, func() {

		Convey("when no configuration item available", func() {
			cfg := plugin.NewPluginConfigType()
			plugin := New()
			So(func() { plugin.GetMetricTypes(cfg) }, ShouldNotPanic)
			mts, err := plugin.GetMetricTypes(cfg)
			So(err, ShouldNotBeNil)
			So(mts, ShouldBeEmpty)
		})

		Convey("when path to setfile is incorrect", func() {
			plg := New()
			// file has not existed yet
			deleteMockFile()

			//create configuration
			config := plugin.NewPluginConfigType()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})

			So(func() { plg.GetMetricTypes(config) }, ShouldNotPanic)
			mts, err := plg.GetMetricTypes(config)

			So(err, ShouldNotBeNil)
			So(mts, ShouldBeEmpty)
		})

		Convey("when setfile is empty", func() {

			plg := New()
			createMockFile(mockFileContEmpty)
			defer deleteMockFile()

			config := plugin.NewPluginConfigType()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})

			So(func() { plg.GetMetricTypes(config) }, ShouldNotPanic)
			mts, err := plg.GetMetricTypes(config)
			So(err, ShouldNotBeNil)
			So(mts, ShouldBeEmpty)
		})

		Convey("successfully obtain metrics name", func() {
			plg := New()
			createMockFile(mockFileCont)
			defer deleteMockFile()

			config := plugin.NewPluginConfigType()
			config.AddItem("setfile", ctypes.ConfigValueStr{Value: mockFilePath})

			So(func() { plg.GetMetricTypes(config) }, ShouldNotPanic)
			mts, err := plg.GetMetricTypes(config)
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeEmpty)

			Convey("and proper metric types are returned", func() {
				So(len(mts), ShouldEqual, 5)
			})
		})
	})
}

func TestCollectMetrics(t *testing.T) {
	Convey("Collecting metrics", t, func() {

		Convey("when no configuration settings available", func() {
			// set metrics config
			config := cdata.NewNode()
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			results, err := plg.CollectMetrics(mts)
			So(err, ShouldNotBeNil)
			So(results, ShouldBeEmpty)
		})

		Convey("when execution timout configuration variable is not available", func() {
			// set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			results, err := plg.CollectMetrics(mts)
			So(err, ShouldNotBeNil)
			So(results, ShouldBeEmpty)
		})

		Convey("when execution timout configuration variable has incorrect type", func() {
			// set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueStr{Value: "1"})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			results, err := plg.CollectMetrics(mts)
			So(err, ShouldNotBeNil)
			So(results, ShouldBeEmpty)
		})

		Convey("when setfile configuration variable has incorrect type", func() {
			// set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueInt{Value: 1})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueInt{Value: 1})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			results, err := plg.CollectMetrics(mts)
			So(err, ShouldNotBeNil)
			So(results, ShouldBeEmpty)
		})

		Convey("when configuration is invalid", func() {
			//set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueInt{Value: 1})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			Convey("incorrect path to setfile", func() {
				// setfile does not  exist
				deleteMockFile()

				plg := New()
				So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
				results, err := plg.CollectMetrics(mts)
				So(err, ShouldNotBeNil)
				So(results, ShouldBeEmpty)
			})

			Convey("setfile is empty", func() {
				//setfile is empty
				createMockFile(mockFileContEmpty)
				defer deleteMockFile()

				plg := New()
				So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
				results, err := plg.CollectMetrics(mts)
				So(err, ShouldNotBeNil)
				So(results, ShouldBeEmpty)
			})
		})

		Convey("when command execution ends with error", func() {
			//create setfile
			createMockFile(mockFileCont)
			defer deleteMockFile()

			//set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueInt{Value: 1})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			plg.cmd = mockExecuteCmdErr
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			_, err := plg.CollectMetrics(mts)
			So(err, ShouldBeNil)
		})

		Convey("when type conversion is impossible", func() {
			//create setfile
			createMockFile(mockFileCont)
			defer deleteMockFile()

			//set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueInt{Value: 1})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			plg.cmd = mockExecuteCmdTypeErr
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			_, err := plg.CollectMetrics(mts)
			So(err, ShouldBeNil)
		})

		Convey("when type exec fails", func() {
			//create setfile
			createMockFile(mockFileContExecError)
			defer deleteMockFile()

			//set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueInt{Value: 1})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			plg.cmd = mockExecuteCmdTypeErr
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			_, err := plg.CollectMetrics(mts)
			So(err, ShouldBeNil)
		})

		Convey("collect metrics with timeout error", func() {
			//create setfile
			createMockFile(mockFileCont)
			defer deleteMockFile()

			//set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueInt{Value: 1})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			plg.cmd = mockExecuteCmdWithTimeout
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			_, err := plg.CollectMetrics(mts)
			So(err, ShouldBeNil)
		})

		Convey("collect metrics successfully", func() {
			//create setfile
			createMockFile(mockFileCont)
			defer deleteMockFile()

			//set metrics config
			config := cdata.NewNode()
			config.AddItem(setFileConfigVar, ctypes.ConfigValueStr{Value: mockFilePath})
			config.AddItem(execTimeOutConfigVar, ctypes.ConfigValueInt{Value: 1})
			mts := mockMts
			for i := range mts {
				mts[i].Config_ = config
			}

			plg := New()
			plg.cmd = mockExecuteCmd
			So(func() { plg.CollectMetrics(mts) }, ShouldNotPanic)
			results, err := plg.CollectMetrics(mts)
			So(err, ShouldBeNil)
			So(results, ShouldNotBeEmpty)

			Convey("Then proper metrics values are returned", func() {
				So(len(results), ShouldEqual, len(mts))
				for _, mt := range results {
					So(mt.Data_, ShouldNotBeNil)
				}
			})
		})
	})
}

func TestGetMetricsFromConfig(t *testing.T) {
	Convey("Calling getMetricsFromConfig function", t, func() {

		Convey("Calling getMetricsFromConfig without setfile configuration variable", func() {
			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldNotBeNil)
		})

		Convey("Calling getMetricsFromConfig with incorrect path to setfile", func() {
			deleteMockFile()

			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldNotBeNil)
		})

		Convey("Calling getMetricsFromConfig with empty setfile", func() {
			createMockFile(mockFileContEmpty)
			defer deleteMockFile()

			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldNotBeNil)
		})

		Convey("Calling getMetricsFromConfig with setfile which cannot be unmarshalled", func() {
			createMockFile(mockFileContStructErr)
			defer deleteMockFile()

			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldNotBeNil)
		})

		Convey("Calling getMetricsFromConfig with invalid structure of setfile", func() {
			createMockFile(mockFileContMapErr)
			defer deleteMockFile()

			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldNotBeNil)
		})

		Convey("Calling getMetricsFromConfig with setfile with missing type field", func() {
			createMockFile(mockFileContMissingType)
			defer deleteMockFile()

			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldNotBeNil)
		})

		Convey("Calling getMetricsFromConfig with setfile with missing exec field", func() {
			createMockFile(mockFileContMissingExec)
			defer deleteMockFile()

			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldNotBeNil)
		})

		Convey("Calling getMetricsFromConfig with correct setfile", func() {
			createMockFile(mockFileCont)
			defer deleteMockFile()

			plg := New()
			serr := plg.getMetricsFromConfig(mockFilePath)
			So(serr, ShouldBeNil)
		})

	})
}

func TestConvertMetricType(t *testing.T) {
	Convey("Calling convertMetricType function with different arguments", t, func() {

		_, err := convertMetricType([]byte("test"), "float64")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45.4"), "float64")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("test"), "float32")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45.4"), "float32")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "int64")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "int64")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "int32")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "int32")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "int16")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "int16")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "int8")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "int8")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "uint64")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "uint64")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "uint32")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "uint32")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "uint16")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "uint16")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45.4"), "uint8")
		So(err, ShouldNotBeNil)

		_, err = convertMetricType([]byte("45"), "uint8")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45"), "string")
		So(err, ShouldBeNil)

		_, err = convertMetricType([]byte("45"), "unsuported_type")
		So(err, ShouldBeNil)
	})
}

func createMockFile(fileCont []byte) {
	deleteMockFile()

	f, _ := os.Create(mockFilePath)
	f.Write(fileCont)
}

func deleteMockFile() {
	os.Remove(mockFilePath)
}

func mockExecuteCmd(executableFilePath string, args []string) ([]byte, serror.SnapError) {
	return []byte("65"), nil
}

func mockExecuteCmdErr(executableFilePath string, args []string) ([]byte, serror.SnapError) {
	return []byte("65"), serror.New(fmt.Errorf("Error"))
}

func mockExecuteCmdTypeErr(executableFilePath string, args []string) ([]byte, serror.SnapError) {
	return []byte("test"), nil
}

func mockExecuteCmdWithTimeout(executableFilePath string, args []string) ([]byte, serror.SnapError) {
	time.Sleep(2 * time.Second)
	return []byte("test"), nil
}

var (
	mockMts = []plugin.PluginMetricType{
		plugin.PluginMetricType{Namespace_: []string{vendor, pluginName, "metric4"}},
		plugin.PluginMetricType{Namespace_: []string{vendor, pluginName, "metric3"}},
		plugin.PluginMetricType{Namespace_: []string{vendor, pluginName, "metric2"}},
		plugin.PluginMetricType{Namespace_: []string{vendor, pluginName, "metric1"}},
		plugin.PluginMetricType{Namespace_: []string{vendor, pluginName, "metric0"}},
	}

	mockFilePath = "./temp_setfile.json"

	mockFileCont = []byte(`{
		 "metric0": {
				"exec": "/bin/sh",
				"type": "string",
				"args": [ "-c", "echo \"test\""]
		},
		"metric1": {
				"exec": "/bin/sh",
				"type": "int64",
				"args": [ "-c", "echo  \"test:1775\" | awk -F':' '{printf $2}'"]
		},
		 "metric2": {
				"exec": "/usr/local/go/bin/go",
				"type": "string",
				"args": ["version"]
		 },
		 "metric3": {
				"exec": "/bin/sh",
				"type": "float64",
				"args": ["-c", "awk 'BEGIN{printf \"%.2f\", (355/100)}'"]
		 },
		"metric4": {
				"exec": "/bin/echo",
				"type": "string",
				"args": ["tes1", "test2"]
		 }
	}
	`)

	mockFileContEmpty = []byte(``)

	mockFileContStructErr = []byte(`{
		 "metric0": {
				"exec": "/bin/sh",
				"type": "string",
				"args": [ "-c", "echo \"test\""]
		},
		"metric1": {
				"exec": "/bin/sh",
				"type": "int64",
				"args": [ "-c", "echo  \"test:1775\" | awk -F':' '{printf $2}'"]
		},
		 "metric2": {
				"exec": "/usr/local/go/bin/go",
				"type": "string",
				"args": ["version"]
		 },
		 "metric3": {
				"exec": "/bin/sh",
				"type": "float64",
				"args": ["-c", "awk 'BEGIN{printf \"%.2f\", (355/100)}'"]
		 },
		"metric4": {
				"exec": "/bin/echo",
				"type": "string",
				"args": ["tes1", "test2"]
		 }
	`)

	mockFileContMapErr = []byte(`{
            "type": "string",
            "command": "echo \"65\""
    }
	`)

	mockFileContMissingType = []byte(`{
		 "metric0": {
				"exec": "/bin/sh",
				"args": [ "-c", "echo \"test\""]
		},
		"metric1": {
				"exec": "/bin/sh",
				"type": "int64",
				"args": [ "-c", "echo  \"test:1775\" | awk -F':' '{printf $2}'"]
		},
		 "metric2": {
				"exec": "/usr/local/go/bin/go",
				"type": "string",
				"args": ["version"]
		 },
		 "metric3": {
				"exec": "/bin/sh",
				"type": "float64",
				"args": ["-c", "awk 'BEGIN{printf \"%.2f\", (355/100)}'"]
		 },
		"metric4": {
				"exec": "/bin/echo",
				"type": "string",
				"args": ["tes1", "test2"]
		 }
	}
	`)

	mockFileContMissingExec = []byte(`{
		 "metric0": {
				"type": "string",
				"args": [ "-c", "echo \"test\""]
		},
		"metric1": {
				"exec": "/bin/sh",
				"type": "int64",
				"args": [ "-c", "echo  \"test:1775\" | awk -F':' '{printf $2}'"]
		},
		 "metric2": {
				"exec": "/usr/local/go/bin/go",
				"type": "string",
				"args": ["version"]
		 },
		 "metric3": {
				"exec": "/bin/sh",
				"type": "float64",
				"args": ["-c", "awk 'BEGIN{printf \"%.2f\", (355/100)}'"]
		 },
		"metric4": {
				"exec": "/bin/echo",
				"type": "string",
				"args": ["tes1", "test2"]
		 }
	}
	`)

	mockFileContExecError = []byte(`{
		 "metric0": {
				"exec": "test1234",
				"type": "string",
				"args": [ "-c", "echo \"test\""]
		},
		"metric1": {
				"exec": "/bin/sh",
				"type": "int64",
				"args": [ "-c", "echo  \"test:1775\" | awk -F':' '{printf $2}'"]
		},
		 "metric2": {
				"exec": "/usr/local/go/bin/go",
				"type": "string",
				"args": ["version"]
		 },
		 "metric3": {
				"exec": "/bin/sh",
				"type": "float64",
				"args": ["-c", "awk 'BEGIN{printf \"%.2f\", (355/100)}'"]
		 },
		"metric4": {
				"exec": "/bin/echo",
				"type": "string",
				"args": ["tes1", "test2"]
		 }
	}
	`)
)
