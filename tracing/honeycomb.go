package tracing

import (
	export "go.opentelemetry.io/otel/sdk/export/trace"
)

type Honeycomb struct {
	APIKey      string `long:"honeycomb-api-key" description:"honeycomb.io api key"`
	Dataset     string `long:"honeycomb-dataset" description:"honeycomb.io dataset name"`
	ServiceName string `long:"honeycomb-service-name" description:"honeycomb.io service name" default:"concourse"`
}

func (h Honeycomb) IsConfigured() bool {
	return h.APIKey != "" && h.Dataset != ""
}

func (h Honeycomb) Exporter() (export.SpanExporter, error) {
	// TODO: ServiceName?
	return OTLP{
		Address: "api.honeycomb.io:443",
		Headers: map[string]string{
			"x-honeycomb-team":    h.APIKey,
			"x-honeycomb-dataset": h.Dataset,
		},
		UseTLS: true,
	}.Exporter()
}
