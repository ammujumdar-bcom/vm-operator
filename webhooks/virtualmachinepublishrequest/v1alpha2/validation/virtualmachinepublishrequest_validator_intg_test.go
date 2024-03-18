// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	vmopv1 "github.com/vmware-tanzu/vm-operator/api/v1alpha2"
	pkgconfig "github.com/vmware-tanzu/vm-operator/pkg/config"
	"github.com/vmware-tanzu/vm-operator/pkg/constants/testlabels"
	"github.com/vmware-tanzu/vm-operator/test/builder"
)

func intgTests() {
	Describe(
		"Create",
		Label(
			testlabels.Create,
			testlabels.EnvTest,
			testlabels.V1Alpha2,
			testlabels.Validation,
			testlabels.Webhook,
		),
		intgTestsValidateCreate,
	)
	Describe(
		"Update",
		Label(
			testlabels.Update,
			testlabels.EnvTest,
			testlabels.V1Alpha2,
			testlabels.Validation,
			testlabels.Webhook,
		),
		intgTestsValidateUpdate,
	)
	Describe(
		"Delete",
		Label(
			testlabels.Delete,
			testlabels.EnvTest,
			testlabels.V1Alpha2,
			testlabels.Validation,
			testlabels.Webhook,
		),
		intgTestsValidateDelete,
	)
}

type intgValidatingWebhookContext struct {
	builder.IntegrationTestContext
	vmPub *vmopv1.VirtualMachinePublishRequest
}

func newIntgValidatingWebhookContext() *intgValidatingWebhookContext {
	ctx := &intgValidatingWebhookContext{
		IntegrationTestContext: *suite.NewIntegrationTestContext(),
	}

	ctx.vmPub = builder.DummyVirtualMachinePublishRequestA2("dummy-vmpub", ctx.Namespace, "dummy-vm",
		"dummy-item", "dummy-cl")
	return ctx
}

func intgTestsValidateCreate() {
	var (
		err error
		ctx *intgValidatingWebhookContext
	)

	BeforeEach(func() {
		pkgconfig.SetContext(suite, func(config *pkgconfig.Config) {
			config.Features.ImageRegistry = true
		})
		ctx = newIntgValidatingWebhookContext()
	})

	AfterEach(func() {
		pkgconfig.SetContext(suite, func(config *pkgconfig.Config) {
			config.Features.ImageRegistry = false
		})
		err = nil
		ctx = nil
	})

	When("WCP_VM_Image_Registry is enabled: create is performed", func() {
		It("should allow the request", func() {
			Eventually(func() error {
				return ctx.Client.Create(ctx, ctx.vmPub)
			}).Should(Succeed())
		})
	})

	When("WCP_VM_Image_Registry is not enabled", func() {
		BeforeEach(func() {
			pkgconfig.SetContext(suite, func(config *pkgconfig.Config) {
				config.Features.ImageRegistry = false
			})
			err = ctx.Client.Create(ctx, ctx.vmPub)
		})
		It("should deny the request", func() {
			Eventually(func() string {
				if err = ctx.Client.Create(ctx, ctx.vmPub); err != nil {
					return err.Error()
				}
				return ""
			}).Should(ContainSubstring("WCP_VM_Image_Registry feature not enabled"))
		})
	})
}

func intgTestsValidateUpdate() {
	var (
		err error
		ctx *intgValidatingWebhookContext
	)

	BeforeEach(func() {
		ctx = newIntgValidatingWebhookContext()
		pkgconfig.SetContext(suite, func(config *pkgconfig.Config) {
			config.Features.ImageRegistry = true
		})

		Expect(ctx.Client.Create(ctx, ctx.vmPub)).To(Succeed())
	})

	JustBeforeEach(func() {
		err = ctx.Client.Update(suite, ctx.vmPub)
	})

	AfterEach(func() {
		Expect(ctx.Client.Delete(ctx, ctx.vmPub)).To(Succeed())
		pkgconfig.SetContext(suite, func(config *pkgconfig.Config) {
			config.Features.ImageRegistry = false
		})

		err = nil
		ctx = nil
	})

	When("update is performed with changed source name", func() {
		BeforeEach(func() {
			ctx.vmPub.Spec.Source.Name = "alternate-vm-name"
		})
		It("should deny the request", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	When("update is performed with changed target info", func() {
		BeforeEach(func() {
			ctx.vmPub.Spec.Target.Location.Name = "alternate-cl"
		})
		It("should deny the request", func() {
			Expect(err).To(HaveOccurred())
		})
	})
}

func intgTestsValidateDelete() {
	var (
		err error
		ctx *intgValidatingWebhookContext
	)

	BeforeEach(func() {
		ctx = newIntgValidatingWebhookContext()
		pkgconfig.SetContext(suite, func(config *pkgconfig.Config) {
			config.Features.ImageRegistry = true
		})

		Expect(ctx.Client.Create(ctx, ctx.vmPub)).To(Succeed())
	})

	JustBeforeEach(func() {
		err = ctx.Client.Delete(suite, ctx.vmPub)
		pkgconfig.SetContext(suite, func(config *pkgconfig.Config) {
			config.Features.ImageRegistry = false
		})
	})

	AfterEach(func() {
		err = nil
		ctx = nil
	})

	When("delete is performed", func() {
		It("should allow the request", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
}