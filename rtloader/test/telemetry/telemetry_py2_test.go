// +build two

package testtelemetry

import (
	"github.com/StackVista/stackstate-agent/rtloader/test/helpers"
	"testing"
)

func TestSubmitTopologyChangeRequestEvents(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`telemetry.submit_topology_event(
							None,
							"checkid",
							{
								'event_type': 'Change Request Normal',
								'tags': ['number:CHG0000001', 'priority:3 - Moderate', 'risk:High', 'state:New', 'category:Software', 'conflict_status:None', 'assigned_to:ITIL User'],
								'timestamp': 1600951343,
								'msg_title': 'CHG0000001: Rollback Oracle \xc2\xae Version',
								'msg_text': 'Performance of the Siebel SFA software has been severely\n            degraded since the upgrade performed this weekend.\n\n            We moved to an unsupported Oracle DB version. Need to rollback the\n            Oracle Instance to a supported version.\n        ',
								'context': {
									'category': 'change_request',
									'source': 'servicenow',
									'data': {'impact': '3 - Low', 'requested_by': 'David Loo'},
									'element_identifiers': ['a9c0c8d2c6112276018f7705562f9cb0', 'urn:host:/Sales \xc2\xa9 Force Automation', 'urn:host:/sales \xc2\xa9 force automation'],
									'source_links': []
								},
								'source_type_name': 'Change Request Normal'
							}
				)
				`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	testChangeRequestEventsBase(t)

	if _topoEvt.Title != "CHG0000001: Rollback Oracle ® Version" {
		t.Fatalf("Unexpected topology event data 'msg_title' value: '%s'", _topoEvt.Title)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}
