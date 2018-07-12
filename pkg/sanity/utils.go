package sanity

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/gomega"
)

const (
	BYTE = 1.0 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
)

func testVolumeDetails(
	req *api.SdkVolumeCreateRequest,
	volume *api.Volume,
) {

	// check volume specs
	Expect(volume.Spec.Ephemeral).To(BeEquivalentTo(req.Spec.Ephemeral))
	Expect(volume.Spec.BlockSize).To(BeEquivalentTo(req.Spec.BlockSize))
	Expect(volume.Spec.Cascaded).To(BeEquivalentTo(req.Spec.Cascaded))
	Expect(volume.Spec.Compressed).To(BeEquivalentTo(req.Spec.Compressed))

	Expect(volume.Spec.Dedupe).To(BeEquivalentTo(req.Spec.Dedupe))

	Expect(volume.Spec.Group).To(BeEquivalentTo(req.Spec.Group))
	Expect(volume.Spec.GroupEnforced).To(BeEquivalentTo(req.Spec.GroupEnforced))

	Expect(volume.Spec.Journal).To(BeEquivalentTo(req.Spec.Journal))
	Expect(volume.Spec.Sharedv4).To(BeEquivalentTo(req.Spec.Sharedv4))
	Expect(volume.Spec.Passphrase).To(BeEquivalentTo(req.Spec.Passphrase))
	Expect(volume.Spec.ReplicaSet).To(BeEquivalentTo(req.Spec.ReplicaSet))
	Expect(volume.Spec.Scale).To(BeEquivalentTo(req.Spec.Scale))
	Expect(volume.Spec.Shared).To(BeEquivalentTo(req.Spec.Shared))
	Expect(volume.Spec.Size).To(BeEquivalentTo(req.Spec.Size))
	Expect(volume.Spec.SnapshotInterval).To(BeEquivalentTo(req.Spec.SnapshotInterval))
	Expect(volume.Spec.SnapshotSchedule).To(BeEquivalentTo(req.Spec.SnapshotSchedule))
	Expect(volume.Source.Parent).To(BeEmpty())
	Expect(volume.Locator.Name).To(BeEquivalentTo(req.Name))

	// TODO: Fake driver mmust honor the below parameters

	//Expect(volume.Spec.AggregationLevel).To(BeEquivalentTo(req.Spec.AggregationLevel))
	//Expect(volume.Spec.Cos).To(BeEquivalentTo(req.Spec.Cos))
	//Expect(volume.Spec.Encrypted).To(BeEquivalentTo(req.Spec.Encrypted))
	//Expect(volume.Spec.Format).To(BeEquivalentTo(req.Spec.Format))
	//Expect(volume.Spec.HaLevel).To(BeEquivalentTo(req.Spec.HaLevel))
	//Expect(volume.Spec.IoProfile).To(BeEquivalentTo(req.Spec.IoProfile))
	//Expect(volume.Spec.Sticky).To(BeEquivalentTo(req.Spec.Sticky))
}

func testVolumeCreation(req *api.SdkVolumeCreateRequest) {

}

func numberOfVolumesInCluster(c api.OpenStorageVolumeClient) int {
	res, err := c.Enumerate(
		context.Background(),
		&api.SdkVolumeEnumerateRequest{},
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(res).NotTo(BeNil())
	return len(res.VolumeIds)
}

// This will create credential for provider listed from cb.yaml file
func parseAndCreateCredentials(credClient api.OpenStorageCredentialsClient) int {
	numCredCreated := 0
	for provider, providerParams := range config.ProviderConfig.CloudProviders {
		if provider == "aws" {
			credReq := &api.SdkCredentialCreateRequest{
				CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
					AwsCredential: &api.SdkAwsCredentialRequest{
						AccessKey: providerParams["CredAccessKey"],
						SecretKey: providerParams["CredSecretKey"],
						Endpoint:  providerParams["CredEndpoint"],
						Region:    providerParams["CredRegion"],
					},
				},
			}

			credResp, err := credClient.Create(context.Background(), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			numCredCreated++

		} else if provider == "azure" {
			credReq := &api.SdkCredentialCreateRequest{
				CredentialType: &api.SdkCredentialCreateRequest_AzureCredential{
					AzureCredential: &api.SdkAzureCredentialRequest{
						AccountKey:  providerParams["CredAccountName"],
						AccountName: providerParams["CredAccountKey"],
					},
				},
			}

			credResp, err := credClient.Create(context.Background(), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			numCredCreated++

		} else if provider == "google" {
			credReq := &api.SdkCredentialCreateRequest{
				CredentialType: &api.SdkCredentialCreateRequest_GoogleCredential{
					GoogleCredential: &api.SdkGoogleCredentialRequest{
						ProjectId: providerParams["CredProjectID"],
						JsonKey:   providerParams["CredJsonKey"],
					},
				},
			}

			credResp, err := credClient.Create(context.Background(), credReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credResp.GetCredentialId()).NotTo(BeEmpty())
			numCredCreated++
		}
	}
	return numCredCreated
}
