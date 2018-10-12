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
	"strconv"
	"time"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"

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
		c  api.OpenStorageVolumeClient
		ic api.OpenStorageIdentityClient
		sc api.OpenStorageSchedulePolicyClient
	)

	BeforeEach(func() {
		c = api.NewOpenStorageVolumeClient(conn)
		ic = api.NewOpenStorageIdentityClient(conn)
		sc = api.NewOpenStorageSchedulePolicyClient(conn)

		isSupported := isCapabilitySupported(
			ic,
			api.SdkServiceCapability_OpenStorageService_VOLUME,
		)

		if !isSupported {
			Skip("Volume Snapshot capability not supported , skipping related tests")
		}
	})

	AfterEach(func() {

	})

	Describe("Volume Snapshot Create", func() {

		var (
			volID  string
			snapID string
		)

		BeforeEach(func() {
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
			volID   string
			snapIDs []string
		)

		BeforeEach(func() {
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
			volID  string
			snapID string
		)

		BeforeEach(func() {
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

	XDescribe("VolumeSnapshotScheduleUpdate", func() {

		var (
			volID      string
			policyName string
		)

		BeforeEach(func() {
			volID = ""
			policyName = ""
		})

		AfterEach(func() {
			var err error

			if len(volID) != 0 {
				_, err = c.Delete(context.Background(),
					&api.SdkVolumeDeleteRequest{VolumeId: volID},
				)
			}
			if len(policyName) != 0 {
				_, err = sc.Delete(context.Background(),
					&api.SdkSchedulePolicyDeleteRequest{Name: policyName})
			}
			Expect(err).ToNot(HaveOccurred())
		})

		It("should set the schedule in the volume spec", func() {

			By("Creating the volume")
			var err error
			req := &api.SdkVolumeCreateRequest{
				Name: fmt.Sprintf("sdk-vol-%d", time.Now().Unix()),
				Spec: &api.VolumeSpec{
					Size:    uint64(5 * GIGABYTE),
					HaLevel: 3,
					Format:  api.FSType_FS_TYPE_EXT4,
				},
			}
			vResp, err := c.Create(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())
			volID = vResp.GetVolumeId()

			By("Creating a schedule policy")
			policyName = fmt.Sprintf("mypolicy-%d", time.Now().Unix())
			policyReq := &api.SdkSchedulePolicyCreateRequest{
				SchedulePolicy: &api.SdkSchedulePolicy{
					Name: policyName,
					Schedules: []*api.SdkSchedulePolicyInterval{
						&api.SdkSchedulePolicyInterval{
							Retain: 2,
							PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
								Weekly: &api.SdkSchedulePolicyIntervalWeekly{
									Day:    api.SdkTimeWeekday_SdkTimeWeekdaySunday,
									Hour:   0,
									Minute: 30,
								},
							},
						},
					},
				},
			}
			_, err = sc.Create(context.Background(), policyReq)
			Expect(err).NotTo(HaveOccurred())

			By("Confirming the schedule name is in the volume spec")
			inspectResponse, err := c.Inspect(context.Background(), &api.SdkVolumeInspectRequest{
				VolumeId: vResp.VolumeId,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResponse).NotTo(BeNil())
			Expect(inspectResponse.GetVolume().GetSpec().GetSnapshotSchedule()).To(Equal(fmt.Sprintf("policy=%s", policyName)))
		})
	})

})
