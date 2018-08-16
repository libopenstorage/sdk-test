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
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"context"

	api "github.com/libopenstorage/openstorage-sdk-clients/sdk/golang"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func numberOfSchedulePoliciesInCluster(c api.OpenStorageSchedulePolicyClient) int {

	resp, err := c.Enumerate(
		context.Background(),
		&api.SdkSchedulePolicyEnumerateRequest{},
	)

	Expect(err).NotTo(HaveOccurred())
	Expect(resp).NotTo(BeNil())

	return len(resp.Policies)
}

var _ = Describe("SchedulePolicy [OpenStorageSchedulePolicy]", func() {
	var (
		c  api.OpenStorageSchedulePolicyClient
		ic api.OpenStorageIdentityClient
	)

	BeforeEach(func() {
		c = api.NewOpenStorageSchedulePolicyClient(conn)
		ic = api.NewOpenStorageIdentityClient(conn)

		isSupported := isCapabilitySupported(
			ic,
			api.SdkServiceCapability_OpenStorageService_SCHEDULE_POLICY,
		)

		if !isSupported {
			Skip("Schedule Policy capability not supported , skipping related tests")
		}
	})

	Describe("Create", func() {

		var (
			policyName                string
			policiesBefore            int
			policiesAfter             int
			isPolicyCreatedInTestCase bool
		)

		BeforeEach(func() {
			policiesBefore = numberOfSchedulePoliciesInCluster(c)
			policyName = ""
			isPolicyCreatedInTestCase = false
		})

		AfterEach(func() {

			// Delete the policy after the test.
			// Delete only if it was created.
			// Controlled by the isPolicyCreatedInTestCase boolean
			if isPolicyCreatedInTestCase {

				resp, err := c.Delete(
					context.Background(),
					&api.SdkSchedulePolicyDeleteRequest{
						Name: policyName,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
			}
		})

		It("Should create schedule policy", func() {
			policyName = "create-test-policy"
			resp, err := c.Create(
				context.Background(),
				&api.SdkSchedulePolicyCreateRequest{
					SchedulePolicy: &api.SdkSchedulePolicy{
						Name: policyName,
						Schedules: []*api.SdkSchedulePolicyInterval{
							&api.SdkSchedulePolicyInterval{
								Retain: 2,
								PeriodType: &api.SdkSchedulePolicyInterval_Daily{
									Daily: &api.SdkSchedulePolicyIntervalDaily{
										Hour:   12,
										Minute: 30,
									},
								},
							},
						},
					},
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())

			// Test if the policy got created.
			policiesAfter = numberOfSchedulePoliciesInCluster(c)
			Expect(policiesAfter).To(BeEquivalentTo(policiesBefore + 1))
			isPolicyCreatedInTestCase = true
		})

		It("Should fail to create policy for empty name", func() {
			resp, err := c.Create(
				context.Background(),
				&api.SdkSchedulePolicyCreateRequest{
					SchedulePolicy: &api.SdkSchedulePolicy{
						Name: policyName,
						Schedules: []*api.SdkSchedulePolicyInterval{
							&api.SdkSchedulePolicyInterval{
								Retain: 2,
								PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
									Weekly: &api.SdkSchedulePolicyIntervalWeekly{
										Day:    api.SdkTimeWeekday_SdkTimeWeekdaySunday,
										Hour:   12,
										Minute: 30,
									},
								},
							},
						},
					},
				})
			Expect(err).To(HaveOccurred())
			Expect(resp).To(BeNil())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})

		It("Should fail to create policy if retention less than 0", func() {
			policyName = "test-policy-retention"
			resp, err := c.Create(
				context.Background(),
				&api.SdkSchedulePolicyCreateRequest{
					SchedulePolicy: &api.SdkSchedulePolicy{
						Name: policyName,
						Schedules: []*api.SdkSchedulePolicyInterval{
							&api.SdkSchedulePolicyInterval{
								Retain: 0,
								PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
									Weekly: &api.SdkSchedulePolicyIntervalWeekly{
										Day:    api.SdkTimeWeekday_SdkTimeWeekdaySunday,
										Hour:   12,
										Minute: 30,
									},
								},
							},
						},
					},
				},
			)
			Expect(err).To(HaveOccurred())
			Expect(resp).To(BeNil())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})

		It("Should fail to create policy for nil object", func() {
			resp, err := c.Create(
				context.Background(),
				&api.SdkSchedulePolicyCreateRequest{
					SchedulePolicy: &api.SdkSchedulePolicy{
						Name:      policyName,
						Schedules: nil,
					},
				},
			)
			Expect(err).To(HaveOccurred())
			Expect(resp).To(BeNil())
			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Delete", func() {
		var (
			policyName     string
			policiesBefore int
			policiesAfter  int
		)

		BeforeEach(func() {
			policiesBefore = numberOfSchedulePoliciesInCluster(c)
			policyName = ""
		})

		It("Should delete a schedule policy", func() {

			By("First create a policy")
			policyName = "delete-test-policy"
			resp, err := c.Create(
				context.Background(),
				&api.SdkSchedulePolicyCreateRequest{
					SchedulePolicy: &api.SdkSchedulePolicy{
						Name: policyName,
						Schedules: []*api.SdkSchedulePolicyInterval{
							&api.SdkSchedulePolicyInterval{
								Retain: 4,
								PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
									Weekly: &api.SdkSchedulePolicyIntervalWeekly{
										Day:    api.SdkTimeWeekday_SdkTimeWeekdaySunday,
										Hour:   12,
										Minute: 30,
									},
								},
							},
						},
					},
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())

			// Test if the policy got created.
			policiesAfter = numberOfSchedulePoliciesInCluster(c)
			Expect(policiesAfter).To(BeEquivalentTo(policiesBefore + 1))

			deleteResponse, err := c.Delete(
				context.Background(),
				&api.SdkSchedulePolicyDeleteRequest{
					Name: policyName,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleteResponse).NotTo(BeNil())
		})

		// TODO: Fake driver to throw an error when deleting
		// a non-existent policy

		// It("Should fail to delete a non-existing schedule policy", func() {

		// 	policyName = "policy-doesnt-exist"

		// 	resp, err := c.Delete(
		// 		context.Background(),
		// 		&api.SdkSchedulePolicyDeleteRequest{
		// 			Name: policyName,
		// 		},
		// 	)
		// 	Expect(err).To(HaveOccurred())
		// 	Expect(resp).To(BeNil())

		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		It("Should fail to delete a schedule policy with empty name", func() {

			resp, err := c.Delete(
				context.Background(),
				&api.SdkSchedulePolicyDeleteRequest{
					Name: policyName,
				},
			)
			Expect(err).To(HaveOccurred())
			Expect(resp).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Inspect", func() {

		var (
			policyName                string
			policiesBefore            int
			policiesAfter             int
			isPolicyCreatedInTestCase bool
		)

		BeforeEach(func() {
			policiesBefore = numberOfSchedulePoliciesInCluster(c)
			policyName = ""
			isPolicyCreatedInTestCase = false
		})

		AfterEach(func() {

			// Delete the policy after the test.
			// Delete only if it was created.
			// Controlled by the isPolicyCreatedInTestCase boolean
			if isPolicyCreatedInTestCase {

				resp, err := c.Delete(
					context.Background(),
					&api.SdkSchedulePolicyDeleteRequest{
						Name: policyName,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
			}
		})

		It("Should inspect a schedule policy", func() {

			policyName = "inspect-test-policy"
			policy := &api.SdkSchedulePolicyCreateRequest{
				SchedulePolicy: &api.SdkSchedulePolicy{
					Name: policyName,
					Schedules: []*api.SdkSchedulePolicyInterval{
						&api.SdkSchedulePolicyInterval{
							Retain: 2,
							PeriodType: &api.SdkSchedulePolicyInterval_Daily{
								Daily: &api.SdkSchedulePolicyIntervalDaily{
									Hour:   12,
									Minute: 30,
								},
							},
						},
					},
				},
			}

			By("First create a policy")
			resp, err := c.Create(
				context.Background(),
				policy,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())

			// Test if the policy got created.
			policiesAfter = numberOfSchedulePoliciesInCluster(c)
			Expect(policiesAfter).To(BeEquivalentTo(policiesBefore + 1))
			isPolicyCreatedInTestCase = true

			By("Inspecting the created policy")

			inspectResponse, err := c.Inspect(
				context.Background(),
				&api.SdkSchedulePolicyInspectRequest{
					Name: policyName,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResponse).NotTo(BeNil())
			Expect(inspectResponse.Policy.Name).To(BeEquivalentTo(policy.SchedulePolicy.Name))
			Expect(inspectResponse.GetPolicy().GetSchedules()).To(HaveLen(len(policy.GetSchedulePolicy().GetSchedules())))
			Expect(inspectResponse.Policy.GetSchedules()[0].Retain).To(BeEquivalentTo(policy.SchedulePolicy.GetSchedules()[0].Retain))

			Expect(inspectResponse.Policy.GetSchedules()[0].GetDaily().Hour).To(BeEquivalentTo(policy.SchedulePolicy.GetSchedules()[0].GetDaily().Hour))
			Expect(inspectResponse.Policy.GetSchedules()[0].GetDaily().Minute).To(BeEquivalentTo(policy.SchedulePolicy.GetSchedules()[0].GetDaily().Minute))
		})

		// TODO: Fake Driver to throw error when inspecting
		// a non-existent shcedule policy

		// It("Should fail to inspect a non-existent schedule policy", func() {

		// 	policyName = "policy-doesnt-exist"

		// 	resp, err := c.Delete(
		// 		context.Background(),
		// 		&api.SdkSchedulePolicyDeleteRequest{
		// 			Name: policyName,
		// 		},
		// 	)
		// 	Expect(err).To(HaveOccurred())
		// 	Expect(resp).To(BeNil())

		// 	serverError, ok := status.FromError(err)
		// 	Expect(ok).To(BeTrue())
		// 	Expect(serverError.Code()).To(BeEquivalentTo(codes.Internal))
		// })

		It("Should fail to inspect a policy of empty name", func() {

			resp, err := c.Delete(
				context.Background(),
				&api.SdkSchedulePolicyDeleteRequest{
					Name: policyName,
				},
			)
			Expect(err).To(HaveOccurred())
			Expect(resp).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})

	Describe("Enumerate", func() {

		var (
			policyNames               []string
			policiesBefore            int
			policiesAfter             int
			isPolicyCreatedInTestCase bool
			count                     int
		)

		BeforeEach(func() {
			policiesBefore = numberOfSchedulePoliciesInCluster(c)
			count = 5
			isPolicyCreatedInTestCase = false
			policyNames = []string{}
		})

		AfterEach(func() {
			// Delete the policy after the test.
			// Delete only if it was created.
			// Controlled by the isPolicyCreatedInTestCase boolean
			if isPolicyCreatedInTestCase {

				for _, policyName := range policyNames {

					resp, err := c.Delete(
						context.Background(),
						&api.SdkSchedulePolicyDeleteRequest{
							Name: policyName,
						},
					)
					Expect(err).NotTo(HaveOccurred())
					Expect(resp).NotTo(BeNil())
				}
			}
		})

		It("Should successfully enumerate all schedule policies in the cluster", func() {

			By("Creating 5 schedule policies")
			for i := 0; i < count; i++ {
				policyName := "test-policy" + strconv.Itoa(i)
				resp, err := c.Create(
					context.Background(),
					&api.SdkSchedulePolicyCreateRequest{
						SchedulePolicy: &api.SdkSchedulePolicy{
							Name: policyName,
							Schedules: []*api.SdkSchedulePolicyInterval{
								&api.SdkSchedulePolicyInterval{
									Retain: 2,
									PeriodType: &api.SdkSchedulePolicyInterval_Daily{
										Daily: &api.SdkSchedulePolicyIntervalDaily{
											Hour:   12,
											Minute: 30,
										},
									},
								},
							},
						},
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				policyNames = append(policyNames, policyName)
			}

			// Test if the policies got created.
			policiesAfter = numberOfSchedulePoliciesInCluster(c)
			Expect(policiesAfter).To(BeEquivalentTo(policiesBefore + count))
			isPolicyCreatedInTestCase = true

			By("Enumerating all the policies in the cluster")

			enumerateResponse, err := c.Enumerate(
				context.Background(),
				&api.SdkSchedulePolicyEnumerateRequest{},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(enumerateResponse.Policies)).To(BeEquivalentTo(policiesAfter))
		})

		It("Should not fail but return zero if there are no policies in the cluster", func() {

			if policiesBefore == 0 {
				resp, err := c.Enumerate(
					context.Background(),
					&api.SdkSchedulePolicyEnumerateRequest{},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(len(resp.Policies)).To(BeEquivalentTo(policiesBefore))
			}
		})
	})

	Describe("Update", func() {

		var (
			policyName                string
			policiesBefore            int
			policiesAfter             int
			isPolicyCreatedInTestCase bool
		)

		BeforeEach(func() {
			policiesBefore = numberOfSchedulePoliciesInCluster(c)
			policyName = ""
			isPolicyCreatedInTestCase = false
		})

		AfterEach(func() {

			// Delete the policy after the test.
			// Delete only if it was created.
			// Controlled by the isPolicyCreatedInTestCase boolean
			if isPolicyCreatedInTestCase {

				resp, err := c.Delete(
					context.Background(),
					&api.SdkSchedulePolicyDeleteRequest{
						Name: policyName,
					},
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
			}
		})

		It("Should update an existing schedule policy successfully", func() {

			By("Creating a schedule policy first")
			policyName = "update-test-policy"
			resp, err := c.Create(
				context.Background(),
				&api.SdkSchedulePolicyCreateRequest{
					SchedulePolicy: &api.SdkSchedulePolicy{
						Name: policyName,
						Schedules: []*api.SdkSchedulePolicyInterval{
							&api.SdkSchedulePolicyInterval{
								Retain: 2,
								PeriodType: &api.SdkSchedulePolicyInterval_Daily{
									Daily: &api.SdkSchedulePolicyIntervalDaily{
										Hour:   11,
										Minute: 45,
									},
								},
							},
						},
					},
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())

			// Test if the policy got created.
			policiesAfter = numberOfSchedulePoliciesInCluster(c)
			Expect(policiesAfter).To(BeEquivalentTo(policiesBefore + 1))
			isPolicyCreatedInTestCase = true

			By("Updating the schedule policy")

			update := &api.SdkSchedulePolicyUpdateRequest{
				SchedulePolicy: &api.SdkSchedulePolicy{
					Name: policyName,
					Schedules: []*api.SdkSchedulePolicyInterval{
						&api.SdkSchedulePolicyInterval{
							Retain: 5,
							PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
								Weekly: &api.SdkSchedulePolicyIntervalWeekly{
									Day:    api.SdkTimeWeekday_SdkTimeWeekdayFriday,
									Hour:   12,
									Minute: 30,
								},
							},
						},
					},
				},
			}
			updateResponse, err := c.Update(
				context.Background(),
				update,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(updateResponse).NotTo(BeNil())

			By("Inspecting the updated schedule policy")
			inspectResponse, err := c.Inspect(
				context.Background(),
				&api.SdkSchedulePolicyInspectRequest{
					Name: policyName,
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(inspectResponse).NotTo(BeNil())
			Expect(inspectResponse.Policy.Name).To(BeEquivalentTo(update.SchedulePolicy.Name))
			Expect(inspectResponse.GetPolicy().GetSchedules()).To(HaveLen(len(update.GetSchedulePolicy().GetSchedules())))
			Expect(inspectResponse.Policy.Schedules[0].Retain).
				To(BeEquivalentTo(update.SchedulePolicy.Schedules[0].Retain))
			Expect(inspectResponse.Policy.Schedules[0].GetWeekly().Day).
				To(BeEquivalentTo(update.SchedulePolicy.Schedules[0].GetWeekly().Day))

			// TODO: Fake driver should update all the fields that are requested
			// in the update struct

			// Expect(inspectResponse.Policy.Schedule.GetWeekly().Hour).
			// 	To(BeEquivalentTo(update.SchedulePolicy.Schedule.GetWeekly().Hour))
			// Expect(inspectResponse.Policy.Schedule.GetWeekly().Minute).
			// 	To(BeEquivalentTo(update.SchedulePolicy.Schedule.GetWeekly().Minute))
		})

		It("Should fail to update the name of the schedule policy", func() {

			By("Creating a schedule policy first")
			policyName = "fail-test-policy-update"
			resp, err := c.Create(
				context.Background(),
				&api.SdkSchedulePolicyCreateRequest{
					SchedulePolicy: &api.SdkSchedulePolicy{
						Name: policyName,
						Schedules: []*api.SdkSchedulePolicyInterval{
							&api.SdkSchedulePolicyInterval{
								Retain: 2,
								PeriodType: &api.SdkSchedulePolicyInterval_Daily{
									Daily: &api.SdkSchedulePolicyIntervalDaily{
										Hour:   11,
										Minute: 45,
									},
								},
							},
						},
					},
				},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())

			// Test if the policy got created.
			policiesAfter = numberOfSchedulePoliciesInCluster(c)
			Expect(policiesAfter).To(BeEquivalentTo(policiesBefore + 1))
			isPolicyCreatedInTestCase = true

			By("Updating the schedule policy")

			update := &api.SdkSchedulePolicyUpdateRequest{
				SchedulePolicy: &api.SdkSchedulePolicy{
					Name: "policy-name-changed",
					Schedules: []*api.SdkSchedulePolicyInterval{
						&api.SdkSchedulePolicyInterval{
							Retain: 5,
							PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
								Weekly: &api.SdkSchedulePolicyIntervalWeekly{
									Day:    api.SdkTimeWeekday_SdkTimeWeekdayFriday,
									Hour:   12,
									Minute: 30,
								},
							},
						},
					},
				},
			}
			updateResponse, err := c.Update(
				context.Background(),
				update,
			)

			Expect(err).To(HaveOccurred())
			Expect(updateResponse).To(BeNil())

			// since policy name can't be found this will fail with not
			// found error
			_, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
		})

		It("Should fail to update a empty schedule policy name", func() {
			update := &api.SdkSchedulePolicyUpdateRequest{
				SchedulePolicy: &api.SdkSchedulePolicy{
					Name: policyName,
					Schedules: []*api.SdkSchedulePolicyInterval{
						&api.SdkSchedulePolicyInterval{
							Retain: 5,
							PeriodType: &api.SdkSchedulePolicyInterval_Weekly{
								Weekly: &api.SdkSchedulePolicyIntervalWeekly{
									Day:    api.SdkTimeWeekday_SdkTimeWeekdayFriday,
									Hour:   12,
									Minute: 30,
								},
							},
						},
					},
				},
			}
			updateResponse, err := c.Update(
				context.Background(),
				update,
			)
			Expect(err).To(HaveOccurred())
			Expect(updateResponse).To(BeNil())

			serverError, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(serverError.Code()).To(BeEquivalentTo(codes.InvalidArgument))
		})
	})
})
