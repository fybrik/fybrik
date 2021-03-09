// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"time"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Blueprint Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Millisecond * 100

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Blueprint", func() {
		It("Should create successfully", func() {
			// Load
			var err error
			blueprintYAML, err := ioutil.ReadFile("../../testdata/blueprint.yaml")
			Expect(err).ToNot(HaveOccurred())
			blueprint := &app.Blueprint{}
			err = yaml.Unmarshal(blueprintYAML, blueprint)
			Expect(err).ToNot(HaveOccurred())

			key, err := client.ObjectKeyFromObject(blueprint)
			Expect(err).ToNot(HaveOccurred())

			// Create Blueprint
			Expect(k8sClient.Create(context.Background(), blueprint)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				f := &app.Blueprint{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name}}
				_ = k8sClient.Get(context.Background(), key, f)
				_ = k8sClient.Delete(context.Background(), f)
			}()

			// Check number of releases - should be two
			By("Expecting to reconcile successfully with copy and read module releases")
			Eventually(func() int {
				f := &app.Blueprint{}
				if err := k8sClient.Get(context.Background(), key, f); err != nil {
					return 0
				}
				if f.Status.Releases == nil {
					return 0
				}
				return len(f.Status.Releases)
			}, timeout, interval).Should(Equal(2))

			// Update
			By("Expecting to update successfully")
			Eventually(func() error {
				f := &app.Blueprint{}
				if err := k8sClient.Get(context.Background(), key, f); err != nil {
					return err
				}
				// remove copy module (the first one in the flow)
				f.Spec.Flow.Steps = f.Spec.Flow.Steps[1:]
				f.Spec.Templates = f.Spec.Templates[1:]
				return k8sClient.Update(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			// Check number of releases - should be only one
			By("Expecting to reconcile successfully with a single read-module release")
			var releaseNames []string
			Eventually(func() int {
				f := &app.Blueprint{}
				releaseNames = []string{}
				if err := k8sClient.Get(context.Background(), key, f); err != nil {
					return 0
				}
				for release, version := range f.Status.Releases {
					Expect(version).To(Equal(f.Status.ObservedGeneration))
					releaseNames = append(releaseNames, release)
				}
				return len(releaseNames)
			}, timeout, interval).Should(Equal(1))

			Expect(releaseNames[0]).Should(ContainSubstring("read"))
			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &app.Blueprint{}
				if err := k8sClient.Get(context.Background(), key, f); err != nil {
					return err
				}
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting delete to finish")
			Eventually(func() error {
				f := &app.Blueprint{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})

	Context("Release name", func() {
		It("Should form a name when name is short", func() {
			blueprint := app.Blueprint{
				ObjectMeta: metav1.ObjectMeta{
					Name: "appns-app-mybp",
				},
				Spec: app.BlueprintSpec{
					Flow: app.DataFlow{
						Name: "dataflow",
						Steps: []app.FlowStep{{Name: "mystep",
							Template:  "template",
							Arguments: app.ModuleArguments{}}},
					},
				},
			}

			relName := getReleaseName(blueprint.Name, blueprint.Spec.Flow.Steps[0])
			Expect(relName).To(Equal("appns-app-mybp-mystep"))
		})

		It("Should limit the release name if it is too long", func() {
			blueprint := app.Blueprint{
				ObjectMeta: metav1.ObjectMeta{
					Name: "appnsisalreadylong-appnameisevenlonger-myblueprintnameisreallytakingitoverthetopkubernetescantevendealwithit",
				},
				Spec: app.BlueprintSpec{
					Flow: app.DataFlow{
						Name: "dataflow",
						Steps: []app.FlowStep{{Name: "ohandnottoforgettheflowstepnamethatincludesthetemplatenameandotherstuff",
							Template:  "template",
							Arguments: app.ModuleArguments{}}},
					},
				},
			}

			relName := getReleaseName(blueprint.Name, blueprint.Spec.Flow.Steps[0])
			Expect(relName).To(Equal("appnsisalreadylong-appnameisevenlonger-mybluepr-3c184"))
			Expect(relName).To(HaveLen(53))

			// Make sure that calling the same method again results in the same result
			relName2 := getReleaseName(blueprint.Name, blueprint.Spec.Flow.Steps[0])
			Expect(relName2).To(Equal("appnsisalreadylong-appnameisevenlonger-mybluepr-3c184"))
			Expect(relName2).To(HaveLen(53))
		})
	})
})
