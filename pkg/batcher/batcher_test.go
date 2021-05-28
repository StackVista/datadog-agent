package batcher

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/config"
	serializer2 "github.com/StackVista/stackstate-agent/pkg/serializer"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	testInstance  = topology.Instance{Type: "mytype", URL: "myurl"}
	testInstance2 = topology.Instance{Type: "mytype2", URL: "myurl2"}
	testHost      = "myhost"
	testAgent     = "myagent"
	testID        = check.ID("myid")
	testID2       = check.ID("myid2")
	testComponent = topology.Component{
		ExternalID: "id",
		Type:       topology.Type{Name: "typename"},
		Data:       map[string]interface{}{},
	}
	testComponent2 = topology.Component{
		ExternalID: "id2",
		Type:       topology.Type{Name: "typename"},
		Data:       map[string]interface{}{},
	}
	testRelation = topology.Relation{
		ExternalID: "id2",
		Type:       topology.Type{Name: "typename"},
		SourceID:   "source",
		TargetID:   "target",
		Data:       map[string]interface{}{},
	}
)

func TestBatchFlushOnStop(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, 100)

	batcher.SubmitStopSnapshot(testID, testInstance)

	message := serializer.GetJSONToV1IntakeMessage()

	assert.Equal(t, message,
		map[string]interface{}{
			"internalHostname": "myhost",
			"topologies": []topology.Topology{
				{
					StartSnapshot: false,
					StopSnapshot:  true,
					Instance:      testInstance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
				},
			},
		})

	batcher.Shutdown()
}

func TestBatchFlushOnComplete(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, 100)

	batcher.SubmitComponent(testID, testInstance, testComponent)

	batcher.SubmitComplete(testID)

	message := serializer.GetJSONToV1IntakeMessage()

	assert.Equal(t, message,
		map[string]interface{}{
			"internalHostname": "myhost",
			"topologies": []topology.Topology{
				{
					StartSnapshot: false,
					StopSnapshot:  false,
					Instance:      testInstance,
					Components:    []topology.Component{testComponent},
					Relations:     []topology.Relation{},
				},
			},
		})

	batcher.Shutdown()
}

func TestBatchNoDataNoComplete(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, 100)

	batcher.SubmitComponent(testID, testInstance, testComponent)

	batcher.SubmitComplete(testID2)

	// We now send a stop to trigger a combined commit
	batcher.SubmitStopSnapshot(testID, testInstance)

	message := serializer.GetJSONToV1IntakeMessage()

	assert.Equal(t, message,
		map[string]interface{}{
			"internalHostname": "myhost",
			"topologies": []topology.Topology{
				{
					StartSnapshot: false,
					StopSnapshot:  true,
					Instance:      testInstance,
					Components:    []topology.Component{testComponent},
					Relations:     []topology.Relation{},
				},
			},
		})

	batcher.Shutdown()
}

func TestBatchMultipleTopologies(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, 100)

	batcher.SubmitStartSnapshot(testID, testInstance)
	batcher.SubmitComponent(testID, testInstance, testComponent)
	batcher.SubmitComponent(testID2, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testInstance2, testComponent)
	batcher.SubmitStopSnapshot(testID, testInstance)

	message := serializer.GetJSONToV1IntakeMessage()

	assert.ObjectsAreEqualValues(map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: true,
				StopSnapshot:  true,
				Instance:      testInstance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{},
			},
			{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      testInstance2,
				Components:    []topology.Component{testComponent, testComponent, testComponent},
				Relations:     []topology.Relation{},
			},
		},
	}, message)

	batcher.Shutdown()
}

func TestBatchFlushOnMaxElements(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, 2)

	batcher.SubmitComponent(testID, testInstance, testComponent)
	batcher.SubmitComponent(testID, testInstance, testComponent2)

	message := serializer.GetJSONToV1IntakeMessage()

	assert.Equal(t, message,
		map[string]interface{}{
			"internalHostname": "myhost",
			"topologies": []topology.Topology{
				{
					StartSnapshot: false,
					StopSnapshot:  false,
					Instance:      testInstance,
					Components:    []topology.Component{testComponent, testComponent2},
					Relations:     []topology.Relation{},
				},
			},
		})

	batcher.Shutdown()
}

func TestBatchFlushOnMaxElementsEnv(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()

	// set batcher max capacity via ENV var
	os.Setenv("DD_BATCHER_CAPACITY", "1")
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, config.GetMaxCapacity())
	assert.Equal(t, 1, batcher.builder.maxCapacity)
	batcher.SubmitComponent(testID, testInstance, testComponent)

	message := serializer.GetJSONToV1IntakeMessage()
	assert.Equal(t, message,
		map[string]interface{}{
			"internalHostname": "myhost",
			"topologies": []topology.Topology{
				{
					StartSnapshot: false,
					StopSnapshot:  false,
					Instance:      testInstance,
					Components:    []topology.Component{testComponent},
					Relations:     []topology.Relation{},
				},
			},
		})

	batcher.Shutdown()
	os.Unsetenv("STS_BATCHER_CAPACITY")
}

func TestBatcherStartSnapshot(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, 100)

	batcher.SubmitStartSnapshot(testID, testInstance)
	batcher.SubmitComplete(testID)

	message := serializer.GetJSONToV1IntakeMessage()

	assert.Equal(t, message,
		map[string]interface{}{
			"internalHostname": "myhost",
			"topologies": []topology.Topology{
				{
					StartSnapshot: true,
					StopSnapshot:  false,
					Instance:      testInstance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
				},
			},
		})

	batcher.Shutdown()
}

func TestBatcherRelation(t *testing.T) {
	serializer := serializer2.NewAgentV1MockSerializer()
	batcher := newAsynchronousBatcher(serializer, testHost, testAgent, 100)

	batcher.SubmitRelation(testID, testInstance, testRelation)
	batcher.SubmitComplete(testID)

	message := serializer.GetJSONToV1IntakeMessage()

	assert.Equal(t, message,
		map[string]interface{}{
			"internalHostname": "myhost",
			"topologies": []topology.Topology{
				{
					StartSnapshot: false,
					StopSnapshot:  false,
					Instance:      testInstance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{testRelation},
				},
			},
		})

	batcher.Shutdown()
}
