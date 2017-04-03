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

package update_priority

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/contrib/vertical-pod-autoscaler/updater/api_mock"
	"k8s.io/contrib/vertical-pod-autoscaler/updater/test"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"testing"
)

const (
	container_name = "container1"
)

func TestSortPriority(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", container_name, "2", "", nil)
	pod2 := test.BuildTestPod("POD2", container_name, "4", "", nil)
	pod3 := test.BuildTestPod("POD3", container_name, "1", "", nil)
	pod4 := test.BuildTestPod("POD4", container_name, "3", "", nil)

	recommendation := test.Recommendation(container_name, "10", "")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)
	calculator.AddPod(pod3, recommendation)
	calculator.AddPod(pod4, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod3, pod1, pod4, pod2}, result, "Wrong priority order")
}

func TestSortPriorityMultiResource(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", container_name, "4", "60M", nil)
	pod2 := test.BuildTestPod("POD2", container_name, "3", "90M", nil)

	recommendation := test.Recommendation(container_name, "6", "100M")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod1, pod2}, result, "Wrong priority order")
}

func TestSortPriorityMultiContainers(t *testing.T) {

	container_name2 := "container2"

	pod1 := test.BuildTestPod("POD1", container_name, "3", "10M", nil)

	pod2 := test.BuildTestPod("POD2", container_name, "4", "10M", nil)
	container2 := test.BuildTestContainer(container_name2, "3", "20M")
	pod2.Spec.Containers = append(pod1.Spec.Containers, container2)

	recommendation := test.Recommendation(container_name, "6", "20M")
	cpuRec, _ := resource.ParseQuantity("4")
	memRec, _ := resource.ParseQuantity("20M")
	container2rec := api_mock.ContainerRecommendation{Name: container_name2, Cpu: cpuRec, Memory: memRec}
	recommendation.Containers = append(recommendation.Containers, container2rec)

	calculator := NewUpdatePriorityCalculator(nil, nil)
	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod2, pod1}, result, "Wrong priority order")
}

func TestSortPriorityResorucesDecrease(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", container_name, "4", "", nil)
	pod2 := test.BuildTestPod("POD2", container_name, "10", "", nil)

	recommendation := test.Recommendation(container_name, "5", "")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod2, pod1}, result, "Wrong priority order")
}

func TestUpdateNotRequired(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)

	pod1 := test.BuildTestPod("POD1", container_name, "4", "", nil)

	recommendation := test.Recommendation(container_name, "4", "")

	calculator.AddPod(pod1, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

func TestUsePolicy(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(
		test.BuildTestPolicy(container_name, "1", "4", "10M", "100M"), nil)

	pod1 := test.BuildTestPod("POD1", container_name, "4", "10M", nil)

	recommendation := test.Recommendation(container_name, "5", "5M")

	calculator.AddPod(pod1, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result, "Pod should not be updated")
}

func TestChangeTooSmall(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, &UpdateConfig{0.5})

	pod1 := test.BuildTestPod("POD1", container_name, "4", "", nil)
	pod2 := test.BuildTestPod("POD2", container_name, "1", "", nil)

	recommendation := test.Recommendation(container_name, "5", "")

	calculator.AddPod(pod1, recommendation)
	calculator.AddPod(pod2, recommendation)

	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{pod2}, result, "Only POD2 should be updated")
}

func TestNoPods(t *testing.T) {
	calculator := NewUpdatePriorityCalculator(nil, nil)
	result := calculator.GetSortedPods()
	assert.Exactly(t, []*apiv1.Pod{}, result)
}
