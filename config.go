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

// TracerConfig defines the expected values for configuring an application tracer. It is recommended to initialize a
// TracerConfig with either bobotel.NewTracerConfig().
type TracerConfig struct {
	bconf.ConfigStruct
	AppID             string   `bconf:"app.id"`
	AppName           string   `bconf:"app.name"`
	OtelExporters     []string `bconf:"otel.exporters"`
	OtelConsoleFormat string   `bconf:"otel.console_format"`
	OtlpEndpointKind  string   `bconf:"otlp.endpoint_kind"`
	OtlpHost          string   `bconf:"otlp.host"`
	OtlpPort          int      `bconf:"otlp.port"`
}

// TracerFieldSets defines the field-sets for an application tracer.
func FieldSets() bconf.FieldSets {
	return bconf.FieldSets{
		OtelFieldSet(),
		OtlpFieldSet(),
	}
}

// OtelFieldSet ...
func OtelFieldSet() *bconf.FieldSet {
	return bconf.FSB(OtelFieldSetKey).Fields(
		bconf.FB(OtelExportersKey, bconf.Strings).Default([]string{"console"}).Validator(
			func(v any) error {
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
			},
		).C(),
		bconf.FB(OtelConsoleFormatKey, bconf.String).Default("production").Enumeration("production", "pretty").C(),
	).C()
}

// OtlpFieldSet ...
func OtlpFieldSet() *bconf.FieldSet {
	return bconf.FSB(OtlpFieldSetKey).Fields(
		bconf.FB(OtlpEndpointKindKey, bconf.String).Default("agent").Enumeration("agent", "collector").C(),
		bconf.FB(OtlpHostKey, bconf.String).Required().C(),
		bconf.FB(OtlpPortKey, bconf.Int).Default(6831).C(),
	).LoadConditions(
		bconf.LCB(
			func(f bconf.FieldValueFinder) (bool, error) {
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
			},
		).AddFieldSetDependencies(OtelFieldSetKey, OtelExportersKey).C(),
	).C()
}
