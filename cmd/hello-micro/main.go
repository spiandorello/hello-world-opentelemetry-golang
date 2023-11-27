package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/spiandorello/hello-word-trace/lib/model"
	"github.com/spiandorello/hello-word-trace/lib/opentelemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	ctx := context.Background()

	shutdown, err := opentelemetry.Init(ctx, "hello-microservice", "v0.0.1")
	if err != nil {
		panic(err)
	}
	defer func() {
		err := errors.Join(shutdown(ctx))
		if err != nil {
			panic(err)
		}
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

	person, err := getPerson(ctx, name)
	if err != nil {
		return "", err
	}

	return formatGreeting(ctx, person)
}

func getPerson(ctx context.Context, name string) (*model.Person, error) {
	ctx, span := opentelemetry.Tracing.Start(ctx, "get-person")
	defer span.End()

	res, err := get(ctx, "http://localhost:8081/getPerson/"+name)
	if err != nil {
		return nil, err
	}

	var person model.Person
	if err = json.Unmarshal(res, &person); err != nil {
		return nil, err
	}

	return &person, err
}

func formatGreeting(ctx context.Context, person *model.Person) (string, error) {
	ctx, span := opentelemetry.Tracing.Start(ctx, "format-greeting")
	defer span.End()

	query := url.Values{}
	query.Set("title", person.Title)
	query.Set("name", person.Name)
	query.Set("description", person.Description)

	body, err := get(ctx, "http://localhost:8082/formatGreeting?"+query.Encode())
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func get(ctx context.Context, url string) ([]byte, error) {
	ctx, span := opentelemetry.Tracing.Start(ctx, "http-get")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	otel.GetTextMapPropagator().Inject(
		ctx,
		propagation.HeaderCarrier(req.Header),
	)

	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
