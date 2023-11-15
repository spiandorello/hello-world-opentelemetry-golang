package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/spiandorello/hello-word-trace/internal/people"
	"github.com/spiandorello/hello-word-trace/lib/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

var repo *people.Repository

func main() {
	ctx := context.Background()

	repo = people.NewRepository()
	defer repo.Close()

	shutdown, err := opentelemetry.Init(ctx, "hello-world-01", "v.0.1")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
	}()

	http.HandleFunc("/sayHello/", handleSayHello)

	log.Print("Listening on http://localhost:8080/")
	http.ListenAndServe(":8080", nil)
}

func handleSayHello(w http.ResponseWriter, r *http.Request) {
	ctx, span := opentelemetry.Tracing.Start(r.Context(), "handle-say-hello")
	defer span.End()

	name := strings.TrimPrefix(r.URL.Path, "/sayHello/")
	greeting, err := SayHello(ctx, name)
	if err != nil {
		span.SetAttributes(attribute.Bool("error", true))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("response", greeting))

	w.Write([]byte(greeting))
}

func SayHello(ctx context.Context, name string) (string, error) {
	ctx, span := opentelemetry.Tracing.Start(ctx, "say-hello")
	defer span.End()

	person, err := repo.GetPerson(ctx, name)
	if err != nil {
		return "", err
	}

	span.SetAttributes(
		attribute.String("Name", person.Name),
		attribute.String("Title", person.Title),
		attribute.String("Description", person.Description),
	)

	return FormatGreeting(
		ctx,
		person.Name,
		person.Title,
		person.Description,
	), nil
}

func FormatGreeting(ctx context.Context, name, title, description string) string {
	ctx, span := opentelemetry.Tracing.Start(ctx, "format-greeting")
	defer span.End()

	response := "Hello, "
	if title != "" {
		response += title + " "
	}

	response += name + "!"
	if description != "" {
		response += " " + description
	}

	return response
}
