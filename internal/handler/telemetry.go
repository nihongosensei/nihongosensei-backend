package handler

import "go.opentelemetry.io/otel"

const name = "github.com/nihongosensei/nihongosensei-backend/internal/handler"

var (
	tracer = otel.Tracer(name)
)
