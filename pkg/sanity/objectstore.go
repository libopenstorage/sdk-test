package sanity

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Objectstore Features[OpenStorageObjectstore]", func() {
	var (
		objClient api.OpenStorageObjectstoreClient
		volClient api.OpenStorageVolumeClient
		volID     string
	)

	BeforeEach(func() {
		objClient = api.NewOpenStorageObjectstoreClient(conn)
		volClient = api.NewOpenStorageVolumeClient(conn)

	})
	AfterEach(func() {
		if volID != "" {
			_, err := volClient.Delete(
				context.Background(),
				&api.SdkVolumeDeleteRequest{VolumeId: volID},
			)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Objectstore Create", func() {
		It("Should create objectstore with given volume ID", func() {
			volReq := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol",
				Spec: &api.VolumeSpec{
					Size:             uint64(5 * GIGABYTE),
					AggregationLevel: 2,
					Encrypted:        true,
					Shared:           false,
					HaLevel:          3,
					IoProfile:        api.IoProfile_IO_PROFILE_DB,
					Cos:              api.CosType_HIGH,
					Sticky:           true,
					Format:           api.FSType_FS_TYPE_XFS,
				},
			}
			volResp, err := volClient.Create(context.Background(), volReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(volResp).NotTo(BeNil())
			Expect(volResp.VolumeId).NotTo(BeEmpty())
			volID = volResp.VolumeId

			// Create objectstore using given volume ID
			objReq := &api.SdkObjectstoreCreateRequest{
				VolumeId: volID}

			objResp, err := objClient.Create(context.Background(), objReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(objResp).NotTo(BeNil())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).NotTo(BeEmpty())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).To(BeEquivalentTo(volID))

		})

		It("Should failed to create objectstore with empty volume ID", func() {
			volReq := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol",
				Spec: &api.VolumeSpec{
					Size:             uint64(5 * GIGABYTE),
					AggregationLevel: 2,
					Encrypted:        true,
					Shared:           false,
					HaLevel:          3,
					IoProfile:        api.IoProfile_IO_PROFILE_DB,
					Cos:              api.CosType_HIGH,
					Sticky:           true,
					Format:           api.FSType_FS_TYPE_XFS,
				},
			}
			volResp, err := volClient.Create(context.Background(), volReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(volResp).NotTo(BeNil())
			Expect(volResp.VolumeId).NotTo(BeEmpty())
			volID = volResp.VolumeId

			// Create objectstore using empty volume ID
			objReq := &api.SdkObjectstoreCreateRequest{
				VolumeId: ""}

			objResp, err := objClient.Create(context.Background(), objReq)
			Expect(err).To(HaveOccurred())
			Expect(objResp).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Objectstore Update", func() {
		It("Should update objectstore status (start/stop)", func() {
			volReq := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol-test1",
				Spec: &api.VolumeSpec{
					Size:             uint64(5 * GIGABYTE),
					AggregationLevel: 2,
					Encrypted:        true,
					Shared:           false,
					HaLevel:          3,
					IoProfile:        api.IoProfile_IO_PROFILE_DB,
					Cos:              api.CosType_HIGH,
					Sticky:           true,
					Format:           api.FSType_FS_TYPE_XFS,
				},
			}
			volResp, err := volClient.Create(context.Background(), volReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(volResp).NotTo(BeNil())
			Expect(volResp.VolumeId).NotTo(BeEmpty())
			volID = volResp.VolumeId

			// Create objectstore using given volume ID
			objReq := &api.SdkObjectstoreCreateRequest{
				VolumeId: volID}

			objResp, err := objClient.Create(context.Background(), objReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(objResp).NotTo(BeNil())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).NotTo(BeEmpty())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).To(BeEquivalentTo(volID))
			Expect(objResp.GetObjectstoreStatus().GetUuid()).NotTo(BeEmpty())

			// Update objectstore status to true, by default it's false when
			// objectstore is created
			updateReq := &api.SdkObjectstoreUpdateRequest{
				ObjectstoreId: objResp.GetObjectstoreStatus().GetUuid(),
				Enable:        true,
			}

			_, err = objClient.Update(context.Background(), updateReq)
			Expect(err).NotTo(HaveOccurred())

			inspectReq := &api.SdkObjectstoreInspectRequest{
				ObjectstoreId: objResp.GetObjectstoreStatus().GetUuid(),
			}

			inspectResp, err := objClient.Inspect(context.Background(), inspectReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResp).NotTo(BeNil())
			Expect(inspectResp.GetObjectstoreStatus().GetUuid()).NotTo(BeEmpty())
			Expect(inspectResp.GetObjectstoreStatus().GetUuid()).To(BeEquivalentTo(updateReq.ObjectstoreId))
			Expect(inspectResp.GetObjectstoreStatus().GetEnabled()).To(BeEquivalentTo(updateReq.Enable))

		})
	})

	Describe("Objectstore Delete", func() {
		It("Should delete objectstore with given UUID", func() {
			volReq := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol-test1",
				Spec: &api.VolumeSpec{
					Size:             uint64(5 * GIGABYTE),
					AggregationLevel: 2,
					Encrypted:        true,
					Shared:           false,
					HaLevel:          3,
					IoProfile:        api.IoProfile_IO_PROFILE_DB,
					Cos:              api.CosType_HIGH,
					Sticky:           true,
					Format:           api.FSType_FS_TYPE_XFS,
				},
			}
			volResp, err := volClient.Create(context.Background(), volReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(volResp).NotTo(BeNil())
			Expect(volResp.VolumeId).NotTo(BeEmpty())
			volID = volResp.VolumeId

			// Create objectstore using given volume ID
			objReq := &api.SdkObjectstoreCreateRequest{
				VolumeId: volID}

			objResp, err := objClient.Create(context.Background(), objReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(objResp).NotTo(BeNil())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).NotTo(BeEmpty())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).To(BeEquivalentTo(volID))
			Expect(objResp.GetObjectstoreStatus().GetUuid()).NotTo(BeEmpty())

			// Delete object store
			deleteReq := &api.SdkObjectstoreDeleteRequest{
				ObjectstoreId: objResp.GetObjectstoreStatus().GetUuid(),
			}

			_, err = objClient.Delete(context.Background(), deleteReq)
			Expect(err).NotTo(HaveOccurred())

			inspectReq := &api.SdkObjectstoreInspectRequest{
				ObjectstoreId: objResp.GetObjectstoreStatus().GetUuid(),
			}

			// Inspect should failed for given objectstore
			inspectResp, err := objClient.Inspect(context.Background(), inspectReq)
			Expect(err).To(HaveOccurred())
			Expect(inspectResp).To(BeNil())
		})

		//TODO : add support in fake driver for non-existance objectstore id
		/*It("Should failed to delete objectstore with empty UUID", func() {
			// Delete object store
			deleteReq := &api.SdkObjectstoreDeleteRequest{
				ObjectstoreId: "invalid",
			}

			_, err := objClient.Delete(context.Background(), deleteReq)
			Expect(err).To(HaveOccurred())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))

		})*/
	})

	Describe("Objectstore Inspect", func() {
		It("Should inspect objectstore with given UUID", func() {
			volReq := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol-test1",
				Spec: &api.VolumeSpec{
					Size:             uint64(5 * GIGABYTE),
					AggregationLevel: 2,
					Encrypted:        true,
					Shared:           false,
					HaLevel:          3,
					IoProfile:        api.IoProfile_IO_PROFILE_DB,
					Cos:              api.CosType_HIGH,
					Sticky:           true,
					Format:           api.FSType_FS_TYPE_XFS,
				},
			}
			volResp, err := volClient.Create(context.Background(), volReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(volResp).NotTo(BeNil())
			Expect(volResp.VolumeId).NotTo(BeEmpty())
			volID = volResp.VolumeId

			// Create objectstore using given volume ID
			objReq := &api.SdkObjectstoreCreateRequest{
				VolumeId: volID}

			objResp, err := objClient.Create(context.Background(), objReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(objResp).NotTo(BeNil())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).NotTo(BeEmpty())
			Expect(objResp.GetObjectstoreStatus().GetVolumeId()).To(BeEquivalentTo(volID))
			Expect(objResp.GetObjectstoreStatus().GetUuid()).NotTo(BeEmpty())

			inspectReq := &api.SdkObjectstoreInspectRequest{
				ObjectstoreId: objResp.GetObjectstoreStatus().GetUuid(),
			}

			// Inspect should failed for given objectstore
			inspectResp, err := objClient.Inspect(context.Background(), inspectReq)
			// verify inspect response with create response
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResp).NotTo(BeNil())
			Expect(inspectResp.GetObjectstoreStatus().GetUuid()).NotTo(BeEmpty())
			Expect(inspectResp.GetObjectstoreStatus().GetUuid()).To(BeEquivalentTo(objResp.GetObjectstoreStatus().GetUuid()))
			Expect(inspectResp.GetObjectstoreStatus().GetEnabled()).To(BeEquivalentTo(objResp.GetObjectstoreStatus().GetEnabled()))
			Expect(inspectResp.GetObjectstoreStatus().GetVolumeId()).To(BeEquivalentTo(objResp.GetObjectstoreStatus().GetVolumeId()))

		})

		// TODO: add check in fake driver for non-existance objectstore UUID
		/*
			It("Should fail inspect objectstore with invalid objectstore UUID", func() {

				inspectReq := &api.SdkObjectstoreInspectRequest{
					ObjectstoreId: "invalid-uuid-1",
				}

				// Inspect should failed for given objectstore
				inspectResp, err := objClient.Inspect(context.Background(), inspectReq)
				// verify inspect response with create response
				Expect(err).To(HaveOccurred())
				Expect(inspectResp).To(BeNil())

			})*/
	})

})
