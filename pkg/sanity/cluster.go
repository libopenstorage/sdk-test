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
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"

	"context"
	"github.com/libopenstorage/openstorage/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Enumerate [OpenStorageCluster]", func() {
	var (
		c api.OpenStorageClusterClient
	)

	BeforeEach(func() {
		c = api.NewOpenStorageClusterClient(conn)
	})

	It("should return a cluster id", func() {
		info, err := c.Enumerate(context.Background(),
			&api.SdkClusterEnumerateRequest{})
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Cluster).NotTo(BeNil())
	})
})
