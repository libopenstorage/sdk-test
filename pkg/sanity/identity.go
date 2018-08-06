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
	"context"
	"fmt"

	"github.com/libopenstorage/openstorage/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Identity Service", func() {

	var (
		c api.OpenStorageIdentityClient
	)

	BeforeEach(func() {
		c = api.NewOpenStorageIdentityClient(conn)
	})

	Describe("GetPluginCapabilities", func() {
		It("should return appropriate capabilities", func() {
			req := &api.SdkIdentityCapabilitiesRequest{}
			res, err := c.Capabilities(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			By("checking successful response")
			Expect(res.Capabilities).NotTo(BeNil())

			for _, cap := range res.GetCapabilities() {
				switch cap.GetService().GetType() {
				case api.SdkServiceCapability_OpenStorageService_CLOUD_BACKUP:
				case api.SdkServiceCapability_OpenStorageService_CLUSTER:
				case api.SdkServiceCapability_OpenStorageService_CREDENTIALS:
				case api.SdkServiceCapability_OpenStorageService_NODE:
				case api.SdkServiceCapability_OpenStorageService_OBJECT_STORAGE:
				case api.SdkServiceCapability_OpenStorageService_SCHEDULE_POLICY:
				case api.SdkServiceCapability_OpenStorageService_VOLUME:
				default:
					Fail(fmt.Sprintf("Unknown capability: %v\n", cap.GetService().GetType()))
				}
			}
		})
	})
})
