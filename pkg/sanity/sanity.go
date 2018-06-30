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

package sanity

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	config *SanityConfiguration
	conn   *grpc.ClientConn
	lock   sync.Mutex
)

// CloudBackupConfig struct for cloud backup configuration
type CloudBackupConfig struct {
	//CloudProvider string `yaml:"providers"`
	// map[string]string is volume.VolumeParams equivalent
	CloudProviders map[string]map[string]string
}

type SanityConfiguration struct {
	Endpoint string
}

// Test will test start the sanity tests
func Test(t *testing.T, reqConfig *SanityConfiguration) {
	lock.Lock()
	defer lock.Unlock()

	config = reqConfig
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenStorage SDK Test Suite")
}

var _ = BeforeSuite(func() {
	var err error

	By("connecting to OpenStorage SDK endpoint")
	conn, err = utils.Connect(config.Address)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	conn.Close()
})
