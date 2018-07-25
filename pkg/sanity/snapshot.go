package sanity

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/libopenstorage/openstorage/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func genSnapName(volid string) string {
	return fmt.Sprintf("%s-snap-%v",
		volid,
		time.Now().Unix())
}

var _ = Describe("Volume [OpenStorageVolume]", func() {
	var (
		c api.OpenStorageVolumeClient
	)

	BeforeEach(func() {
		c = api.NewOpenStorageVolumeClient(conn)
	})

	AfterEach(func() {

	})

	Describe("Volume Snapshot Create", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volID            string
			snapID           string
		)

		BeforeEach(func() {
			numVolumesBefore = numberOfVolumesInCluster(c)
			volID = ""
		})

		AfterEach(func() {
			var err error

			_, err = c.Delete(context.Background(),
				&api.SdkVolumeDeleteRequest{VolumeId: volID},
			)
			Expect(err).ToNot(HaveOccurred())

		})

		It("Should create Volume successfully for snapshot", func() {

			By("Creating the volume")

			var err error
			req := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol",
				Spec: &api.VolumeSpec{
					Size:      uint64(5 * GIGABYTE),
					Shared:    false,
					HaLevel:   3,
					IoProfile: api.IoProfile_IO_PROFILE_DB,
					Cos:       api.CosType_HIGH,
					Format:    api.FSType_FS_TYPE_XFS,
				},
			}

			vResp, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())

			By("Checking if volume created successfully with the provided params")

			// Test Volume Details by Calling Volume Inspect
			inspectReq := &api.SdkVolumeInspectRequest{
				VolumeId: vResp.VolumeId,
			}
			inspectResponse, err := c.Inspect(context.Background(), inspectReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResponse).NotTo(BeNil())

			// Test the details of the created volume
			testVolumeDetails(req, inspectResponse.Volume)
			numVolumesAfter = numberOfVolumesInCluster(c)
			Expect(numVolumesAfter).To(BeEquivalentTo(numVolumesBefore + 1))
			volID = vResp.VolumeId

			By("Creating a snapshot based on the created volume")

			vreq := &api.SdkVolumeSnapshotCreateRequest{
				VolumeId: vResp.VolumeId,
				Name:     genSnapName(vResp.VolumeId),
				Labels: map[string]string{
					"Name": "snapshot-of" + volID,
				},
			}
			resp, err := c.SnapshotCreate(context.Background(), vreq)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.GetSnapshotId()).To(Not(BeNil()))
			snapID = resp.GetSnapshotId()

			By("Checking the Parent field of the created snapshot")

			volumes, err := c.Inspect(context.Background(), &api.SdkVolumeInspectRequest{
				VolumeId: snapID,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(volumes.GetVolume().GetSource().GetParent()).To(BeEquivalentTo(volID))
			// TODO: fake driver does not imlement readonly flag
			//			Expect(volumes.GetVolume().GetReadonly()).To(BeTrue())
		})
	})

	Describe("Volume Snapshot Enumerate", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volID            string
			snapIDs          []string
		)

		BeforeEach(func() {
			numVolumesBefore = numberOfVolumesInCluster(c)
			volID = ""
		})

		AfterEach(func() {
			var err error

			_, err = c.Delete(context.Background(),
				&api.SdkVolumeDeleteRequest{VolumeId: volID},
			)
			Expect(err).ToNot(HaveOccurred())

			for _, snapID := range snapIDs {
				_, err = c.Delete(context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: snapID},
				)
				Expect(err).ToNot(HaveOccurred())
			}

		})

		It("Should enumerate Volume snapshots", func() {

			By("Creating the volume")

			var err error
			req := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol",
				Spec: &api.VolumeSpec{
					Size:      uint64(5 * GIGABYTE),
					Shared:    false,
					HaLevel:   3,
					IoProfile: api.IoProfile_IO_PROFILE_DB,
					Cos:       api.CosType_HIGH,
					Format:    api.FSType_FS_TYPE_XFS,
				},
			}

			vResp, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())

			inspectReq := &api.SdkVolumeInspectRequest{
				VolumeId: vResp.VolumeId,
			}
			inspectResponse, err := c.Inspect(context.Background(), inspectReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResponse).NotTo(BeNil())

			By("Checking if volume created successfully with the provided params")
			testVolumeDetails(req, inspectResponse.Volume)
			numVolumesAfter = numberOfVolumesInCluster(c)
			Expect(numVolumesAfter).To(BeEquivalentTo(numVolumesBefore + 1))

			By("Creating a multiple [3] snapshots based on the created volume")
			volID = vResp.GetVolumeId()
			numOfSnaps := 3

			for i := 0; i < numOfSnaps; i++ {

				snapReq := &api.SdkVolumeSnapshotCreateRequest{
					VolumeId: vResp.VolumeId,
					Name:     genSnapName(vResp.VolumeId),
					Labels: map[string]string{
						"Name": "snapshot-" + strconv.Itoa(i) + "-of" + volID,
					},
				}

				snapResp, err := c.SnapshotCreate(context.Background(), snapReq)
				Expect(err).NotTo(HaveOccurred())
				Expect(snapResp.GetSnapshotId()).To(Not(BeNil()))
				snapIDs = append(snapIDs, snapResp.GetSnapshotId())

				By("Checking the Parent field of the created snapshot")

				volResp, err := c.Inspect(context.Background(), &api.SdkVolumeInspectRequest{
					VolumeId: snapResp.GetSnapshotId(),
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(volResp.GetVolume().GetSource().GetParent()).To(BeEquivalentTo(volID))
			}

			By("Enumerating the snapshots with the volumeID")

			snapEnumResp, err := c.SnapshotEnumerate(context.Background(), &api.SdkVolumeSnapshotEnumerateRequest{
				VolumeId: volID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(snapEnumResp.GetVolumeSnapshotIds())).To(BeEquivalentTo(numOfSnaps))
		})
	})

	Describe("Volume Snapshot Restore", func() {

		var (
			numVolumesBefore int
			numVolumesAfter  int
			volID            string
			snapID           string
		)

		BeforeEach(func() {
			numVolumesBefore = numberOfVolumesInCluster(c)
			volID = ""
		})

		AfterEach(func() {
			var err error

			_, err = c.Delete(context.Background(),
				&api.SdkVolumeDeleteRequest{VolumeId: volID},
			)
			Expect(err).ToNot(HaveOccurred())

		})

		It("Should restore Volume successfully for snapshot", func() {

			By("Creating the volume")

			var err error
			req := &api.SdkVolumeCreateRequest{
				Name: "sdk-vol",
				Spec: &api.VolumeSpec{
					Size:      uint64(5 * GIGABYTE),
					Shared:    false,
					HaLevel:   3,
					IoProfile: api.IoProfile_IO_PROFILE_DB,
					Cos:       api.CosType_HIGH,
					Format:    api.FSType_FS_TYPE_XFS,
				},
			}

			vResp, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			volID = vResp.GetVolumeId()

			inspectReq := &api.SdkVolumeInspectRequest{
				VolumeId: vResp.VolumeId,
			}
			inspectResponse, err := c.Inspect(context.Background(), inspectReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResponse).NotTo(BeNil())

			By("Checking if volume created successfully with the provided params")
			testVolumeDetails(req, inspectResponse.Volume)

			numVolumesAfter = numberOfVolumesInCluster(c)
			Expect(numVolumesAfter).To(BeEquivalentTo(numVolumesBefore + 1))

			By("Creating a snapshot based on the created volume")

			snapReq := &api.SdkVolumeSnapshotCreateRequest{
				VolumeId: volID,
				Name:     genSnapName(volID),
				Labels: map[string]string{
					"Name": "snapshot-of" + volID,
				},
			}

			snapResp, err := c.SnapshotCreate(context.Background(), snapReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapResp.GetSnapshotId()).To(Not(BeNil()))
			snapID = snapResp.GetSnapshotId()

			By("Checking the Parent field of the created snapshot")

			volResp, err := c.Inspect(context.Background(), &api.SdkVolumeInspectRequest{
				VolumeId: snapResp.GetSnapshotId(),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(volResp.GetVolume().GetSource().GetParent()).To(BeEquivalentTo(volID))

			By("Restoring the volume from snapshot")

			_, err = c.SnapshotRestore(context.Background(), &api.SdkVolumeSnapshotRestoreRequest{
				VolumeId:   volID,
				SnapshotId: snapID,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
