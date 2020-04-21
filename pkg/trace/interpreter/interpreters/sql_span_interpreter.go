package interpreters

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/util"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// SQLSpanInterpreter sets up the sql span interpreter
type SQLSpanInterpreter struct {
	interpreter
}

// SQLSpanInterpreterName is the name used for matching this interpreter
const SQLSpanInterpreterName = "sql"

// DatabaseTypeName returns the default database type
const DatabaseTypeName = "database"

// MakeSQLSpanInterpreter creates an instance of the sql span interpreter
func MakeSQLSpanInterpreter(config *config.Config) *SQLSpanInterpreter {
	return &SQLSpanInterpreter{interpreter{Config: config}}
}

// Interpret performs the interpretation for the SQLSpanInterpreter
func (in *SQLSpanInterpreter) Interpret(span *util.SpanWithMeta) *pb.Span {
	dbType := DatabaseTypeName

	// no meta, add a empty map
	if span.Meta == nil {
		span.Meta = map[string]string{}
	}

	if database, found := span.Meta["db.type"]; found {
		dbType = database
	}
	span.Meta["span.serviceType"] = dbType

	// create the service instance identifier using the already interpreted name
	span.Meta["span.serviceInstanceURN"] = util.CreateServiceInstanceURN(span.Meta["span.serviceName"], span.Hostname, span.PID, span.CreateTime)

	return span.Span
}