/*
Copyright 2019 Portworx

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
	"time"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ownership Test Suite", func() {

	var (
		vc       api.OpenStorageVolumeClient
		user1vol string
	)

	BeforeEach(func() {
		if len(config.SharedSecret) == 0 {
			Skip("Not running with authentication")
		}
		vc = api.NewOpenStorageVolumeClient(conn)

		By("creating a volume for user1")
		respv, err := vc.Create(
			setContextWithToken(context.Background(), users["user1"]),
			&api.SdkVolumeCreateRequest{
				Name: fmt.Sprintf("sdk-vol-%v", time.Now().Unix()),
				Spec: &api.VolumeSpec{
					Size:    uint64(1 * GIGABYTE),
					HaLevel: 1,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			})
		Expect(err).ToNot(HaveOccurred())
		Expect(respv).NotTo(BeNil())

		user1vol = respv.GetVolumeId()
		Expect(user1vol).NotTo(BeEmpty())
	})

	AfterEach(func() {
		By("cleaning up volume for user1")
		err := deleteVol(
			setContextWithToken(context.Background(), users["user1"]),
			vc,
			user1vol)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should have ownership set in volume", func() {
		By("checking it sets the ownership accordingly")
		respInspectVol, err := vc.Inspect(
			setContextWithToken(context.Background(), users["user1"]),
			&api.SdkVolumeInspectRequest{
				VolumeId: user1vol,
			})
		Expect(err).ToNot(HaveOccurred())
		Expect(respInspectVol).NotTo(BeNil())
		Expect(respInspectVol.GetVolume()).NotTo(BeNil())
		Expect(respInspectVol.GetVolume().GetSpec()).NotTo(BeNil())
		o := respInspectVol.GetVolume().GetSpec().GetOwnership()
		Expect(o).NotTo(BeNil())
		Expect(o.GetOwner()).To(BeEquivalentTo("user1"))
	})

	It("should allow the owner to set acls but not others", func() {

		By("user2 unable to see the volume in Enumerate")
		respEnum, err := vc.Enumerate(
			setContextWithToken(context.Background(), users["user2"]),
			&api.SdkVolumeEnumerateRequest{})
		Expect(err).ToNot(HaveOccurred())
		Expect(respEnum).NotTo(BeNil())
		Expect(respEnum.GetVolumeIds()).NotTo(ContainElement(user1vol))

		By("admin able to see the volume in Enumerate")
		respEnum, err = vc.Enumerate(
			setContextWithToken(context.Background(), users["admin"]),
			&api.SdkVolumeEnumerateRequest{})
		Expect(err).ToNot(HaveOccurred())
		Expect(respEnum).NotTo(BeNil())
		Expect(respEnum.GetVolumeIds()).To(ContainElement(user1vol))

		By("setting group access to 'users'")
		// Owner setting value
		_, err = vc.Update(
			setContextWithToken(context.Background(), users["user1"]),
			&api.SdkVolumeUpdateRequest{
				VolumeId: user1vol,
				Spec: &api.VolumeSpecUpdate{
					Ownership: &api.Ownership{
						Acls: &api.Ownership_AccessControl{
							Groups: []string{"users"},
						},
					},
				},
			})

		By("checking it sets the ownership accordingly")
		respInspectVol, err := vc.Inspect(
			setContextWithToken(context.Background(), users["user1"]),
			&api.SdkVolumeInspectRequest{
				VolumeId: user1vol,
			})
		Expect(err).ToNot(HaveOccurred())
		Expect(respInspectVol).NotTo(BeNil())
		Expect(respInspectVol.GetVolume()).NotTo(BeNil())
		Expect(respInspectVol.GetVolume().GetSpec()).NotTo(BeNil())
		o := respInspectVol.GetVolume().GetSpec().GetOwnership()
		Expect(o).NotTo(BeNil())
		Expect(o.GetOwner()).To(BeEquivalentTo("user1"))
		Expect(o.GetAcls()).NotTo(BeNil())
		Expect(o.GetAcls().GetGroups()).To(HaveLen(1))
		Expect(o.GetAcls().GetGroups()).To(ContainElement("users"))

		By("user2 able to see the volume in Enumerate")
		respEnum, err = vc.Enumerate(
			setContextWithToken(context.Background(), users["user2"]),
			&api.SdkVolumeEnumerateRequest{})
		Expect(err).ToNot(HaveOccurred())
		Expect(respEnum).NotTo(BeNil())
		Expect(respEnum.GetVolumeIds()).To(ContainElement(user1vol))

		By("admin able to see the volume in Enumerate")
		respEnum, err = vc.Enumerate(
			setContextWithToken(context.Background(), users["admin"]),
			&api.SdkVolumeEnumerateRequest{})
		Expect(err).ToNot(HaveOccurred())
		Expect(respEnum).NotTo(BeNil())
		Expect(respEnum.GetVolumeIds()).To(ContainElement(user1vol))

		By("user2 unable to update the volume")
		// user2 is in group users but is not the owner
		_, err = vc.Update(
			setContextWithToken(context.Background(), users["user2"]),
			&api.SdkVolumeUpdateRequest{
				VolumeId: user1vol,
				Spec: &api.VolumeSpecUpdate{
					Ownership: &api.Ownership{
						Acls: &api.Ownership_AccessControl{
							Groups: []string{"users", "anothergroup"},
						},
					},
				},
			})
		Expect(err).To(HaveOccurred())
		serverError, ok := status.FromError(err)
		Expect(ok).To(BeTrue())
		Expect(serverError.Code()).To(BeEquivalentTo(codes.PermissionDenied))

		By("admin adding group access to 'others' group")
		// Owner setting value
		_, err = vc.Update(
			setContextWithToken(context.Background(), users["admin"]),
			&api.SdkVolumeUpdateRequest{
				VolumeId: user1vol,
				Spec: &api.VolumeSpecUpdate{
					Ownership: &api.Ownership{
						Acls: &api.Ownership_AccessControl{
							Groups: []string{"users", "others"},
						},
					},
				},
			})

		By("user1 checking it group 'others' was added")
		respInspectVol, err = vc.Inspect(
			setContextWithToken(context.Background(), users["user1"]),
			&api.SdkVolumeInspectRequest{
				VolumeId: user1vol,
			})
		Expect(err).ToNot(HaveOccurred())
		Expect(respInspectVol).NotTo(BeNil())
		Expect(respInspectVol.GetVolume()).NotTo(BeNil())
		Expect(respInspectVol.GetVolume().GetSpec()).NotTo(BeNil())
		o = respInspectVol.GetVolume().GetSpec().GetOwnership()
		Expect(o).NotTo(BeNil())
		Expect(o.GetOwner()).To(BeEquivalentTo("user1"))
		Expect(o.GetAcls()).NotTo(BeNil())
		Expect(o.GetAcls().GetGroups()).To(HaveLen(2))
		Expect(o.GetAcls().GetGroups()).To(ContainElement("users"))
		Expect(o.GetAcls().GetGroups()).To(ContainElement("others"))
	})
})
