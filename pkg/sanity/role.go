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

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Role Service Test Suite", func() {

	var (
		rc    api.OpenStorageRoleClient
		users map[string]string
	)

	BeforeEach(func() {
		if len(config.SharedSecret) == 0 {
			Skip("Not running with authentication")
		}
		rc = api.NewOpenStorageRoleClient(conn)
		users = createUsersTokens()
	})

	Describe("Role APIs", func() {
		It("should deny without permission", func() {
			ctx := setContextWithToken(context.Background(), users["user1"])
			_, err := rc.Create(ctx, &api.SdkRoleCreateRequest{})
			Expect(err).To(HaveOccurred())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.PermissionDenied))
		})

		It("should create, list, get, and delete a role", func() {
			role := "tester"

			By("creating role")
			ctx := setContextWithToken(context.Background(), users["admin"])
			r, err := rc.Create(ctx, &api.SdkRoleCreateRequest{
				Role: &api.SdkRole{
					Name: role,
					Rules: []*api.SdkRule{
						&api.SdkRule{
							Services: []string{"identity"},
							Apis:     []string{"*"},
						},
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(r).NotTo(BeNil())
			Expect(r.GetRole()).NotTo(BeNil())
			Expect(r.GetRole().GetName()).To(BeEquivalentTo(role))

			By("getting role")
			i, err := rc.Inspect(ctx, &api.SdkRoleInspectRequest{
				Name: role,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(i).NotTo(BeNil())
			Expect(i.GetRole()).NotTo(BeNil())
			Expect(i.GetRole().GetName()).To(BeEquivalentTo(role))

			By("listing roles")
			roles, err := rc.Enumerate(ctx, &api.SdkRoleEnumerateRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(roles).NotTo(BeNil())
			Expect(roles.GetNames()).To(ContainElement(role))

			By("deleting roles")
			_, err = rc.Delete(ctx, &api.SdkRoleDeleteRequest{
				Name: role,
			})
			Expect(err).ToNot(HaveOccurred())

			By("checking role was deleted")
			roles, err = rc.Enumerate(ctx, &api.SdkRoleEnumerateRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(roles).NotTo(BeNil())
			Expect(roles.GetNames()).ToNot(ContainElement(role))

		})
	})
})
