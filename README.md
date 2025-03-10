# `bobotel`: base observability / open-telemetry

`bobotel` is a support package that provides bconf configuration for open-telemetry tracing, which is a great way to
improve application observability via tracing.

This package additionally provides helper functions for initializing a global otel trace provider, and creating new
tracers.

```sh
go get github.com/xavi-group/bobotel
```

## Configuration

```
Conditionally Required Configuration:
        otlp.host string
                Otlp host defines the host location of the trace collector.
                Environment key: 'OTLP_HOST'
                Flag argument: '--otlp_host'
                Loading depends on field(s): 'otel.exporters'
Optional Configuration:
        otel.console_format string
                Otel console format defines the format of traces output to the console where 'pretty' is more 
                human readable (adds whitespace). 
                Accepted values: ['production', 'pretty']
                Default value: 'production'
                Environment key: 'OTEL_CONSOLE_FORMAT'
                Flag argument: '--otel_console_format'
        otel.exporters []string
                Otel exporters defines where traces will be sent (accepted values are 'console' and 'otlp'). 
                Exporters accepts a list and can be configured to export traces to multiple destinations. 
                Default value: '[console]'
                Environment key: 'OTEL_EXPORTERS'
                Flag argument: '--otel_exporters'
        otlp.endpoint_kind string
                Otlp endpoint kind defines the protocol used by the trace collector.
                Accepted values: ['http', 'grpc']
                Default value: 'http'
                Environment key: 'OTLP_ENDPOINT_KIND'
                Flag argument: '--otlp_endpoint_kind'
                Loading depends on field(s): 'otel.exporters'
        otlp.port int
                Otlp port defines the port of the trace collector process. For a GRPC endpoint the default is 
                4317. 
                Default value: '4318'
                Environment key: 'OTLP_PORT'
                Flag argument: '--otlp_port'
                Loading depends on field(s): 'otel.exporters'
```

## Example

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/segmentio/ksuid"
	"github.com/xavi-group/bconf"
	"github.com/xavi-group/bobotel"
	"go.opentelemetry.io/otel/codes"
)

func main() {
	config := bconf.NewAppConfig(
		"bobotelexample",
		"Example application showcasing bobotel tracing",
		bconf.WithAppIDFunc(func() string { return ksuid.New().String() }),
		bconf.WithAppVersion("1.0.0"),
		bconf.WithEnvironmentLoader(),
		bconf.WithFlagLoader(),
	)

	config.AddFieldSetGroup("bobotel", bobotel.FieldSets())

	config.AttachConfigStructs(
		bobotel.NewConfig(),
	)

	// Load when called without any options will also handle the help flag (--help or -h)
	if errs := config.Load(); len(errs) > 0 {
		fmt.Printf("problem(s) loading application configuration: %v\n", errs)
		os.Exit(1)
	}

	// -- Initialize application observability --
	if err := bobotel.InitializeTraceProvider(); err != nil {
		fmt.Printf("problem initializing application tracing: %s\n", err)
		os.Exit(1)
	}

	tracer := bobotel.NewTracer("main")

	startupCtx := context.Background()
	_, span := tracer.Start(startupCtx, "main")

	span.SetStatus(codes.Ok, "startup success")
	span.End()
}
```

## Support

For more information on open-telemetry, check out and support the open-telemetry project at
[opentelemetry.io](https://opentelemetry.io/)

For more information on bconf, check out and support the bconf project at
[github.com/xavi-group/bconf](https://github.com/xavi-group/bconf)
