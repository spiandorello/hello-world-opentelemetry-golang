package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/spiandorello/hello-word-trace/internal/people"
	"github.com/spiandorello/hello-word-trace/lib/model"
	"github.com/spiandorello/hello-word-trace/lib/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

var repo *people.Repository

func main() {
	ctx := context.Background()

	repo = people.NewRepository()
	defer repo.Close()

	shutdown, err := opentelemetry.Init(ctx, "bigbrother-api", "v0.0.1")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
		if err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/getPerson/", handleGetPerson)

	log.Print("Listening on http://localhost:8081/")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleGetPerson(w http.ResponseWriter, r *http.Request) {
	ctx := otel.GetTextMapPropagator().Extract(
		r.Context(), propagation.HeaderCarrier(r.Header),
	)

	ctx, span := opentelemetry.Tracing.Start(ctx, "/handle-get-person")
	defer span.End()

	name := strings.TrimPrefix(r.URL.Path, "/getPerson/")

	person, err := getPerson(ctx, name)
	if err != nil {
		span.SetAttributes(attribute.Bool("error", true))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, _ := json.Marshal(person)
	w.Write(bytes)
}

func getPerson(ctx context.Context, name string) (model.Person, error) {
	ctx, span := opentelemetry.Tracing.Start(ctx, "get-person")
	defer span.End()

	person, err := repo.GetPerson(ctx, name)
	if err != nil {
		return model.Person{}, err
	}

	span.SetAttributes(
		attribute.String("Name", person.Name),
		attribute.String("Title", person.Title),
		attribute.String("Description", person.Description),
	)

	return person, nil
}
