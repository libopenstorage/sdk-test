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

var _ = Describe("Credentials [OpenStorageCredentials]", func() {
	var (
		credClient api.OpenStorageCredentialsClient
		ic         api.OpenStorageIdentityClient
	)

	BeforeEach(func() {

		ic = api.NewOpenStorageIdentityClient(conn)

		isSupported := isCapabilitySupported(
			ic,
			api.SdkServiceCapability_OpenStorageService_CREDENTIALS,
		)

		if !isSupported {
			Skip("Credentials capability not supported , skipping related tests")
		}

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
			credClient.Delete(context.Background(), &api.SdkCredentialDeleteRequest{CredentialId: credID})
		}
	})

	Describe("Credentials Create [Buggy]", func() {
		It("Should Create Credentials", func() {
			credEnumReq := &api.SdkCredentialEnumerateRequest{}
			credEnumResp, err := credClient.Enumerate(context.Background(), credEnumReq)

			By("checking what is there already")
			Expect(err).NotTo(HaveOccurred())
			numCreds := len(credEnumResp.GetCredentialIds())

			By("creating new credentials")
			numCredCreated := parseAndCreateCredentials(credClient)
			credEnumResp, err = credClient.Enumerate(context.Background(), credEnumReq)

			By("checking the new credentials were created")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(credEnumResp.GetCredentialIds())).To(Equal(numCredCreated + numCreds))

		})

		It("Should provide detail,verify, and delete given credential ID", func() {
			credID := ""
			accessKey := ""
			region := ""

			// Create credential from cb.yaml
			for provider, providerParams := range config.ProviderConfig.CloudProviders {
				if provider == "aws" {
					credReq := &api.SdkCredentialCreateRequest{
						Name: providerParams["CredName"],
						CredentialType: &api.SdkCredentialCreateRequest_AwsCredential{
							AwsCredential: &api.SdkAwsCredentialRequest{
								AccessKey:  providerParams["CredAccessKey"],
								SecretKey:  providerParams["CredSecretKey"],
								Endpoint:   providerParams["CredEndpoint"],
								Region:     providerParams["CredRegion"],
								DisableSsl: providerParams["CredDisableSSL"] == "true",
							},
						},
					}

					By("creating credentials")
					credResp, err := credClient.Create(context.Background(), credReq)
					Expect(err).NotTo(HaveOccurred())
					credID = credResp.GetCredentialId()
					Expect(credID).NotTo(BeEmpty())
					accessKey = credReq.GetAwsCredential().GetAccessKey()
					region = credReq.GetAwsCredential().GetRegion()

					By("verfiying credentials")
					_, err = credClient.Validate(context.Background(), &api.SdkCredentialValidateRequest{CredentialId: credID})
					Expect(err).NotTo(HaveOccurred())

					By("inspecting credentials")
					inspectReq := &api.SdkCredentialInspectRequest{CredentialId: credID}
					inspectResp, err := credClient.Inspect(context.Background(), inspectReq)
					Expect(err).NotTo(HaveOccurred())
					Expect(inspectResp.GetAwsCredential().GetAccessKey()).To(BeEquivalentTo(accessKey))
					Expect(inspectResp.GetAwsCredential().GetRegion()).To(BeEquivalentTo(region))

					By("deleting credentials")
					_, err = credClient.Delete(context.Background(), &api.SdkCredentialDeleteRequest{CredentialId: credID})
					Expect(err).NotTo(HaveOccurred())
					break
				}
			}
		})
	})

	Describe("Credentials Delete", func() {
		It("Should failed to delete non-existanant Credentials", func() {

			_, err := credClient.Delete(context.Background(), &api.SdkCredentialDeleteRequest{CredentialId: ""})
			Expect(err).To(HaveOccurred())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

})
