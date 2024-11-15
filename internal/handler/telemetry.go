package handler

import "go.opentelemetry.io/otel"

const name = "github.com/nihongosensei/nihongosensei-go/internal/handler"

var (
	tracer = otel.Tracer(name)
)
