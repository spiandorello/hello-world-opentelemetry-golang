package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/spiandorello/hello-word-trace/lib/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	ctx := context.Background()

	shutdown, err := opentelemetry.Init(ctx, "formatter-api", "v.0.1")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
		if err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/formatGreeting", handleFormatGreeting)

	log.Print("Listening on http://localhost:8082/")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func handleFormatGreeting(w http.ResponseWriter, r *http.Request) {
	ctx := otel.GetTextMapPropagator().Extract(
		r.Context(), propagation.HeaderCarrier(r.Header),
	)

	_, span := opentelemetry.Tracing.Start(ctx, "/formatGreeting")
	defer span.End()

	title := r.URL.Query().Get("title")
	name := r.URL.Query().Get("name")
	description := r.URL.Query().Get("description")

	response := "Hello, "
	if title != "" {
		response += title + " "
	}

	response += name + "!"
	if description != "" {
		response += " " + description
	}

	w.Write([]byte(response))
}
