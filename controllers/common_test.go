package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestFinalizerLogic ensures that the add and remove finalizer
// functions work as expected
func TestFinalizerLogic(t *testing.T) {
	g := NewGomegaWithT(t)
	meta := metav1.ObjectMeta{}

	g.Expect(len(meta.Finalizers)).To(Equal(0))

	EnsureFinalizerExists(&meta, "foo")
	g.Expect(len(meta.Finalizers)).To(Equal(1))
	EnsureFinalizerExists(&meta, "bar")
	g.Expect(len(meta.Finalizers)).To(Equal(2))

	// Ad shold be idempotent
	EnsureFinalizerExists(&meta, "foo")
	g.Expect(len(meta.Finalizers)).To(Equal(2))

	EnsureFinalizerRemoved(&meta, "foo")
	g.Expect(len(meta.Finalizers)).To(Equal(1))
	EnsureFinalizerRemoved(&meta, "bar")
	g.Expect(len(meta.Finalizers)).To(Equal(0))
}
