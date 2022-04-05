package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/technologize/otel-go-contrib/otelginmetrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	metricGlobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func RandomValue(values []string) string {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	return values[rand.Intn(len(values))]
}

func initMetrics() {

	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		panic(err)
	}
	fmt.Println(metricExporter)
	res, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String("PG2")),
		resource.WithAttributes(semconv.ServiceNamespaceKey.String("Spark")),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		panic(err)
	}
	metricProvider := controller.New(
		processor.NewFactory(
			selector.NewWithInexpensiveDistribution(),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
		controller.WithCollectPeriod(1*time.Second),
		controller.WithResource(res),
		controller.WithExporter(metricExporter),
	)
	if err := metricProvider.Start(context.Background()); err != nil {
		log.Fatalln("failed to start the metric controller:", err)
	}
	metricGlobal.SetMeterProvider(metricProvider)
}

func main() {
	router := gin.New()
	initMetrics()
	router.Use(otelginmetrics.Middleware(
		"TEST-SERVICE",
		// Custom attributes
		otelginmetrics.WithAttributes(func(serverName, route string, request *http.Request) []attribute.KeyValue {
			return append(otelginmetrics.DefaultAttributes(serverName, route, request), attribute.String("Custom-attribute", "value"))
		}),
	))

	logic := func(ctx *gin.Context, sleep int) {
		// xxx, _ := strconv.Atoi(ctx.Param("xxx"))
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	router.GET("/test/:xxx", func(ctx *gin.Context) {
		logic(ctx, 1)
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("xxx"),
		})
	})

	go func() {
		for {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test/1")
		}
	}()

	_ = router.Run(":9199")

}
