package bobotel

import (
	"fmt"
	"slices"

	"github.com/xavi-group/bconf"
)

const (
	// OtelFieldSetKey defines the field-set key for open-telemetery configuration fields.
	OtelFieldSetKey = "otel"
	// OtlpFieldSetKey defines the field-set key for open-telemetry protocol configuration fields.
	OtlpFieldSetKey = "otlp"

	// OtelExportersKey defines the field key for the open-telemetry exporters field.
	OtelExportersKey = "exporters"
	// OtelConsoleFormatKey defines the field key for the open-telemetry console_format field.
	OtelConsoleFormatKey = "console_format"

	// OtlpEndpointKindKey defines the field key for the open-telemetry protocol endpoint_kind field.
	OtlpEndpointKindKey = "endpoint_kind"
	// OtlpHostKey defines the field key for the open-telemetry protocol host field.
	OtlpHostKey = "host"
	// OtlpPortKey defines the field key for the open-telemetry protocol port field.
	OtlpPortKey = "port"
)

// NewConfig provides an initialized Config struct, and sets the returned config struct as the default config used when
// calling InitializeTraceProvider(config ...*Config) with no args.
func NewConfig() *Config {
	configLock.Lock()
	defer configLock.Unlock()

	defaultConfig = &Config{}

	return defaultConfig
}

// Config defines the expected values for configuring an open-telemetry tracer. It is recommended to initialize a
// Config with bobotel.NewConfig(), which will set the default configuration struct for initializing a trace provider.
type Config struct {
	bconf.ConfigStruct
	AppID             string   `bconf:"app.id"`
	AppName           string   `bconf:"app.name"`
	OtelExporters     []string `bconf:"otel.exporters"`
	OtelConsoleFormat string   `bconf:"otel.console_format"`
	OtlpEndpointKind  string   `bconf:"otlp.endpoint_kind"`
	OtlpHost          string   `bconf:"otlp.host"`
	OtlpPort          int      `bconf:"otlp.port"`
}

// FieldSets defines the field-sets for an open-telemetry tracer.
func FieldSets() bconf.FieldSets {
	return bconf.FieldSets{
		OtelFieldSet(),
		OtlpFieldSet(),
	}
}

// OtelFieldSet ...
func OtelFieldSet() *bconf.FieldSet {
	return bconf.FSB(OtelFieldSetKey).Fields(
		bconf.FB(OtelExportersKey, bconf.Strings).Default([]string{"console"}).Validator(otelExportersValidator).
			Description(
				"Otel exporters defines where traces will be sent (accepted values are 'console' and 'otlp'). ",
				"Exporters accepts a list and can be configured to export traces to multiple destinations.",
			).C(),
		bconf.FB(OtelConsoleFormatKey, bconf.String).Default("production").Enumeration("production", "pretty").
			Description(
				"Otel console format defines the format of traces output to the console where 'pretty' is more ",
				"human readable (adds whitespace).",
			).C(),
	).C()
}

// OtlpFieldSet ...
func OtlpFieldSet() *bconf.FieldSet {
	return bconf.FSB(OtlpFieldSetKey).Fields(
		bconf.FB(OtlpEndpointKindKey, bconf.String).Default("http").Enumeration("http", "grpc").
			Description("Otlp endpoint kind defines the protocol used by the trace collector.").C(),
		bconf.FB(OtlpHostKey, bconf.String).Required().
			Description("Otlp host defines the host location of the trace collector.").C(),
		bconf.FB(OtlpPortKey, bconf.Int).Default(4318).
			Description(
				"Otlp port defines the port of the trace collector process. For a GRPC endpoint the default is 4317.",
			).C(),
	).LoadConditions(
		bconf.LCB(otlpLoadCondition).AddFieldSetDependencies(OtelFieldSetKey, OtelExportersKey).C(),
	).C()
}

func otlpLoadCondition(f bconf.FieldValueFinder) (bool, error) {
	exporters, found, err := f.GetStrings(OtelFieldSetKey, OtelExportersKey)
	if !found || err != nil {
		return false, fmt.Errorf("problem getting exporters field value")
	}

	otlpExporterFound := false
	for _, exporter := range exporters {
		if exporter == "otlp" {
			otlpExporterFound = true

			break
		}
	}

	return otlpExporterFound, nil
}

func otelExportersValidator(v any) error {
	acceptedValues := []string{"console", "otlp"}

	fieldValues, ok := v.([]string)
	if !ok {
		return fmt.Errorf("unexpected field-value type provided to validator")
	}

	for _, value := range fieldValues {
		if found := slices.Contains(acceptedValues, value); !found {
			return fmt.Errorf("invalid exporter value: '%s'", value)
		}
	}

	return nil
}
