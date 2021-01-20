package controllers

import (
	"math/rand"
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

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
