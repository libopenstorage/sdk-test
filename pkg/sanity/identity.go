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

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"

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

	Describe("Version", func() {
		It("should return version information", func() {
			res, err := c.Version(context.Background(), &api.SdkIdentityVersionRequest{})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(res.GetSdkVersion()).NotTo(BeNil())
			Expect(res.GetSdkVersion().GetVersion()).NotTo(BeEmpty())
			Expect(res.GetVersion()).NotTo(BeNil())

			By("Checking SDK version")
			Expect(res.GetSdkVersion().GetMajor()).To(BeNumerically(">=", 0))
			Expect(res.GetSdkVersion().GetMinor()).To(BeNumerically(">=", 0))
			Expect(res.GetSdkVersion().GetPatch()).To(BeNumerically(">=", 0))

			expectVersion := fmt.Sprintf("%d.%d.%d",
				res.GetSdkVersion().GetMajor(),
				res.GetSdkVersion().GetMinor(),
				res.GetSdkVersion().GetPatch())
			Expect(res.GetSdkVersion().GetVersion()).To(Equal(expectVersion))

			By("Checking driver version")
			Expect(res.GetVersion().GetDriver()).NotTo(BeEmpty())
			Expect(res.GetVersion().GetVersion()).NotTo(BeEmpty())
		})
	})

	Describe("Capabilities", func() {
		It("should return appropriate capabilities", func() {
			req := &api.SdkIdentityCapabilitiesRequest{}
			res, err := c.Capabilities(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())

			By("checking successful response")
			Expect(res.Capabilities).NotTo(BeNil())
			Expect(res.Capabilities).NotTo(BeEmpty())
		})
	})
})
