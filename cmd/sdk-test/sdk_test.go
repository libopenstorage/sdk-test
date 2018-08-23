/*
Copyright 2018 Portworx

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
package sdktest

import (
	"flag"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/libopenstorage/sdk-test/pkg/sanity"
	yaml "gopkg.in/yaml.v2"
)

const (
	prefix string = "sdk."
)

var (
	VERSION                 = "(dev)"
	endpoint                string
	mountpath               string
	version                 bool
	cloudProviderConfigPath string
)

func init() {
	flag.StringVar(&endpoint, prefix+"endpoint", "", "OpenStorage SDK endpoint")
	flag.StringVar(&mountpath, prefix+"mountpath", "", "Mount path for volumes")
	flag.BoolVar(&version, prefix+"version", false, "Version of this program")
	flag.StringVar(&cloudProviderConfigPath, prefix+"cpg", "", "Cloud Provider config file , optional")
	flag.Parse()
}

func TestSanity(t *testing.T) {

	var cfg *sanity.CloudProviderConfig
	var err error

	if version {
		fmt.Printf("Version = %s\n", VERSION)
		return
	}
	if len(endpoint) == 0 {
		t.Fatalf("--%s.endpoint must be provided with an OpenStorage SDK endpoint", prefix)
	}

	if len(cloudProviderConfigPath) == 0 {
		t.Logf("No Cloud provider config file provided , Cloud related Tests will be skipped")
	}

	if len(cloudProviderConfigPath) != 0 {
		cfg, err = cloudProviderConfigParse(cloudProviderConfigPath)
		if err != nil {
			t.Logf("Error parsing cloud provider Config , skipping cloud related tests")
		}
	}
	sanity.Test(t, &sanity.SanityConfiguration{
		Address:        endpoint,
		MountPath:      mountpath,
		ProviderConfig: cfg,
	})
}

// cloudProviderConfigParse parses the config file of cloud provider
func cloudProviderConfigParse(filePath string) (*sanity.CloudProviderConfig, error) {

	config := &sanity.CloudProviderConfig{}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read the Cloud provider configuration file (%s): %s", filePath, err.Error())
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("Unable to parse Cloud provider configuration: %s", err.Error())
	}
	return config, nil

}
