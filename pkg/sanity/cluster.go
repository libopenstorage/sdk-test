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

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster [OpenStorageCluster]", func() {
	var (
		c  api.OpenStorageClusterClient
		v  api.OpenStorageVolumeClient
		n  api.OpenStorageNodeClient
		ic api.OpenStorageIdentityClient

		volID string
	)

	BeforeEach(func() {
		c = api.NewOpenStorageClusterClient(conn)
		v = api.NewOpenStorageVolumeClient(conn)
		n = api.NewOpenStorageNodeClient(conn)
		ic = api.NewOpenStorageIdentityClient(conn)

		isSupported := isCapabilitySupported(
			ic,
			api.SdkServiceCapability_OpenStorageService_CLUSTER,
		)

		if !isSupported {
			Skip("Cluster capability not supported , skipping related tests")
		}

		volID = ""
	})

	AfterEach(func() {
		if volID != "" {
			_, err := v.Delete(
				context.Background(),
				&api.SdkVolumeDeleteRequest{VolumeId: volID},
			)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("should return a cluster id", func() {
		info, err := c.InspectCurrent(context.Background(),
			&api.SdkClusterInspectCurrentRequest{})
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Cluster).NotTo(BeNil())
	})

	Describe("Node Enumerate", func() {

		It("Should successfully enumerate nodes", func() {

			enumResp, err := n.Enumerate(
				context.Background(),
				&api.SdkNodeEnumerateRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(enumResp.NodeIds).NotTo(BeEmpty())
		})
	})

	Describe("Node Inspect", func() {

		It("Should inspect all the nodes Successfully", func() {
			By("Enumerating the nodes and getting the node id")
			enumResp, err := n.Enumerate(
				context.Background(),
				&api.SdkNodeEnumerateRequest{},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(enumResp.NodeIds).NotTo(BeEmpty())

			for _, nodeID := range enumResp.NodeIds {

				inspectResp, err := n.Inspect(
					context.Background(),
					&api.SdkNodeInspectRequest{
						NodeId: nodeID,
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(inspectResp).NotTo(BeNil())
				Expect(inspectResp.Node.Id).To(BeEquivalentTo(nodeID))
			}
		})

		It("Should fail to inspect a a non-existent node id", func() {

			inspectResp, err := n.Inspect(
				context.Background(),
				&api.SdkNodeInspectRequest{
					NodeId: "node-id-doesnt-exist",
				},
			)
			Expect(err).To(HaveOccurred())
			Expect(inspectResp).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		})

		It("Should fail to inspect an empty node id", func() {
			inspectResp, err := n.Inspect(
				context.Background(),
				&api.SdkNodeInspectRequest{
					NodeId: "",
				},
			)
			Expect(err).To(HaveOccurred())
			Expect(inspectResp).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Node InspectCurrent", func() {
		It("Should inspect the current node successfully", func() {
			resp, err := n.InspectCurrent(
				context.Background(),
				&api.SdkNodeInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.Node.Id).NotTo(BeNil())
		})
	})
})
