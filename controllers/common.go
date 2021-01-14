package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureFinalizerExists idempotently adds a finalizer to resource metadata
func EnsureFinalizerExists(meta *metav1.ObjectMeta, finalizer string) {
	if findStringInSlice(meta.GetFinalizers(), finalizer) < 0 {
		meta.SetFinalizers(append(meta.Finalizers, finalizer))
	}
}

// EnsureFinalizerRemoved removes the given finalizer from resource metadata
func EnsureFinalizerRemoved(meta *metav1.ObjectMeta, finalizer string) {
	meta.SetFinalizers(removeStringFromSlice(meta.Finalizers, finalizer))
}

// Utility stuff
func findStringInSlice(slice []string, value string) int {
	for idx, item := range slice {
		if item == value {
			return idx
		}
	}
	return -1
}

// remove the first instance of an item from the slice, by value
func removeStringFromSlice(slice []string, value string) []string {
	idx := findStringInSlice(slice, value)
	if idx >= 0 {
		return append(slice[:idx], slice[idx+1:]...)
	}
	return slice
}
