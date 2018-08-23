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
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volume [OpenStorageVolume]", func() {
	var (
		c  api.OpenStorageVolumeClient
		ic api.OpenStorageIdentityClient
	)

	BeforeEach(func() {
		c = api.NewOpenStorageVolumeClient(conn)
		ic = api.NewOpenStorageIdentityClient(conn)

		isSupported := isCapabilitySupported(
			ic,
			api.SdkServiceCapability_OpenStorageService_VOLUME,
		)

		if !isSupported {
			Skip("Volume capability not supported , skipping related tests")
		}
	})

	Describe("Volume Create", func() {

		var (
			volID string
		)
		BeforeEach(func() {
			volID = ""
		})

		AfterEach(func() {
			if volID != "" {
				_, err := c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should create a volume and return the volume uuid", func() {

			req := &api.SdkVolumeCreateRequest{
				Name: fmt.Sprintf("sdk-vol-%v", time.Now().Unix()),
				Spec: &api.VolumeSpec{
					Size:      uint64(5 * GIGABYTE),
					Shared:    false,
					HaLevel:   3,
					IoProfile: api.IoProfile_IO_PROFILE_DB,
					Cos:       api.CosType_HIGH,
					Format:    api.FSType_FS_TYPE_XFS,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			// Test Volume Details by Calling Volume Inspect
			inspectReq := &api.SdkVolumeInspectRequest{
				VolumeId: createResponse.VolumeId,
			}
			inspectResponse, err := c.Inspect(context.Background(), inspectReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResponse).NotTo(BeNil())

			// Test the details of the created volume
			testVolumeDetails(req, inspectResponse.Volume)
		})

		It("Return error if volume name is not passed", func() {
			req := &api.SdkVolumeCreateRequest{
				Name: "",
				Spec: &api.VolumeSpec{
					Size: uint64(5 * GIGABYTE),
				},
			}
			info, err := c.Create(context.Background(), req)
			Expect(info).To(BeNil())
			Expect(err).To(HaveOccurred())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})

		It("Should return error if size is zero", func() {
			req := &api.SdkVolumeCreateRequest{
				Name: "zero-size-vol",
				Spec: &api.VolumeSpec{
					Size: uint64(0 * GIGABYTE),
				},
			}
			info, err := c.Create(context.Background(), req)
			Expect(info).To(BeNil())

			Expect(err).To(HaveOccurred())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		})

		// TODO: Fake Driver replaces the volume if a volume name already exists
		// Error code that must be returned is codes.AlreadyExists

		// It("Should return error if creating a volume with same name that exists", func() {

		// 	By("Creating a volume")
		// 	req := &api.SdkVolumeCreateRequest{
		// 		Name: "already-exists-vol",
		// 		Spec: &api.VolumeSpec{
		// 			Size: uint64(5 * GIGABYTE),
		// 		},
		// 	}
		// 	info, err := c.Create(context.Background(), req)
		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(info.VolumeId).NotTo(BeEmpty())

		// 	By("Creating a volume with existing volume name")
		// 	info, err = c.Create(context.Background(), req)

		// 	Expect(err).To(HaveOccurred())
		// 	Expect(info.VolumeId).To(BeEmpty())

		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.AlreadyExists))
		// })
	})

	Describe("Volume Inspect", func() {

		var (
			volID string
		)
		BeforeEach(func() {
			volID = ""
		})

		AfterEach(func() {
			if volID != "" {
				_, err := c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should be able to inspect the created Volume", func() {

			By("Creating a volume")
			req := &api.SdkVolumeCreateRequest{
				Name: "inspect-vol",
				Spec: &api.VolumeSpec{
					Size:    uint64(5 * GIGABYTE),
					HaLevel: 2,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			By("Inspecting the created volume")
			resp, err := c.Inspect(
				context.Background(),
				&api.SdkVolumeInspectRequest{
					VolumeId: volID,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			testVolumeDetails(req, resp.Volume)
		})
	})

	// TODO : Fake driver to throw appropriate response on
	// trying to delete a non existing volume

	// It("Should fail to inspect a non-existing Volume", func() {
	// 	By("Using a volume id that doesn't exist")

	// 	resp, err := c.Inspect(
	// 		context.Background(),
	// 		&api.SdkVolumeInspectRequest{
	// 			VolumeId: "junk-id-doesnt-exist",
	// 		},
	// 	)

	// 	Expect(err).To(HaveOccurred())

	// serverError, ok := status.FromError(err)
	// Expect(ok).To(BeTrue())
	// Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
	// })

	Describe("Volume Delete", func() {

		var (
			volID string
		)
		BeforeEach(func() {
			volID = ""
		})

		It("Should delete the volume successfully", func() {

			By("Creating the volume first")
			req := &api.SdkVolumeCreateRequest{
				Name: "delete-vol",
				Spec: &api.VolumeSpec{
					Size:    uint64(5 * GIGABYTE),
					HaLevel: 3,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			By("Deleting the created volume")
			_, err = c.Delete(
				context.Background(),
				&api.SdkVolumeDeleteRequest{
					VolumeId: volID,
				},
			)

			Expect(err).NotTo(HaveOccurred())
		})

		// TODO: Fake driver should throw appropriate error message when fails
		// to delete a non-existing volume.

		// It("Should throw a error for deleting a non-existent volume", func() {

		// 	_, err := c.Delete(
		// 		context.Background(),
		// 		&api.SdkVolumeDeleteRequest{
		// 			VolumeId: "dummy-id",
		// 		},
		// 	)
		// Expect(err).To(HaveOccurred())

		// serverError, ok := status.FromError(err)
		// Expect(ok).To(BeTrue())
		// Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))

		// })

		It("Should throw a error for passing empty volume id", func() {

			_, err := c.Delete(
				context.Background(),
				&api.SdkVolumeDeleteRequest{
					VolumeId: "",
				},
			)
			Expect(err).To(HaveOccurred())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Volume Enumerate", func() {

		var (
			volIDs []string
		)

		BeforeEach(func() {
		})

		AfterEach(func() {
			for _, id := range volIDs {
				_, err := c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: id},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		// It("Should enumerate all the volumes in the Cluster Successfully", func() {
		// 	By("Creating 5 volumes with same labels")
		// 	numVolumes := 5
		// 	for i := 0; i < numVolumes; i++ {
		// 		req := &api.SdkVolumeCreateRequest{
		// 			Name: "enumerate-vol" + strconv.Itoa(i),
		// 			Spec: &api.VolumeSpec{
		// 				Size: uint64(5 * GIGABYTE),
		// 				VolumeLabels: map[string]string{
		// 					"test": "enumerate",
		// 				},
		// 			},
		// 		}
		// 		createResponse, err := c.Create(context.Background(), req)
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(createResponse).NotTo(BeNil())
		// 		Expect(createResponse.VolumeId).NotTo(BeEmpty())
		// 		id := createResponse.VolumeId
		// 		volIDs = append(volIDs, id)
		// 	}

		// 	By("Creating 5 more volumes with different labels")

		// 	for i := 0; i < numVolumes; i++ {
		// 		req := &api.SdkVolumeCreateRequest{
		// 			Name: "enumerate-vol-different-label" + strconv.Itoa(i),
		// 			Spec: &api.VolumeSpec{
		// 				Size: uint64(5 * GIGABYTE),
		// 				VolumeLabels: map[string]string{
		// 					"test": "enumerate" + strconv.Itoa(i),
		// 				},
		// 			},
		// 		}
		// 		createResponse, err := c.Create(context.Background(), req)
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(createResponse).NotTo(BeNil())
		// 		Expect(createResponse.VolumeId).NotTo(BeEmpty())
		// 		id := createResponse.VolumeId
		// 		volIDs = append(volIDs, id)
		// 	}

		// 	By("Enumerating the volumes that match the label")

		// 	resp, err := c.Enumerate(
		// 		context.Background(),
		// 		&api.SdkVolumeEnumerateRequest{
		// 			Locator: &api.VolumeLocator{
		// 				VolumeLabels: map[string]string{
		// 					"test": "enumerate",
		// 				},
		// 			},
		// 		},
		// 	)

		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(resp).NotTo(BeNil())
		// 	Expect(len(resp.VolumeIds)).To(BeEquivalentTo(numVolumes))

		// 	By("Enumerating all the volumes in the cluster")

		// 	resp, err = c.Enumerate(
		// 		context.Background(),
		// 		&api.SdkVolumeEnumerateRequest{
		// 			Locator: &api.VolumeLocator{
		// 				VolumeLabels: map[string]string{},
		// 			},
		// 		},
		// 	)

		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(resp).NotTo(BeNil())
		// 	Expect(len(resp.VolumeIds)).To(BeEquivalentTo(volumesAfter))
		// })

		It("Should throw appropriate error when failed to enumerate", func() {
			_, err := c.Enumerate(context.Background(), nil)
			Expect(err).To(HaveOccurred())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		})
	})

	Describe("Volume Attach", func() {
		var (
			volID string
		)

		BeforeEach(func() {
			volID = ""
		})

		AfterEach(func() {
			if volID != "" {

				// Detach the attached Volume
				// before deleting
				detachResponse, err := c.Detach(
					context.Background(),
					&api.SdkVolumeDetachRequest{
						VolumeId: volID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(detachResponse).NotTo(BeNil())

				_, err = c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should attach volume successfully", func() {

			By("Creating the volume first")
			req := &api.SdkVolumeCreateRequest{
				Name: "attach-vol",
				Spec: &api.VolumeSpec{
					Size:    uint64(5 * GIGABYTE),
					HaLevel: 1,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			By("Attaching the created Volume")

			resp, err := c.Attach(
				context.Background(),
				&api.SdkVolumeAttachRequest{
					VolumeId: volID,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.DevicePath).NotTo(BeEmpty())
		})

		// TODO: Fake driver to throw error if trying to attach a volume
		// that doesn't exist

		// It("Should throw appropriate error when failed to attach volume", func() {

		// 	By("Passing a non-existent volume id for Attach")

		// 	resp, err := c.Attach(
		// 		context.Background(),
		// 		&api.SdkVolumeAttachRequest{
		// 			VolumeId: "attach-doesnt-exist",
		// 		},
		// 	)

		// 	Expect(err).To(HaveOccurred())
		// 	Expect(resp.DevicePath).To(BeEmpty())
		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		It("Should throw appropriate error when empty volume id is passed", func() {

			By("Passing a empty volume id for Attach")

			resp, err := c.Attach(
				context.Background(),
				&api.SdkVolumeAttachRequest{
					VolumeId: "",
				},
			)

			Expect(err).To(HaveOccurred())
			Expect(resp).To(BeNil())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})

	})

	Describe("Volume Detach", func() {

		var (
			volID string
		)

		BeforeEach(func() {
			volID = ""
		})
		AfterEach(func() {
			if volID != "" {
				_, err := c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should detach the volume successfully", func() {

			By("Creating the volume first")
			req := &api.SdkVolumeCreateRequest{
				Name: "detach-vol",
				Spec: &api.VolumeSpec{
					Size:    uint64(5 * GIGABYTE),
					HaLevel: 3,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			By("Attaching the created Volume")

			resp, err := c.Attach(
				context.Background(),
				&api.SdkVolumeAttachRequest{
					VolumeId: volID,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.DevicePath).NotTo(BeEmpty())

			By("Detaching the attached volume")

			detachResponse, err := c.Detach(
				context.Background(),
				&api.SdkVolumeDetachRequest{
					VolumeId: volID,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(detachResponse).NotTo(BeNil())
		})

		// TODO: Fake driver to throw an error if trying to detach
		// an non-attached volume

		// It("Should fail to detach a non-attached volume", func() {
		// 	By("Creating the volume first")
		// 	req := &api.SdkVolumeCreateRequest{
		// 		Name: "detach-vol-non-attached",
		// 		Spec: &api.VolumeSpec{
		// 			Size: uint64(5 * GIGABYTE),
		// 		},
		// 	}
		// 	createResponse, err := c.Create(context.Background(), req)
		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(createResponse).NotTo(BeNil())
		// 	Expect(createResponse.VolumeId).NotTo(BeEmpty())
		// 	volID = createResponse.VolumeId

		// 	By("Detaching a non-attached volume")

		// 	_, err = c.Detach(
		// 		context.Background(),
		// 		&api.SdkVolumeDetachRequest{
		// 			VolumeId: volID,
		// 		},
		// 	)

		// 	Expect(err).To(HaveOccurred())
		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		// It("Should fail to detach a non-existent volume", func() {

		// 	By("Detaching a non-existent volume")

		// 	_, err := c.Detach(
		// 		context.Background(),
		// 		&api.SdkVolumeDetachRequest{
		// 			VolumeId: "dummy-doesn't exist",
		// 		},
		// 	)

		// 	Expect(err).To(HaveOccurred())
		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		It("Should fail with bad argument of empty volume id", func() {

			By("Detaching a non-existent volume")

			_, err := c.Detach(
				context.Background(),
				&api.SdkVolumeDetachRequest{
					VolumeId: "",
				},
			)
			Expect(err).To(HaveOccurred())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Volume Mount", func() {
		var (
			volID string
		)

		BeforeEach(func() {
			volID = ""

			if len(config.MountPath) == 0 {
				Skip("Mount path was not provided")
			}
		})

		AfterEach(func() {
			if volID != "" {
				// Unmount the mounted volume first
				// before deleting
				unmountResponse, err := c.Unmount(
					context.Background(),
					&api.SdkVolumeUnmountRequest{
						VolumeId:  volID,
						MountPath: config.MountPath,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(unmountResponse).NotTo(BeNil())
				// Detach the attached Volume
				// before deleting
				detachResponse, err := c.Detach(
					context.Background(),
					&api.SdkVolumeDetachRequest{
						VolumeId: volID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(detachResponse).NotTo(BeNil())

				_, err = c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should Mount the volume successfully", func() {
			By("Creating the volume first")
			req := &api.SdkVolumeCreateRequest{
				Name: "mount-vol",
				Spec: &api.VolumeSpec{
					Size:      uint64(5 * GIGABYTE),
					HaLevel:   3,
					IoProfile: api.IoProfile_IO_PROFILE_DB,
					Cos:       api.CosType_HIGH,
					Format:    api.FSType_FS_TYPE_XFS,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			By("Attaching the created Volume")

			resp, err := c.Attach(
				context.Background(),
				&api.SdkVolumeAttachRequest{
					VolumeId: volID,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.DevicePath).NotTo(BeEmpty())

			By("Mounting the attached Volume")

			mountResponse, err := c.Mount(
				context.Background(),
				&api.SdkVolumeMountRequest{
					VolumeId:  volID,
					MountPath: config.MountPath,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(mountResponse).NotTo(BeNil())
		})

		// TODO: Fake driver to throw appropriate error when
		// mounting a non-attached volume

		// It("Should fail to mount a non attached volume", func() {

		// 	By("Creating the volume first")
		// 	req := &api.SdkVolumeCreateRequest{
		// 		Name: "mount-vol",
		// 		Spec: &api.VolumeSpec{
		// 			Size: uint64(5 * GIGABYTE),
		// 		},
		// 	}
		// 	createResponse, err := c.Create(context.Background(), req)
		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(createResponse).NotTo(BeNil())
		// 	Expect(createResponse.VolumeId).NotTo(BeEmpty())
		// 	volID = createResponse.VolumeId

		// 	By("Attaching the created Volume")

		// 	resp, err := c.Attach(
		// 		context.Background(),
		// 		&api.SdkVolumeAttachRequest{
		// 			VolumeId: volID,
		// 		},
		// 	)
		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(resp.DevicePath).NotTo(BeEmpty())

		// 	By("Mounting a non-attached Volume")

		// 	_, err = c.Mount(
		// 		context.Background(),
		// 		&api.SdkVolumeMountRequest{
		// 			VolumeId:  volID,
		// 			MountPath: config.MountPath,
		// 		},
		// 	)
		// 	Expect(err).To(HaveOccurred())
		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		// TODO: Error code should be internal
		// Fix in Fakedriver.

		// It("Should fail to mount a non-existent volume", func() {

		// 	By("Mounting a non-existent Volume")

		// 	_, err := c.Mount(
		// 		context.Background(),
		// 		&api.SdkVolumeMountRequest{
		// 			VolumeId:  "dummy-doesnt-exist",
		// 			MountPath: config.MountPath,
		// 		},
		// 	)
		// 	Expect(err).To(HaveOccurred())
		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		It("Should fail to mount on empty volume id", func() {

			By("Mounting a Volume with empty id")

			_, err := c.Mount(
				context.Background(),
				&api.SdkVolumeMountRequest{
					VolumeId:  "",
					MountPath: config.MountPath,
				},
			)
			Expect(err).To(HaveOccurred())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Volume Clone", func() {
		var (
			volID    string
			clonedID string
		)

		BeforeEach(func() {
			volID = ""
			clonedID = ""
		})

		AfterEach(func() {
			if volID != "" {
				_, err := c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
				Expect(err).NotTo(HaveOccurred())

				if clonedID != "" {
					_, err = c.Delete(
						context.Background(),
						&api.SdkVolumeDeleteRequest{VolumeId: clonedID},
					)
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})

		It("Should clone the volume successfully", func() {
			By("Creating the volume first")
			req := &api.SdkVolumeCreateRequest{
				Name: "volume-to-be-cloned",
				Spec: &api.VolumeSpec{
					Size:    uint64(5 * GIGABYTE),
					HaLevel: 2,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			By("Cloning the volume")

			cloneRespose, err := c.Clone(
				context.Background(),
				&api.SdkVolumeCloneRequest{
					Name:     "cloned-vol",
					ParentId: volID,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(cloneRespose).NotTo(BeNil())
			Expect(cloneRespose.VolumeId).NotTo(BeEmpty())
			clonedID = cloneRespose.VolumeId
		})

		// TODO: Fake Driver should throw error on size mismatch from clone.
		// Also Clone should only take the parent Id and map[string]string for volume
		// labels , that way user won't commit the mistake to redefining the attrributes
		// in the Volume Spec , which may be different from the Parent volume

		// 	It("Should fail to clone if clone size is different", func() {
		// 		By("Creating the volume first")
		// 		req := &api.SdkVolumeCreateRequest{
		// 			Name: "volume-to-be-cloned",
		// 			Spec: &api.VolumeSpec{
		// 				Size: uint64(5 * GIGABYTE),
		// 			},
		// 		}
		// 		createResponse, err := c.Create(context.Background(), req)
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(createResponse).NotTo(BeNil())
		// 		Expect(createResponse.VolumeId).NotTo(BeEmpty())
		// 		volID = createResponse.VolumeId

		// 		By("Cloning the volume")

		// 		cloneRespose, err := c.Clone(
		// 			context.Background(),
		// 			&api.SdkVolumeCloneRequest{
		// 				Name:     "cloned-vol",
		// 				ParentId: volID,
		// 				Spec: &api.VolumeSpec{
		// 					Size: uint64(10 * GIGABYTE),
		// 					VolumeLabels: map[string]string{
		// 						"test": "clone",
		// 					},
		// 				},
		// 			},
		// 		)
		// 		Expect(err).To(HaveOccurred())
		// 		Expect(cloneRespose).To(BeNil())
		// 		serverError, ok := status.FromError(err)
		// 		Expect(ok).To(BeTrue())
		// 		Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// 	})
	})
	Describe("Volume stats", func() {
		var (
			volID string
		)

		BeforeEach(func() {
			volID = ""
		})

		AfterEach(func() {
			if volID != "" {
				_, err := c.Delete(
					context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should retrieve stats of volume successfully", func() {
			By("Creating the volume first")
			req := &api.SdkVolumeCreateRequest{
				Name: "mount-vol",
				Spec: &api.VolumeSpec{
					Size:      uint64(5 * GIGABYTE),
					HaLevel:   3,
					IoProfile: api.IoProfile_IO_PROFILE_DB,
					Cos:       api.CosType_HIGH,
					Format:    api.FSType_FS_TYPE_XFS,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			statsResp, err := c.Stats(
				context.Background(),
				&api.SdkVolumeStatsRequest{
					VolumeId:      volID,
					NotCumulative: true,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(statsResp.Stats).NotTo(BeNil())
		})

		It("Should retrieve stats of volume successfully for  cumulative", func() {
			By("Creating the volume first")
			req := &api.SdkVolumeCreateRequest{
				Name: "stats-vol",
				Spec: &api.VolumeSpec{
					Size:      uint64(5 * GIGABYTE),
					HaLevel:   3,
					IoProfile: api.IoProfile_IO_PROFILE_DB,
					Cos:       api.CosType_HIGH,
					Format:    api.FSType_FS_TYPE_XFS,
				},
			}
			createResponse, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(createResponse).NotTo(BeNil())
			Expect(createResponse.VolumeId).NotTo(BeEmpty())
			volID = createResponse.VolumeId

			statsResp, err := c.Stats(
				context.Background(),
				&api.SdkVolumeStatsRequest{
					VolumeId:      volID,
					NotCumulative: false,
				},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(statsResp.Stats).NotTo(BeNil())
		})

		It("Should fail to retrieve stats of non-existent volume", func() {
			Skip("PWX-6056")

			statsResp, err := c.Stats(
				context.Background(),
				&api.SdkVolumeStatsRequest{
					VolumeId:      "volID-doesnt-exist",
					NotCumulative: true,
				},
			)

			Expect(err).To(HaveOccurred())
			Expect(statsResp).To(BeNil())
		})

		It("Should fail to retrieve stats of empty volume", func() {
			statsResp, err := c.Stats(
				context.Background(),
				&api.SdkVolumeStatsRequest{
					VolumeId:      volID,
					NotCumulative: true,
				},
			)

			Expect(err).To(HaveOccurred())
			Expect(statsResp).To(BeNil())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})
})
