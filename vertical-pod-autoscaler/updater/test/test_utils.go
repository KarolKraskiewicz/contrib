/*
Copyright 2017 The Kubernetes Authors.

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

package test

import (
	"fmt"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"
	"k8s.io/contrib/vertical-pod-autoscaler/updater/api_mock"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/testapi"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	extensions "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"
	v1 "k8s.io/kubernetes/pkg/client/listers/core/v1"
)

// BuildTestPod creates a pod with specified resources.
func BuildTestPod(name, container_name, cpu, mem string, creator runtime.Object) *apiv1.Pod {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{BuildTestContainer(container_name, cpu, mem)},
		},
	}

	if creator != nil {
		pod.ObjectMeta.Annotations = map[string]string{apiv1.CreatedByAnnotation: RefJSON(creator)}
	}

	if len(cpu) > 0 {
		cpuVal, _ := resource.ParseQuantity(cpu)
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = cpuVal
	}
	if len(mem) > 0 {
		memVal, _ := resource.ParseQuantity(mem)
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = memVal
	}

	return pod
}

func BuildTestContainer(container_name, cpu, mem string) apiv1.Container {
	container := apiv1.Container{
		Name: container_name,
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{},
		},
	}

	if len(cpu) > 0 {
		cpuVal, _ := resource.ParseQuantity(cpu)
		container.Resources.Requests[apiv1.ResourceCPU] = cpuVal
	}
	if len(mem) > 0 {
		memVal, _ := resource.ParseQuantity(mem)
		container.Resources.Requests[apiv1.ResourceMemory] = memVal
	}

	return container
}

func BuildTestPolicy(containerName, minCpu, maxCpu, minMemory, maxMemory string) *api_mock.ResourcesPolicy {
	minCpuVal, _ := resource.ParseQuantity(minCpu)
	maxCpuVal, _ := resource.ParseQuantity(maxCpu)
	minMemVal, _ := resource.ParseQuantity(minMemory)
	maxMemVal, _ := resource.ParseQuantity(maxMemory)
	return &api_mock.ResourcesPolicy{Containers: []api_mock.ContainerPolicy{{
		Name: containerName,
		MemoryPolicy: api_mock.Policy{
			Min: minMemVal,
			Max: maxMemVal},
		CpuPolicy: api_mock.Policy{
			Min: minCpuVal,
			Max: maxCpuVal},
	},
	}}
}

func BuildTestVpaObject(containerName, minCpu, maxCpu, minMemory, maxMemory string, labels map[string]string) *api_mock.VerticalPodAutoscaler {
	resourcesPolicy := BuildTestPolicy(containerName, minCpu, maxCpu, minMemory, maxMemory)

	return &api_mock.VerticalPodAutoscaler{
		Spec: api_mock.Spec{
			Target:          api_mock.Target{MatchLabels: labels},
			UpdatePolicy:    api_mock.UpdatePolicy{Mode: api_mock.Mode{}},
			ResourcesPolicy: *resourcesPolicy,
		},
	}

}

// RefJSON builds string reference to
func RefJSON(o runtime.Object) string {
	ref, err := apiv1.GetReference(api.Scheme, o)
	if err != nil {
		panic(err)
	}

	codec := testapi.Default.Codec()
	json := runtime.EncodeOrDie(codec, &apiv1.SerializedReference{Reference: *ref})
	return string(json)
}

func Recommendation(containerName, cpu, mem string) *api_mock.Recommendation {
	result := &api_mock.Recommendation{Containers: []api_mock.ContainerRecommendation{
		{Name: containerName}},
	}
	if len(cpu) > 0 {
		cpuVal, _ := resource.ParseQuantity(cpu)
		result.Containers[0].Cpu = cpuVal
	}

	if len(mem) > 0 {
		memVal, _ := resource.ParseQuantity(mem)
		result.Containers[0].Memory = memVal
	}

	return result
}

type RecommenderAPIMock struct {
	mock.Mock
}

func (m *RecommenderAPIMock) GetRecommendation(spec *apiv1.PodSpec) (*api_mock.Recommendation, error) {
	args := m.Called(spec)
	var returnArg *api_mock.Recommendation = nil
	if args.Get(0) != nil {
		returnArg = args.Get(0).(*api_mock.Recommendation)
	}
	return returnArg, args.Error(1)
}

type RecommenderMock struct {
	mock.Mock
}

func (m *RecommenderMock) Get(spec *apiv1.PodSpec) (*api_mock.Recommendation, error) {
	args := m.Called(spec)
	var returnArg *api_mock.Recommendation = nil
	if args.Get(0) != nil {
		returnArg = args.Get(0).(*api_mock.Recommendation)
	}
	return returnArg, args.Error(1)
}

type EvictionControllerMock struct {
	mock.Mock
}

func (m *EvictionControllerMock) Evict(pod *apiv1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

func (m *EvictionControllerMock) CanEvict(pod *apiv1.Pod) bool {
	args := m.Called(pod)
	return args.Bool(0)
}

type PodListerMock struct {
	mock.Mock
}

func (*PodListerMock) Pods(namespace string) v1.PodNamespaceLister {
	panic("implement me")
}

func (m *PodListerMock) List(selector labels.Selector) (ret []*apiv1.Pod, err error) {
	args := m.Called()
	var returnArg []*apiv1.Pod = nil
	if args.Get(0) != nil {
		returnArg = args.Get(0).([]*apiv1.Pod)
	}
	return returnArg, args.Error(1)
}

type VpaListerMock struct {
	mock.Mock
}

func (m *VpaListerMock) List() (ret []*api_mock.VerticalPodAutoscaler, err error) {
	args := m.Called()
	var returnArg []*api_mock.VerticalPodAutoscaler = nil
	if args.Get(0) != nil {
		returnArg = args.Get(0).([]*api_mock.VerticalPodAutoscaler)
	}
	return returnArg, args.Error(1)
}

func FakeClient(rc *apiv1.ReplicationController, rs *extensions.ReplicaSet, pods []*apiv1.Pod) kube_client.Interface {
	fakeClient := &fake.Clientset{}
	register := func(resource string, obj runtime.Object, meta metav1.ObjectMeta) {
		fakeClient.Fake.AddReactor("get", resource, func(action core.Action) (bool, runtime.Object, error) {
			getAction := action.(core.GetAction)
			if getAction.GetName() == meta.GetName() && getAction.GetNamespace() == meta.GetNamespace() {
				return true, obj, nil
			}
			return false, nil, fmt.Errorf("Not found")
		})
	}
	if rc != nil {
		register("replicationcontrollers", rc, rc.ObjectMeta)
	}
	if rs != nil {
		register("replicasets", rs, rs.ObjectMeta)
	}
	return fakeClient
}
