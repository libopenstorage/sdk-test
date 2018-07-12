package sanity

import (
	"context"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credentials [OpenStorageCredentials]", func() {
	var (
		credClient api.OpenStorageCredentialsClient
	)

	BeforeEach(func() {

		credClient = api.NewOpenStorageCredentialsClient(conn)
		if config.ProviderConfig == nil {
			Skip("Skipping credentials tests")
		}
	})

	AfterEach(func() {
		// Delete all credential stored after each test
		credEnumReq := &api.SdkCredentialEnumerateRequest{}
		credEnumResp, err := credClient.Enumerate(context.Background(), credEnumReq)
		Expect(err).NotTo(HaveOccurred())

		for _, credID := range credEnumResp.GetCredentialIds() {
			_, err = credClient.Delete(context.Background(), &api.SdkCredentialDeleteRequest{CredentialId: credID})
			Expect(err).NotTo(HaveOccurred())
		}

	})

	Describe("Credentials Create", func() {
		It("Should Create Credentials", func() {

			numCredCreated := parseAndCreateCredentials(credClient)
			credEnumReq := &api.SdkCredentialEnumerateRequest{}
			credEnumResp, err := credClient.Enumerate(context.Background(), credEnumReq)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(credEnumResp.GetCredentialIds())).To(Equal(numCredCreated))

		})

		It("Should provide detail of given credential ID", func() {
			credID := ""
			accessKey := ""
			region := ""

			// Create credential from cb.yaml
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
					credID = credResp.GetCredentialId()
					accessKey = credReq.GetAwsCredential().GetAccessKey()
					region = credReq.GetAwsCredential().GetRegion()

					break
				}
			}

			inspectReq := &api.SdkCredentialInspectRequest{CredentialId: credID}
			inspectResp, err := credClient.Inspect(context.Background(), inspectReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResp.GetAwsCredential().GetAccessKey()).To(BeEquivalentTo(accessKey))
			Expect(inspectResp.GetAwsCredential().GetRegion()).To(BeEquivalentTo(region))

		})
	})

	Describe("Credentials Validate", func() {

		It("Should validate created Credentials", func() {

			numCredCreated := parseAndCreateCredentials(credClient)
			credEnumReq := &api.SdkCredentialEnumerateRequest{}
			credEnumResp, err := credClient.Enumerate(context.Background(), credEnumReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(credEnumResp.GetCredentialIds())).To(Equal(numCredCreated))
			for _, credID := range credEnumResp.GetCredentialIds() {
				_, err = credClient.Validate(context.Background(), &api.SdkCredentialValidateRequest{CredentialId: credID})
				Expect(err).NotTo(HaveOccurred())
			}

		})
	})

	Describe("Credentials Delete", func() {
		It("Should delete created Credentials", func() {

			numCredCreated := parseAndCreateCredentials(credClient)
			credEnumReq := &api.SdkCredentialEnumerateRequest{}
			credEnumResp, err := credClient.Enumerate(context.Background(), credEnumReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(credEnumResp.GetCredentialIds())).To(Equal(numCredCreated))
			for _, credID := range credEnumResp.GetCredentialIds() {
				_, err = credClient.Delete(context.Background(), &api.SdkCredentialDeleteRequest{CredentialId: credID})
				Expect(err).NotTo(HaveOccurred())
			}
			credEnumReq = &api.SdkCredentialEnumerateRequest{}
			credEnumResp, err = credClient.Enumerate(context.Background(), credEnumReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(credEnumResp.GetCredentialIds()).To(BeNil())

		})

		It("Should failed to delete non-existanant Credentials", func() {

			_, err := credClient.Delete(context.Background(), &api.SdkCredentialDeleteRequest{CredentialId: ""})
			Expect(err).To(HaveOccurred())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

})
