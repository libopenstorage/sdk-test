package sanity

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cloud backup [OpenStorageCluster]", func() {
	var (
		cc api.OpenStorageCredentialsClient
		vc api.OpenStorageVolumeClient
		bc api.OpenStorageCloudBackupClient

		volID        string
		credID       string
		credsUUIDMap map[string]string
	)

	BeforeEach(func() {

		cc = api.NewOpenStorageCredentialsClient(conn)
		bc = api.NewOpenStorageCloudBackupClient(conn)
		vc = api.NewOpenStorageVolumeClient(conn)

		volID = ""
		credID = ""
		credsUUIDMap = make(map[string]string)

		if config.ProviderConfig == nil {
			Skip("Skipping cloud backup tests")
		}
	})

	AfterEach(func() {
		if len(credsUUIDMap) != 0 {
			for _, uuid := range credsUUIDMap {
				credID = uuid

				_, err := cc.Delete(
					context.Background(),
					&api.SdkCredentialDeleteRequest{
						CredentialId: credID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
			}
		}

		if volID != "" {
			_, err := vc.Detach(
				context.Background(),
				&api.SdkVolumeDetachRequest{
					VolumeId: volID,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			_, err = vc.Delete(
				context.Background(),
				&api.SdkVolumeDeleteRequest{VolumeId: volID},
			)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Describe("Backup Schedule	 Create", func() {

		It("Should create a cloud backup schedule successfully", func() {
			By("First creating the volume")
			volID = newTestVolume(vc)

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Attaching the created volume")
				str, err := vc.Attach(
					context.Background(),
					&api.SdkVolumeAttachRequest{
						VolumeId: volID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				By("Creating a backup schedule on " + provider)

				schedule := &api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Hour:   0,
							Minute: 30,
						},
					},
				}

				schedCreateResp, err := bc.SchedCreate(
					context.Background(),
					&api.SdkCloudBackupSchedCreateRequest{
						CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
							CredentialId: credID,
							MaxBackups:   3,
							SrcVolumeId:  volID,
							Schedule:     schedule,
						},
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(schedCreateResp.BackupScheduleId).NotTo(BeNil())
			}
		})

		It("Should fail to create back up schedule if non-existent volume id is passed", func() {
			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Creating a backup schedule on " + provider)

				schedule := &api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Hour:   0,
							Minute: 30,
						},
					},
				}

				schedCreateResp, err := bc.SchedCreate(
					context.Background(),
					&api.SdkCloudBackupSchedCreateRequest{
						CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
							CredentialId: credID,
							MaxBackups:   3,
							SrcVolumeId:  "volid-doesnt-exist",
							Schedule:     schedule,
						},
					},
				)

				Expect(err).To(HaveOccurred())
				Expect(schedCreateResp).To(BeNil())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
			}
		})

		It("Should fail to create back up schedule if invalid schedule object is passed", func() {
			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Creating a backup schedule on " + provider)

				schedule := &api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Hour:   0,
							Minute: -30,
						},
					},
				}

				schedCreateResp, err := bc.SchedCreate(
					context.Background(),
					&api.SdkCloudBackupSchedCreateRequest{
						CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
							CredentialId: credID,
							MaxBackups:   3,
							SrcVolumeId:  "volid-doesnt-exist",
							Schedule:     schedule,
						},
					},
				)
				Expect(err).To(HaveOccurred())
				Expect(schedCreateResp).To(BeNil())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
			}
		})

		It("Should fail to create back up schedule if empty volume id is passed", func() {
			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Creating a backup schedule on " + provider)

				schedule := &api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Hour:   0,
							Minute: 30,
						},
					},
				}

				schedCreateResp, err := bc.SchedCreate(
					context.Background(),
					&api.SdkCloudBackupSchedCreateRequest{
						CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
							CredentialId: credID,
							MaxBackups:   3,
							SrcVolumeId:  "",
							Schedule:     schedule,
						},
					},
				)

				Expect(err).To(HaveOccurred())
				Expect(schedCreateResp).To(BeNil())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
			}
		})
	})

	Describe("Cloudbackup Schedule Enumerate", func() {

		It("Should enumerate cloud backups successfully", func() {
			By("First creating the volume")
			volID = newTestVolume(vc)

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Attaching the created volume")
				str, err := vc.Attach(
					context.Background(),
					&api.SdkVolumeAttachRequest{
						VolumeId: volID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				By("Creating a backup schedule on " + provider)

				schedule := &api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Hour:   0,
							Minute: 30,
						},
					},
				}

				schedCreateResp, err := bc.SchedCreate(
					context.Background(),
					&api.SdkCloudBackupSchedCreateRequest{
						CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
							CredentialId: credID,
							MaxBackups:   3,
							SrcVolumeId:  volID,
							Schedule:     schedule,
						},
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(schedCreateResp.BackupScheduleId).NotTo(BeNil())
			}

			enumResp, err := bc.SchedEnumerate(
				context.Background(),
				&api.SdkCloudBackupSchedEnumerateRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(enumResp.CloudSchedList).NotTo(BeEmpty())
		})
	})

	Describe("Cloudbackup Schedule Delete", func() {

		It("Should delete cloud backups successfully", func() {
			By("First creating the volume")
			volID = newTestVolume(vc)

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Attaching the created volume")
				str, err := vc.Attach(
					context.Background(),
					&api.SdkVolumeAttachRequest{
						VolumeId: volID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				By("Creating a backup schedule on " + provider)

				schedule := &api.SdkSchedulePolicyInterval{
					Retain: 1,
					PeriodType: &api.SdkSchedulePolicyInterval_Daily{
						Daily: &api.SdkSchedulePolicyIntervalDaily{
							Hour:   0,
							Minute: 30,
						},
					},
				}

				schedCreateResp, err := bc.SchedCreate(
					context.Background(),
					&api.SdkCloudBackupSchedCreateRequest{
						CloudSchedInfo: &api.SdkCloudBackupScheduleInfo{
							CredentialId: credID,
							MaxBackups:   3,
							SrcVolumeId:  volID,
							Schedule:     schedule,
						},
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(schedCreateResp.BackupScheduleId).NotTo(BeNil())

				_, err = bc.SchedDelete(
					context.Background(),
					&api.SdkCloudBackupSchedDeleteRequest{
						BackupScheduleId: schedCreateResp.BackupScheduleId,
					},
				)

				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
