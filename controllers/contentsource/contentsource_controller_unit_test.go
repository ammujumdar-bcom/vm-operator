// Copyright (c) 2019-2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package contentsource_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"

	"github.com/vmware-tanzu/vm-operator/controllers/contentsource"
	providerfake "github.com/vmware-tanzu/vm-operator/pkg/vmprovider/fake"
	"github.com/vmware-tanzu/vm-operator/test/builder"
)

func unitTests() {
	Describe("Invoking VirtualMachineImage CRUD unit tests", unitTestsCRUDImage)
	Describe("Invoking ReconcileProviderRef unit tests", reconcileProviderRef)
}

func reconcileProviderRef() {
	var (
		ctx            *builder.UnitTestContextForController
		reconciler     *contentsource.ContentSourceReconciler
		fakeVmProvider *providerfake.FakeVmProvider
		initObjects    []runtime.Object

		cs v1alpha1.ContentSource
		cl v1alpha1.ContentLibraryProvider
	)

	cl = v1alpha1.ContentLibraryProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy-cl",
		},
		Spec: v1alpha1.ContentLibraryProviderSpec{
			UUID: "dummy-cl-uuid",
		},
	}
	cs = v1alpha1.ContentSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dummy-cs",
		},
		Spec: v1alpha1.ContentSourceSpec{
			ProviderRef: v1alpha1.ContentProviderReference{
				Name:      cl.Name,
				Namespace: cl.Namespace,
				Kind:      "ContentLibraryProvider",
			},
		},
	}

	JustBeforeEach(func() {
		ctx = suite.NewUnitTestContextForController(initObjects...)
		reconciler = contentsource.NewReconciler(
			ctx.Client,
			ctx.Logger,
			ctx.Recorder,
			ctx.VmProvider,
		)
		fakeVmProvider = ctx.VmProvider.(*providerfake.FakeVmProvider)
	})

	AfterEach(func() {
		ctx.AfterEach()
		ctx = nil
		initObjects = nil
		reconciler = nil
		fakeVmProvider.Reset()
		fakeVmProvider = nil
	})

	Context("ReconcileProviderRef", func() {
		When("the ContentLibraryProvider does not exist", func() {
			BeforeEach(func() {
				initObjects = append(initObjects, &cs)
			})
			It("returns an error", func() {
				err := reconciler.ReconcileProviderRef(ctx, &cs)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})
		When("error in checking if the content library exists on vSphere", func() {
			BeforeEach(func() {
				initObjects = []runtime.Object{&cs, &cl}
			})

			It("fails with an error", func() {
				expectedError := fmt.Errorf("error in checking if a content library exists on vSphere")
				fakeVmProvider.Lock()
				fakeVmProvider.DoesContentLibraryExistFn = func(ctx context.Context, cl *v1alpha1.ContentLibraryProvider) (bool, error) {
					return false, expectedError
				}
				fakeVmProvider.Unlock()

				err := reconciler.ReconcileProviderRef(ctx, &cs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(expectedError))
			})
		})

		When("the content library does not exist on vSphere", func() {
			BeforeEach(func() {
				initObjects = []runtime.Object{&cs, &cl}
			})

			It("fails with an error", func() {
				fakeVmProvider.Lock()
				fakeVmProvider.DoesContentLibraryExistFn = func(ctx context.Context, cl *v1alpha1.ContentLibraryProvider) (bool, error) {
					return false, nil
				}
				fakeVmProvider.Unlock()

				err := reconciler.ReconcileProviderRef(ctx, &cs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(fmt.Errorf("Content library does not exist on provider. contentLibraryUUID: %s", cl.Spec.UUID)))
			})
		})

		Context("with a ContentLibraryProvider pointing to a vSphere content library", func() {
			BeforeEach(func() {
				initObjects = []runtime.Object{&cs, &cl}
			})

			It("updates the ContentLibraryProvider to add the OwnerRef", func() {
				err := reconciler.ReconcileProviderRef(ctx, &cs)
				Expect(err).NotTo(HaveOccurred())

				clAfterReconcile := &v1alpha1.ContentLibraryProvider{}
				clKey := client.ObjectKey{Name: cl.ObjectMeta.Name}
				err = ctx.Client.Get(ctx, clKey, clAfterReconcile)
				Expect(err).NotTo(HaveOccurred())
				Expect(clAfterReconcile.OwnerReferences[0].Name).To(Equal(cs.Name))
			})
		})

	})
}
func unitTestsCRUDImage() {
	var (
		ctx            *builder.UnitTestContextForController
		reconciler     *contentsource.ContentSourceReconciler
		fakeVmProvider *providerfake.FakeVmProvider
		initObjects    []runtime.Object

		cs v1alpha1.ContentSource
		cl v1alpha1.ContentLibraryProvider
	)

	cl = v1alpha1.ContentLibraryProvider{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy-cl",
			Namespace: "dummy-ns",
		},
	}
	cs = v1alpha1.ContentSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy-cs",
			Namespace: "dummy-ns",
		},
		Spec: v1alpha1.ContentSourceSpec{
			ProviderRef: v1alpha1.ContentProviderReference{
				Name:      cl.Name,
				Namespace: cl.Namespace,
			},
		},
	}

	JustBeforeEach(func() {
		ctx = suite.NewUnitTestContextForController(initObjects...)
		reconciler = contentsource.NewReconciler(
			ctx.Client,
			ctx.Logger,
			ctx.Recorder,
			ctx.VmProvider,
		)
		fakeVmProvider = ctx.VmProvider.(*providerfake.FakeVmProvider)
	})

	AfterEach(func() {
		ctx.AfterEach()
		ctx = nil
		initObjects = nil
		reconciler = nil
		fakeVmProvider.Reset()
		fakeVmProvider = nil
	})

	Context("DeleteImages", func() {
		var (
			images []v1alpha1.VirtualMachineImage
			image  v1alpha1.VirtualMachineImage
		)

		image = v1alpha1.VirtualMachineImage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy-vm-image",
				Namespace: "dummy-ns",
			},
		}

		When("no images are specified", func() {
			It("does not throw an error", func() {
				err := reconciler.DeleteImages(ctx, images)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("non-empty list of images is specified", func() {
			BeforeEach(func() {
				initObjects = append(initObjects, &image)
				images = append(images, image)
			})

			It("successfully deletes the images", func() {
				err := reconciler.DeleteImages(ctx, images)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("when client delete fails because the image doesnt exist", func() {
			BeforeEach(func() {
				images = append(images, image)
			})

			It("returns an error", func() {
				err := reconciler.DeleteImages(ctx, images)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("UpdateImages", func() {
		var (
			images []v1alpha1.VirtualMachineImage
			image  v1alpha1.VirtualMachineImage
		)

		image = v1alpha1.VirtualMachineImage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy-vm-image",
				Namespace: "dummy-ns",
			},
		}

		When("no images are specified", func() {
			It("does not throw an error", func() {
				err := reconciler.UpdateImages(ctx, images)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("non-empty list of images is specified", func() {
			BeforeEach(func() {
				initObjects = append(initObjects, &image)
				images = append(images, image)
			})

			It("successfully updates the images", func() {
				// Modify the VirtualMachineImage spec
				images[0].Spec.Type = "updated-dummy-type"
				imgBeforeUpdate := images[0]

				err := reconciler.UpdateImages(ctx, images)
				Expect(err).NotTo(HaveOccurred())

				imgAfterUpdate := &v1alpha1.VirtualMachineImage{}
				objKey := client.ObjectKey{Name: images[0].Name, Namespace: images[0].Namespace}
				Expect(ctx.Client.Get(ctx, objKey, imgAfterUpdate)).To(Succeed())

				Expect(imgBeforeUpdate.Spec.Type).To(Equal(imgAfterUpdate.Spec.Type))
			})
		})

		When("when client update fails", func() {
			BeforeEach(func() {
				initObjects = append(initObjects, &image)
				images = append(images, image)
			})

			It("fails to update the images", func() {
				images[0].Name = "invalid_name" // invalid namespace, to fail the Update op.

				err := reconciler.UpdateImages(ctx, images)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("CreateImages", func() {
		var (
			images []v1alpha1.VirtualMachineImage
			image  v1alpha1.VirtualMachineImage
		)

		image = v1alpha1.VirtualMachineImage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dummy-vm-image",
				Namespace: "dummy-ns",
			},
		}

		When("non-empty list of images is specified", func() {
			BeforeEach(func() {
				images = append(images, image)
			})

			It("successfully creates the images", func() {
				err := reconciler.CreateImages(ctx, images)
				Expect(err).NotTo(HaveOccurred())

				img := images[0]
				Expect(ctx.Client.Get(ctx, client.ObjectKey{Name: img.Name, Namespace: img.Namespace}, &img)).To(Succeed())
			})
		})

		When("when client create fails because the image already exist", func() {
			BeforeEach(func() {
				initObjects = append(initObjects, &image)
				images = append(images, image)
			})

			It("fails to create the images", func() {
				err := reconciler.CreateImages(ctx, images)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsAlreadyExists(err)).To(BeTrue())
			})
		})
	})

	Context("DiffImages: Difference VirtualMachineImage resources", func() {
		Context("when left and right is empty", func() {
			var left []v1alpha1.VirtualMachineImage
			var right []v1alpha1.VirtualMachineImage

			It("return empty sets", func() {
				added, removed, updated := reconciler.DiffImages(left, right)
				Expect(added).To(BeEmpty())
				Expect(removed).To(BeEmpty())
				Expect(updated).To(BeEmpty())
			})
		})

		Context("when left is empty and right is non-empty", func() {
			var left []v1alpha1.VirtualMachineImage
			var right []v1alpha1.VirtualMachineImage
			var image v1alpha1.VirtualMachineImage

			BeforeEach(func() {
				image = v1alpha1.VirtualMachineImage{}
				right = append(right, image)
			})

			It("return a non-empty added set", func() {
				added, removed, updated := reconciler.DiffImages(left, right)
				Expect(added).ToNot(BeEmpty())
				Expect(added).To(HaveLen(1))
				Expect(removed).To(BeEmpty())
				Expect(updated).To(BeEmpty())
			})
		})

		Context("when left is non-empty and right is empty", func() {
			var left []v1alpha1.VirtualMachineImage
			var right []v1alpha1.VirtualMachineImage
			var image v1alpha1.VirtualMachineImage

			BeforeEach(func() {
				image = v1alpha1.VirtualMachineImage{}
				left = append(left, image)
			})

			It("return a non-empty removed set", func() {
				added, removed, updated := reconciler.DiffImages(left, right)
				Expect(added).To(BeEmpty())
				Expect(removed).ToNot(BeEmpty())
				Expect(removed).To(HaveLen(1))
				Expect(updated).To(BeEmpty())
			})
		})

		Context("when left and right are non-empty and the same", func() {
			var left []v1alpha1.VirtualMachineImage
			var right []v1alpha1.VirtualMachineImage
			var imageL v1alpha1.VirtualMachineImage
			var imageR v1alpha1.VirtualMachineImage

			BeforeEach(func() {
				imageL = v1alpha1.VirtualMachineImage{}
				imageR = v1alpha1.VirtualMachineImage{}
			})

			JustBeforeEach(func() {
				left = append(left, imageL)
				right = append(right, imageR)
			})

			Context("when left and right have a different spec", func() {
				BeforeEach(func() {
					imageL = v1alpha1.VirtualMachineImage{
						Spec: v1alpha1.VirtualMachineImageSpec{
							Type: "left-type",
						},
					}

					imageR = v1alpha1.VirtualMachineImage{
						Spec: v1alpha1.VirtualMachineImageSpec{
							Type: "right-type",
						},
					}
				})

				It("should return a non-empty updated spec", func() {
					added, removed, updated := reconciler.DiffImages(left, right)
					Expect(added).To(BeEmpty())
					Expect(removed).To(BeEmpty())
					Expect(updated).ToNot(BeEmpty())
					Expect(updated).To(HaveLen(1))
				})
			})

			Context("when left and right have samespec", func() {
				It("should return an empty updated spec", func() {
					added, removed, updated := reconciler.DiffImages(left, right)
					Expect(added).To(BeEmpty())
					Expect(removed).To(BeEmpty())
					Expect(updated).ToNot(BeEmpty())
					Expect(updated).To(HaveLen(1))
				})
			})

		})

		Context("when left and right are non-empty and unique", func() {
			var left []v1alpha1.VirtualMachineImage
			var right []v1alpha1.VirtualMachineImage
			var imageLeft v1alpha1.VirtualMachineImage
			var imageRight v1alpha1.VirtualMachineImage

			BeforeEach(func() {
				imageLeft = v1alpha1.VirtualMachineImage{
					ObjectMeta: metav1.ObjectMeta{
						Name: "left",
					},
				}
				imageRight = v1alpha1.VirtualMachineImage{
					ObjectMeta: metav1.ObjectMeta{
						Name: "right",
					},
				}
				left = append(left, imageLeft)
				right = append(right, imageRight)
			})

			It("return a non-empty added and removed set", func() {
				added, removed, updated := reconciler.DiffImages(left, right)
				Expect(added).ToNot(BeEmpty())
				Expect(added).To(HaveLen(1))
				Expect(removed).ToNot(BeEmpty())
				Expect(removed).To(HaveLen(1))
				Expect(updated).To(BeEmpty())
			})
		})

		Context("when left and right are non-empty and have a non-complete intersection", func() {
			var left []v1alpha1.VirtualMachineImage
			var right []v1alpha1.VirtualMachineImage
			var imageLeft v1alpha1.VirtualMachineImage
			var imageRight v1alpha1.VirtualMachineImage

			BeforeEach(func() {
				imageLeft = v1alpha1.VirtualMachineImage{
					ObjectMeta: metav1.ObjectMeta{
						Name: "left",
					},
				}
				imageRight = v1alpha1.VirtualMachineImage{
					ObjectMeta: metav1.ObjectMeta{
						Name: "right",
					},
				}
				left = append(left, imageLeft)
				right = append(right, imageLeft)
				right = append(right, imageRight)
			})

			It("return a non-empty added set with a single entry", func() {
				added, removed, updated := reconciler.DiffImages(left, right)
				Expect(added).ToNot(BeEmpty())
				Expect(added).To(HaveLen(1))
				Expect(added).To(ContainElement(imageRight))
				Expect(removed).To(BeEmpty())
				Expect(updated).To(BeEmpty())
			})
		})
	})

	Context("GetImagesFromContentProvider", func() {
		var (
			cs v1alpha1.ContentSource
			cl v1alpha1.ContentLibraryProvider
		)

		BeforeEach(func() {

		})

		Context("when the ContentLibraryProvider resource doesnt exist", func() {
			It("returns error", func() {
				images, err := reconciler.GetImagesFromContentProvider(ctx.Context, cs)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
				Expect(images).To(BeNil())
			})
		})

		When("provider returns error in listing images from CL", func() {
			BeforeEach(func() {
				initObjects = append(initObjects, &cs, &cl)
			})

			It("provider returns error when listing images", func() {
				fakeVmProvider.ListVirtualMachineImagesFromContentLibraryFn = func(ctx context.Context, _ v1alpha1.ContentLibraryProvider) ([]*v1alpha1.VirtualMachineImage, error) {
					return nil, fmt.Errorf("error listing images from provider")
				}

				images, err := reconciler.GetImagesFromContentProvider(ctx.Context, cs)
				Expect(err).To(HaveOccurred())
				Expect(images).To(BeNil())
			})
		})

		Context("when ContentSource resource passes to a valid vSphere CL", func() {
			var images []*v1alpha1.VirtualMachineImage

			BeforeEach(func() {
				initObjects = append(initObjects, &cs, &cl)

				images = []*v1alpha1.VirtualMachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "dummy-image-1",
							Namespace: "dummy-ns",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "dummy-image-2",
							Namespace: "dummy-ns",
						},
					},
				}
			})

			It("provider successfully lists images", func() {
				fakeVmProvider.ListVirtualMachineImagesFromContentLibraryFn = func(ctx context.Context, _ v1alpha1.ContentLibraryProvider) ([]*v1alpha1.VirtualMachineImage, error) {
					return images, nil
				}

				clImages, err := reconciler.GetImagesFromContentProvider(ctx.Context, cs)
				Expect(err).NotTo(HaveOccurred())
				Expect(clImages).Should(HaveLen(2))
				Expect(clImages).Should(Equal(images))

			})
		})
	})

	Context("DifferenceImages", func() {

		var (
			img1 *v1alpha1.VirtualMachineImage
			img2 *v1alpha1.VirtualMachineImage
		)

		BeforeEach(func() {
			img1 = &v1alpha1.VirtualMachineImage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy-image-1",
				},
			}
			img2 = &v1alpha1.VirtualMachineImage{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dummy-image-2",
				},
			}
		})

		imageExists := func(imageName string, images []v1alpha1.VirtualMachineImage) bool {
			for _, img := range images {
				if imageName == img.Name {
					return true
				}
			}

			return false
		}

		When("no ContentSources exist", func() {
			It("returns no error and no added/removed/updated images", func() {
				err, added, removed, updated := reconciler.DifferenceImages(ctx.Context)
				Expect(err).NotTo(HaveOccurred())

				Expect(added).To(BeEmpty())
				Expect(removed).To(BeEmpty())
				Expect(updated).To(BeEmpty())
			})
		})

		When("Images exist on the API server and provider", func() {
			BeforeEach(func() {
				initObjects = append(initObjects, img1, &cl, &cs)
			})

			It("Should remove the image from APIServer and add image from provider", func() {
				fakeVmProvider.ListVirtualMachineImagesFromContentLibraryFn = func(ctx context.Context, cl v1alpha1.ContentLibraryProvider) ([]*v1alpha1.VirtualMachineImage, error) {
					return []*v1alpha1.VirtualMachineImage{img2}, nil
				}

				err, added, removed, updated := reconciler.DifferenceImages(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(added).NotTo(BeEmpty())
				Expect(added).To(HaveLen(1))
				Expect(imageExists(img2.Name, added)).To(BeTrue())

				Expect(removed).NotTo(BeEmpty())
				Expect(removed).To(HaveLen(1))
				Expect(imageExists(img1.Name, removed)).To(BeTrue())

				Expect(updated).To(BeEmpty())
			})
		})
	})
}
