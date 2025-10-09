package testutil

import (
	"context"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

// FakeDynamicClient implements a fake Kubernetes dynamic client for testing
type FakeDynamicClient struct {
	resources map[string]map[string]*unstructured.Unstructured // namespace -> name -> resource
	mu        sync.RWMutex
}

// NewFakeDynamicClient creates a new fake dynamic client
func NewFakeDynamicClient() *FakeDynamicClient {
	return &FakeDynamicClient{
		resources: make(map[string]map[string]*unstructured.Unstructured),
	}
}

// Resource returns a namespace-able resource interface
func (f *FakeDynamicClient) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return &fakeNamespaceableResource{
		client: f,
		gvr:    resource,
	}
}

type fakeNamespaceableResource struct {
	client    *FakeDynamicClient
	gvr       schema.GroupVersionResource
	namespace string
}

func (f *fakeNamespaceableResource) Namespace(ns string) dynamic.ResourceInterface {
	return &fakeNamespaceableResource{
		client:    f.client,
		gvr:       f.gvr,
		namespace: ns,
	}
}

func (f *fakeNamespaceableResource) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	f.client.mu.Lock()
	defer f.client.mu.Unlock()

	if _, ok := f.client.resources[f.namespace]; !ok {
		f.client.resources[f.namespace] = make(map[string]*unstructured.Unstructured)
	}

	name := obj.GetName()
	if _, exists := f.client.resources[f.namespace][name]; exists {
		return nil, fmt.Errorf("resource %s already exists", name)
	}

	// Clone the object
	created := obj.DeepCopy()
	f.client.resources[f.namespace][name] = created
	return created, nil
}

func (f *fakeNamespaceableResource) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	f.client.mu.Lock()
	defer f.client.mu.Unlock()

	name := obj.GetName()
	if _, ok := f.client.resources[f.namespace]; !ok {
		return nil, fmt.Errorf("resource %s not found", name)
	}
	if _, exists := f.client.resources[f.namespace][name]; !exists {
		return nil, fmt.Errorf("resource %s not found", name)
	}

	updated := obj.DeepCopy()
	f.client.resources[f.namespace][name] = updated
	return updated, nil
}

func (f *fakeNamespaceableResource) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return f.Update(ctx, obj, metav1.UpdateOptions{}, "status")
}

func (f *fakeNamespaceableResource) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	f.client.mu.Lock()
	defer f.client.mu.Unlock()

	if _, ok := f.client.resources[f.namespace]; !ok {
		return fmt.Errorf("resource %s not found", name)
	}
	if _, exists := f.client.resources[f.namespace][name]; !exists {
		return fmt.Errorf("resource %s not found", name)
	}

	delete(f.client.resources[f.namespace], name)
	return nil
}

func (f *fakeNamespaceableResource) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	f.client.mu.Lock()
	defer f.client.mu.Unlock()

	delete(f.client.resources, f.namespace)
	return nil
}

func (f *fakeNamespaceableResource) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	f.client.mu.RLock()
	defer f.client.mu.RUnlock()

	if _, ok := f.client.resources[f.namespace]; !ok {
		return nil, fmt.Errorf("resource %s not found", name)
	}
	obj, exists := f.client.resources[f.namespace][name]
	if !exists {
		return nil, fmt.Errorf("resource %s not found", name)
	}

	return obj.DeepCopy(), nil
}

func (f *fakeNamespaceableResource) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	f.client.mu.RLock()
	defer f.client.mu.RUnlock()

	list := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{},
	}

	if resources, ok := f.client.resources[f.namespace]; ok {
		for _, obj := range resources {
			list.Items = append(list.Items, *obj.DeepCopy())
		}
	}

	return list, nil
}

func (f *fakeNamespaceableResource) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, fmt.Errorf("watch not implemented")
}

func (f *fakeNamespaceableResource) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("patch not implemented")
}

func (f *fakeNamespaceableResource) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("apply not implemented")
}

func (f *fakeNamespaceableResource) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return nil, fmt.Errorf("apply status not implemented")
}

// Helper functions for creating test proxy rules

// NewProxyRule creates a test proxy rule
func NewProxyRule(name, domain, destination string, port int) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "bausteln.io/v1",
			"kind":       "Proxyrule",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "proxy-rules",
			},
			"spec": map[string]interface{}{
				"domain":      domain,
				"destination": destination,
				"tls":         true,
			},
		},
	}

	if port > 0 {
		spec := obj.Object["spec"].(map[string]interface{})
		spec["port"] = int64(port)
	}

	return obj
}

// SeedProxyRule adds a proxy rule to the fake client
func (f *FakeDynamicClient) SeedProxyRule(name, namespace, domain, destination string, port int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.resources[namespace]; !ok {
		f.resources[namespace] = make(map[string]*unstructured.Unstructured)
	}

	obj := NewProxyRule(name, domain, destination, port)
	obj.SetNamespace(namespace)
	f.resources[namespace][name] = obj
}
