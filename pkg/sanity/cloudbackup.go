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
	"time"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getBackupId(bc api.OpenStorageCloudBackupClient, clusterId, volumeId, credentialId string) string {

	enumerateReq := &api.SdkCloudBackupEnumerateRequest{
		ClusterId:    clusterId,
		SrcVolumeId:  volumeId,
		CredentialId: credentialId,
	}
	enumerateResp, err := bc.Enumerate(context.Background(), enumerateReq)
	Expect(err).NotTo(HaveOccurred())
	Expect(enumerateResp).NotTo(BeNil())
	Expect(enumerateResp.GetBackups()).NotTo(BeEmpty())

	var backupId string
	for _, backup := range enumerateResp.GetBackups() {
		if backup.GetSrcVolumeId() == volumeId {
			backupId = backup.GetId()
			break
		}
	}
	Expect(backupId).NotTo(BeEmpty())

	return backupId
}

var _ = Describe("Cloud backup [OpenStorageCluster]", func() {
	var (
		cc api.OpenStorageCredentialsClient
		vc api.OpenStorageVolumeClient
		bc api.OpenStorageCloudBackupClient
		c  api.OpenStorageClusterClient
		nc api.OpenStorageNodeClient
		ic api.OpenStorageIdentityClient

		bkpStatusReq *api.SdkCloudBackupStatusRequest
		bkpStatus    *api.SdkCloudBackupStatus
		volID        string
		credID       string
		credsUUIDMap map[string]string
		clusterID    string
		nodeID       string
	)

	BeforeEach(func() {

		cc = api.NewOpenStorageCredentialsClient(conn)
		bc = api.NewOpenStorageCloudBackupClient(conn)
		vc = api.NewOpenStorageVolumeClient(conn)
		c = api.NewOpenStorageClusterClient(conn)
		nc = api.NewOpenStorageNodeClient(conn)
		ic = api.NewOpenStorageIdentityClient(conn)

		isSupported := isCapabilitySupported(
			ic,
			api.SdkServiceCapability_OpenStorageService_CLOUD_BACKUP,
		)

		if !isSupported {
			Skip("Cloud Backup capability not supported , skipping related tests")
		}

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

				if volID != "" {

					_, err := bc.DeleteAll(
						context.Background(),
						&api.SdkCloudBackupDeleteAllRequest{
							SrcVolumeId:  volID,
							CredentialId: credID,
						},
					)
					Expect(err).NotTo(HaveOccurred())
				}

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

	Describe("Backup Create", func() {

		It("Should create a cloud backup successfully", func() {
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

				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}

				_, err = bc.Create(context.Background(), backupReq)
				Expect(err).NotTo(HaveOccurred())

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.SdkCloudBackupStatusRequest{
						VolumeId: volID,
					}
					bkpStatusResp, err := bc.Status(context.Background(), bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volID]

					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone {
						break
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone))
			}
		})

		It("Should fail to create back up if non-existent volume id is passed", func() {
			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid
				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     "this-doesnt-exist",
					CredentialId: credID,
					Full:         false,
				}

				_, err := bc.Create(context.Background(), backupReq)
				Expect(err).To(HaveOccurred())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))

			}
		})

		It("Should fail to create back up if empty volume id is passed", func() {
			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}
				_, err := bc.Create(context.Background(), backupReq)
				Expect(err).To(HaveOccurred())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
			}
		})

		It("Should fail to create back up if non-existent credentials is passed", func() {

			By("First creating the volume")
			volID = newTestVolume(vc)

			By("Creating backup with invalid credentials")

			backupReq := &api.SdkCloudBackupCreateRequest{
				VolumeId:     volID,
				CredentialId: "cred-uuid-doesnt-exist",
				Full:         false,
			}
			_, err := bc.Create(context.Background(), backupReq)
			Expect(err).To(HaveOccurred())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		})

		It("Should fail to create back up if empty credentials is passed", func() {

			By("First creating the volume")
			volID = newTestVolume(vc)

			By("Creating backup with invalid credentials")

			backupReq := &api.SdkCloudBackupCreateRequest{
				VolumeId:     volID,
				CredentialId: "",
				Full:         false,
			}
			_, err := bc.Create(context.Background(), backupReq)
			Expect(err).To(HaveOccurred())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Backup Enumerate", func() {

		It("Should Enumerate cloud backups successfully", func() {

			By("First creating the volume")
			volID = newTestVolume(vc)

			By("Getting the cluster id of the cluster")
			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

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

				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}

				_, err = bc.Create(context.Background(), backupReq)
				Expect(err).NotTo(HaveOccurred())

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.SdkCloudBackupStatusRequest{
						VolumeId: volID,
					}
					bkpStatusResp, err := bc.Status(context.Background(), bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volID]

					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone {
						break
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone))

				By("Enumerating the cloud backups")

				enumerateResp, err := bc.Enumerate(
					context.Background(),
					&api.SdkCloudBackupEnumerateRequest{
						All:          true,
						ClusterId:    clusterID,
						CredentialId: credID,
						SrcVolumeId:  volID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(enumerateResp.Backups).NotTo(BeEmpty())
			}
		})

		It("Should fail to list cloud backups for non-existing volume", func() {

			By("Getting the cluster id of the cluster")
			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("Doing Backup on " + provider)

				enumerateReq := &api.SdkCloudBackupEnumerateRequest{
					ClusterId:    clusterID,
					SrcVolumeId:  "this-doesnt-exist",
					CredentialId: credID,
				}

				enumerateResp, err := bc.Enumerate(context.Background(), enumerateReq)
				Expect(err).NotTo(HaveOccurred())
				Expect(enumerateResp).ToNot(BeNil())
				Expect(enumerateResp.GetBackups()).To(HaveLen(0))
			}
		})

		//TODO  Fake driver enumerate all the backs if the volume id is emmpty.

		// It("Should fail to list cloud backups for empty volume id", func() {

		// 	By("Getting the cluster id of the cluster")
		// 	inpectResp, err := c.InspectCurrent(
		// 		context.Background(),
		// 		&api.SdkClusterInspectCurrentRequest{},
		// 	)

		// 	Expect(err).NotTo(HaveOccurred())
		// 	clusterID = inpectResp.Cluster.Id

		// 	credsUUIDMap = parseAndCreateCredentials2(cc)
		// 	for provider, uuid := range credsUUIDMap {
		// 		credID = uuid

		// 		By("Doing Backup on " + provider)

		// 		enumerateReq := &api.SdkCloudBackupEnumerateRequest{
		// 			ClusterId:    clusterID,
		// 			SrcVolumeId:  "",
		// 			CredentialId: credID,
		// 		}

		// 		enumerateResp, err := bc.Enumerate(context.Background(), enumerateReq)
		// 		Expect(err).To(HaveOccurred())
		// 		Expect(enumerateResp).To(BeNil())

		// 		serverError, ok := status.FromError(err)
		// 		Expect(ok).To(BeTrue())
		// 		Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		// 	}
		// })

		It("Should fail to enumerate back up if non-existent credentials is passed", func() {
			By("Getting the cluster id of the cluster")
			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)
			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

			By("Creating the volume")
			volID = newTestVolume(vc)

			enumerateReq := &api.SdkCloudBackupEnumerateRequest{
				ClusterId:    clusterID,
				SrcVolumeId:  volID,
				CredentialId: "dummy-credentials",
			}

			enumerateResp, err := bc.Enumerate(context.Background(), enumerateReq)
			Expect(err).To(HaveOccurred())
			Expect(enumerateResp).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		})

		It("Should fail to enumerate back up if empty credentials is passed", func() {
			By("Getting the cluster id of the cluster")
			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

			By("Creating the volume")
			volID = newTestVolume(vc)

			By("Doing cloudbakup enumerate")
			enumerateReq := &api.SdkCloudBackupEnumerateRequest{
				ClusterId:    clusterID,
				SrcVolumeId:  volID,
				CredentialId: "",
			}

			enumerateResp, err := bc.Enumerate(context.Background(), enumerateReq)
			Expect(err).To(HaveOccurred())
			Expect(enumerateResp).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Backup Catalog", func() {

		It("Should successfully list out the catalog", func() {

			By("Getting the cluster id of the cluster")

			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

			By("First creating the volume")
			volID = newTestVolume(vc)

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for provider, uuid := range credsUUIDMap {
				credID = uuid

				By("attaching the created volume")
				str, err := vc.Attach(
					context.Background(),
					&api.SdkVolumeAttachRequest{
						VolumeId: volID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(str).NotTo(BeNil())

				By("doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}

				_, err = bc.Create(context.Background(), backupReq)
				Expect(err).NotTo(HaveOccurred())

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.SdkCloudBackupStatusRequest{
						VolumeId: volID,
					}
					bkpStatusResp, err := bc.Status(context.Background(), bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volID]

					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone {
						break
					} else if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed {
						break
					} else {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone))

				By("checking the catalog")
				catalogResp, err := bc.Catalog(
					context.Background(),
					&api.SdkCloudBackupCatalogRequest{
						BackupId:     getBackupId(bc, clusterID, volID, credID),
						CredentialId: credID,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(catalogResp.Contents).NotTo(BeEmpty())
			}
		})

		It("Should fail list out the catalog of non existent backup id", func() {

			By("Getting the cluster id of the cluster")

			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for _, uuid := range credsUUIDMap {
				credID = uuid

				catalogResp, err := bc.Catalog(
					context.Background(),
					&api.SdkCloudBackupCatalogRequest{
						BackupId:     "dummy-backupid",
						CredentialId: credID,
					},
				)
				Expect(err).To(HaveOccurred())
				Expect(catalogResp).To(BeNil())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
			}
		})
		It("Should fail list out the catalog of empty backup id", func() {

			By("Getting the cluster id of the cluster")

			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for _, uuid := range credsUUIDMap {
				credID = uuid

				catalogResp, err := bc.Catalog(
					context.Background(),
					&api.SdkCloudBackupCatalogRequest{
						BackupId:     "",
						CredentialId: credID,
					},
				)
				Expect(err).To(HaveOccurred())
				Expect(catalogResp).To(BeNil())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
			}
		})
	})

	Describe("Cloudbackup History", func() {
		It("Should successfully list out the history", func() {

			By("Getting the cluster id of the cluster")

			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

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

				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}

				_, err = bc.Create(context.Background(), backupReq)
				Expect(err).NotTo(HaveOccurred())

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.SdkCloudBackupStatusRequest{
						VolumeId: volID,
					}
					bkpStatusResp, err := bc.Status(context.Background(), bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volID]

					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone {
						break
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone))
			}

			By("Getting cloud backup history of the created volume")
			historyResp, err := bc.History(
				context.Background(),
				&api.SdkCloudBackupHistoryRequest{
					SrcVolumeId: volID,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(historyResp.GetHistoryList()).NotTo(BeEmpty())
		})

		// TODO Fake driver to return error code 13  but instead returns 3
		// It("Should successfully fail to get cloud backup history of non-existent volume", func() {

		// 	By("Getting cloud backup history of the created volume")
		// 	_, err := bc.History(
		// 		context.Background(),
		// 		&api.SdkCloudBackupHistoryRequest{
		// 			SrcVolumeId: volID,
		// 		},
		// 	)
		// 	Expect(err).To(HaveOccurred())
		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		It("Should successfully fail to get cloud backup history of empty volume id", func() {

			By("Getting cloud backup history of the created volume")
			_, err := bc.History(
				context.Background(),
				&api.SdkCloudBackupHistoryRequest{
					SrcVolumeId: volID,
				},
			)
			Expect(err).To(HaveOccurred())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Cloub backup Restore", func() {
		It("Should successfully restore the cloud backup", func() {

			By("Getting the cluster id of the cluster")

			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

			By("Getting the node id")

			nodeResp, err := nc.InspectCurrent(
				context.Background(),
				&api.SdkNodeInspectCurrentRequest{},
			)
			Expect(err).NotTo(HaveOccurred())
			nodeID = nodeResp.Node.Id

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

				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}

				_, err = bc.Create(context.Background(), backupReq)
				Expect(err).NotTo(HaveOccurred())

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.SdkCloudBackupStatusRequest{
						VolumeId: volID,
					}
					bkpStatusResp, err := bc.Status(context.Background(), bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volID]

					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone {
						break
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone))

				By("Doing restore of the cloud backup")
				restoreResp, err := bc.Restore(
					context.Background(),
					&api.SdkCloudBackupRestoreRequest{
						BackupId:          getBackupId(bc, clusterID, volID, credID),
						CredentialId:      credID,
						NodeId:            nodeID,
						RestoreVolumeName: "restored-volume-" + volID,
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(restoreResp).NotTo(BeNil())

				By("Inspecting the restored volume")

				inspectResp, err := vc.Inspect(
					context.Background(),
					&api.SdkVolumeInspectRequest{
						VolumeId: restoreResp.RestoreVolumeId,
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(inspectResp.Volume).NotTo(BeNil())
			}
		})
	})

	Describe("Cloud backup Delete", func() {

		It("Should successfully delete the cloud backup", func() {

			By("Getting the cluster id of the cluster")

			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

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

				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}

				_, err = bc.Create(context.Background(), backupReq)
				Expect(err).NotTo(HaveOccurred())

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.SdkCloudBackupStatusRequest{
						VolumeId: volID,
					}
					bkpStatusResp, err := bc.Status(context.Background(), bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volID]

					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone {
						break
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone))

				_, err = bc.Delete(
					context.Background(),
					&api.SdkCloudBackupDeleteRequest{
						BackupId:     getBackupId(bc, clusterID, volID, credID),
						CredentialId: credID,
					},
				)

				Expect(err).NotTo(HaveOccurred())

			}
		})

		// TODO : Fake driver to return error if the cloud backup id doesn't exist

		// It("Should fail to delete the cloud backup for non existent cloud backup id", func() {

		// 	By("Creating all the credentials provided in the cloud provider config file.")

		// 	credsUUIDMap = parseAndCreateCredentials2(cc)
		// 	for _, uuid := range credsUUIDMap {
		// 		credID = uuid

		// 		_, err := bc.Delete(
		// 			context.Background(),
		// 			&api.SdkCloudBackupDeleteRequest{
		// 				BackupId:     "doesnt-exist",
		// 				CredentialId: credID,
		// 			},
		// 		)
		// 		Expect(err).To(HaveOccurred())

		// 		serverError, ok := status.FromError(err)
		// 		Expect(ok).To(BeTrue())
		// 		Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))

		// 	}
		// })

		It("Should fail to delete the cloud backup for empty cloud backup id", func() {

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for _, uuid := range credsUUIDMap {
				credID = uuid

				_, err := bc.Delete(
					context.Background(),
					&api.SdkCloudBackupDeleteRequest{
						BackupId:     "",
						CredentialId: credID,
					},
				)
				Expect(err).To(HaveOccurred())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))

			}
		})
	})

	Describe("Cloud backup Deleteall", func() {

		It("Should successfully delete all the cloud backup", func() {

			By("Getting the cluster id of the cluster")

			inpectResp, err := c.InspectCurrent(
				context.Background(),
				&api.SdkClusterInspectCurrentRequest{},
			)

			Expect(err).NotTo(HaveOccurred())
			clusterID = inpectResp.Cluster.Id

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

				By("Doing Backup on " + provider)

				backupReq := &api.SdkCloudBackupCreateRequest{
					VolumeId:     volID,
					CredentialId: credID,
					Full:         false,
				}

				_, err = bc.Create(context.Background(), backupReq)
				Expect(err).NotTo(HaveOccurred())

				// timeout after 5 mins
				timeout := 300
				timespent := 0
				for timespent < timeout {
					bkpStatusReq = &api.SdkCloudBackupStatusRequest{
						VolumeId: volID,
					}
					bkpStatusResp, err := bc.Status(context.Background(), bkpStatusReq)
					Expect(err).To(BeNil())

					bkpStatus = bkpStatusResp.Statuses[volID]

					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone {
						break
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive {
						time.Sleep(time.Second * 10)
						timeout += 10
					}
					if bkpStatus.Status == api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed {
						break
					}
				}
				Expect(bkpStatus.Status).To(BeEquivalentTo(api.SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone))

				_, err = bc.DeleteAll(
					context.Background(),
					&api.SdkCloudBackupDeleteAllRequest{
						SrcVolumeId:  volID,
						CredentialId: credID,
					},
				)

				Expect(err).NotTo(HaveOccurred())

			}
		})

		// TODO : Fake driver to return error if the cloud backup id doesn't exist

		// It("Should fail to delete the cloud backup for non existent volume id", func() {

		// 	By("Creating all the credentials provided in the cloud provider config file.")

		// 	credsUUIDMap = parseAndCreateCredentials2(cc)
		// 	for _, uuid := range credsUUIDMap {
		// 		credID = uuid

		// 		_, err := bc.DeleteAll(
		// 			context.Background(),
		// 			&api.SdkCloudBackupDeleteAllRequest{
		// 				SrcVolumeId:     "doesnt-exist",
		// 				CredentialId: credID,
		// 			},
		// 		)
		// 		Expect(err).To(HaveOccurred())

		// 		serverError, ok := status.FromError(err)
		// 		Expect(ok).To(BeTrue())
		// 		Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))

		// 	}
		// })

		It("Should fail to delete the cloud backup for empty volume id", func() {

			By("Creating all the credentials provided in the cloud provider config file.")

			credsUUIDMap = parseAndCreateCredentials2(cc)
			for _, uuid := range credsUUIDMap {
				credID = uuid

				_, err := bc.DeleteAll(
					context.Background(),
					&api.SdkCloudBackupDeleteAllRequest{
						SrcVolumeId:  "",
						CredentialId: credID,
					},
				)
				Expect(err).To(HaveOccurred())

				serverError, ok := status.FromError(err)
				Expect(ok).To(BeTrue())
				Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
			}
		})
	})

})
