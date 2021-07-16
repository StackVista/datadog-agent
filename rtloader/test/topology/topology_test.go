package testtopology

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/StackVista/stackstate-agent/rtloader/test/helpers"
)

func TestMain(m *testing.M) {
	err := setUp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up tests: %v", err)
		os.Exit(-1)
	}

	os.Exit(m.Run())
}

const topoData = `
{
  "key": "value ®",
  "stringlist": ["a", "b", "c", "4"],
  "boollist": [True, False],
  "intlist": [1],
  "doublelist": [0.7, 1.42],
  "emptykey": None,
  "nestedobject": {"nestedkey": "nestedValue"}
}`

func testTopoData(t *testing.T) {
	if result["key"] != "value ®" {
		t.Fatalf("Unexpected component data 'key' value: %s", result["key"])
	}
	var stringlist = result["stringlist"].([]interface{})
	if len(stringlist) != 4 {
		t.Fatalf("Unexpected component data 'stringlist' size: %v", len(stringlist))
	}
	if stringlist[0] != "a" && stringlist[1] != "b"  && stringlist[2] != "c" && stringlist[3] != "4" {
		t.Fatalf("Unexpected component data 'stringlist' value: %s", result["stringlist"])
	}
	var boollist = result["boollist"].([]interface{})
	if len(boollist) != 2 {
		t.Fatalf("Unexpected component data 'boollist' size: %v", len(boollist))
	}
	if boollist[0] != true && boollist[1] != false {
		t.Fatalf("Unexpected component data 'boollist' value: %s", result["boollist"])
	}
	var intlist = result["intlist"].([]interface{})
	if len(intlist) != 1 {
		t.Fatalf("Unexpected component data 'intlist' size: %v", len(intlist))
	}
	if intlist[0] != 1 {
		t.Fatalf("Unexpected component data 'intlist' value: %s", result["intlist"])
	}
	var doublelist = result["doublelist"].([]interface{})
	if len(doublelist) != 2 {
		t.Fatalf("Unexpected component data 'doublelist' size: %v", len(doublelist))
	}
	if doublelist[0] != 0.7 && doublelist[1] != 1.42 {
		t.Fatalf("Unexpected component data 'doublelist' value: %s", result["doublelist"])
	}
	if result["emptykey"] != nil {
		t.Fatalf("Unexpected component data 'emptykey' value: %s", result["emptykey"])
	}
	if result["nestedobject"] == nil {
		t.Fatalf("Unexpected component data 'nestedobject' value: %s", result["nestedobject"])
	}
	var nestedObj = result["nestedobject"].(map[interface{}]interface{})
	if nestedObj["nestedkey"] != "nestedValue" {
		t.Fatalf("Unexpected component data 'nestedkey' value: %s", nestedObj["nestedkey"])
	}
}

func TestSubmitComponent(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_component(None, "checkid", {"type": "instance.type", "url": "instance.url"}, "compid", "comptype", ` + topoData + ` )`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}
	if _instance.Type != "instance.type" {
		t.Fatalf("Unexpected instance type value: %s", _instance.Type)
	}
	if _instance.URL != "instance.url" {
		t.Fatalf("Unexpected instance type value: %s", _instance.URL)
	}
	if _externalID != "compid" {
		t.Fatalf("Unexpected component id value: %s", _externalID)
	}
	if _componentType != "comptype" {
		t.Fatalf("Unexpected component type value: %s", _componentType)
	}

	testTopoData(t)

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestSubmitComponentNoDict(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_component(None, "checkid", {"type": "instance.type", "url": "instance.url"}, "compid", "comptype", "I should be a dict")`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "TypeError: component data must be a dict" {
		t.Errorf("wrong printed value: '%s'", out)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestSubmitComponentCannotBeSerialized(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_component(None, "checkid", {"type": "instance.type", "url": "instance.url"}, "compid", "comptype", {object(): object()})`)

	if err != nil {
		t.Fatal(err)
	}
	// example error 'RepresenterError: ('cannot represent an object', <object object at 0x7fc1df8f3e90>)'
	if !strings.HasPrefix(out, "RepresenterError: ('cannot represent an object'") {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestSubmitRelation(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_relation(None, "checkid", {"type": "instance.type", "url": "instance.url"}, "source", "target", "mytype", ` + topoData + ` )`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}
	if _instance.Type != "instance.type" {
		t.Fatalf("Unexpected instance type value: %s", _instance.Type)
	}
	if _instance.URL != "instance.url" {
		t.Fatalf("Unexpected instance url value: %s", _instance.URL)
	}
	if _sourceID != "source" {
		t.Fatalf("Unexpected relation source value: %s", _sourceID)
	}
	if _targetID != "target" {
		t.Fatalf("Unexpected relation target value: %s", _targetID)
	}
	if _relationType != "mytype" {
		t.Fatalf("Unexpected relation type value: %s", _relationType)
	}

	testTopoData(t)

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestSubmitRelationNoDict(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_relation(None, "checkid", {"type": "instance.type", "url": "instance.url"}, "source", "target", "mytype", "I should be a dict")`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "TypeError: relation data must be a dict" {
		t.Errorf("wrong printed value: '%s'", out)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestSubmitRelationCannotBeSerialized(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_relation(None, "checkid", {"type": "instance.type", "url": "instance.url"}, "source", "target", "mytype", {object(): object()})`)

	if err != nil {
		t.Fatal(err)
	}
	// example error 'RepresenterError: ('cannot represent an object', <object object at 0x7fc1df8f3e90>)'
	if !strings.HasPrefix(out, "RepresenterError: ('cannot represent an object'") {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestStartSnapshot(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_start_snapshot(None, "checkid", {"type": "instance.type", "url": "instance.url"})`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}
	if _instance.Type != "instance.type" {
		t.Fatalf("Unexpected instance type value: %s", _instance.Type)
	}
	if _instance.URL != "instance.url" {
		t.Fatalf("Unexpected instance url value: %s", _instance.URL)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}

func TestStopSnapshot(t *testing.T) {
	// Reset memory counters
	helpers.ResetMemoryStats()

	out, err := run(`topology.submit_stop_snapshot(None, "checkid", {"type": "instance.type", "url": "instance.url"})`)

	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("Unexpected printed value: '%s'", out)
	}

	if checkID != "checkid" {
		t.Fatalf("Unexpected check id value: %s", checkID)
	}
	if _instance.Type != "instance.type" {
		t.Fatalf("Unexpected instance type value: %s", _instance.Type)
	}
	if _instance.URL != "instance.url" {
		t.Fatalf("Unexpected instance url value: %s", _instance.URL)
	}

	// Check for leaks
	helpers.AssertMemoryUsage(t)
}
