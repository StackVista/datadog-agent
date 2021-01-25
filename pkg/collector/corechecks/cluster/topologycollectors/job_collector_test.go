// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var parralelism int32
var backoffLimit int32

func TestJobCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	parralelism = int32(2)
	backoffLimit = int32(5)
	var numberOfComponents int

	jc := NewJobCollector(componentChannel, relationChannel, NewTestCommonClusterCollector(MockJobAPICollectorClient{}))
	expectedCollectorName := "Job Collector"
	RunCollectorTest(t, jc, expectedCollectorName)

	for _, tc := range []struct {
		testCase          string
		expectedComponent *topology.Component
		expectedRelations []*topology.Relation
	}{
		{
			testCase: "Test Job 1 + Cron Job Relations",
			expectedComponent: &topology.Component{
				ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
				Type:       topology.Type{Name: "job"},
				Data: topology.Data{
					"name":              "test-job-1",
					"creationTimestamp": creationTime,
					"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					"uid":               types.UID("test-job-1"),
					"backoffLimit":      &backoffLimit,
					"parallelism":       &parralelism,
				},
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
					Type:     topology.Type{Name: "creates"},
					SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:cronjob/test-cronjob-1",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-1",
					Data:     map[string]interface{}{},
				},
			},
		},
		{
			testCase: "Test Job 2",
			expectedComponent: &topology.Component{
				ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
				Type:       topology.Type{Name: "job"},
				Data: topology.Data{
					"name":              "test-job-2",
					"creationTimestamp": creationTime,
					"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					"uid":               types.UID("test-job-2"),
					"backoffLimit":      &backoffLimit,
					"parallelism":       &parralelism,
				},
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
					Type:     topology.Type{Name: "encloses"},
					SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-2",
					Data:     map[string]interface{}{},
				},
			},
		},
		{
			testCase: "Test Job 3 - Kind + Generate Name",
			expectedComponent: &topology.Component{
				ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
				Type:       topology.Type{Name: "job"},
				Data: topology.Data{
					"name":              "test-job-3",
					"creationTimestamp": creationTime,
					"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					"uid":               types.UID("test-job-3"),
					"kind":              "some-specified-kind",
					"generateName":      "some-specified-generation",
					"backoffLimit":      &backoffLimit,
					"parallelism":       &parralelism,
				},
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
					Type:     topology.Type{Name: "encloses"},
					SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:job/test-job-3",
					Data:     map[string]interface{}{},
				},
			},
		},
		{
			testCase:          "Test Job 4 - No container or relation should not be created",
			expectedComponent: nil,
			expectedRelations: nil,
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			if tc.expectedComponent == nil && tc.expectedRelations == nil {
				// test-job-4 is skipped because it's completed
				assert.Empty(t, componentChannel)
				assert.Empty(t, relationChannel)
			} else {
				// test-job-1 .. test-job-3
				component := <-componentChannel
				numberOfComponents += 1
				assert.EqualValues(t, tc.expectedComponent, component)

				for _, expectedRelation := range tc.expectedRelations {
					cronJobRelation := <-relationChannel
					assert.EqualValues(t, expectedRelation, cronJobRelation)
				}
			}
		})
	}

	t.Run("noTestJob4", func(t *testing.T) {
		// we should have only 3 components, test-job-4 is completed so no component is created
		assert.Equal(t, 3, numberOfComponents)
	})

}

type MockJobAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockJobAPICollectorClient) GetJobs() ([]batchV1.Job, error) {
	completionTime := v1.Time{Time: time.Now()}

	jobs := make([]batchV1.Job, 0)
	for i := 1; i <= 4; i++ {
		job := batchV1.Job{
			TypeMeta: v1.TypeMeta{
				Kind: "",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-job-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:          types.UID(fmt.Sprintf("test-job-%d", i)),
				GenerateName: "",
			},
			Spec: batchV1.JobSpec{
				Parallelism:  &parralelism,
				BackoffLimit: &backoffLimit,
			},
		}

		if i == 1 {
			job.OwnerReferences = []v1.OwnerReference{
				{Kind: CronJob, Name: "test-cronjob-1"},
			}
		}

		if i == 3 {
			job.TypeMeta.Kind = "some-specified-kind"
			job.ObjectMeta.GenerateName = "some-specified-generation"
		}

		if i == 4 {
			job.Status.CompletionTime = &completionTime
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
