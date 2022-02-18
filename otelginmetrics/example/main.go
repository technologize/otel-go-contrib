package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/technologize/otel-go-contrib/otelginmetrics"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricGlobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func initMetrics() {

	// metricExporter, _ := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	metricExporter, err := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	fmt.Println(metricExporter)
	res, err := resource.New(context.Background(),
		resource.WithAttributes(semconv.ServiceNameKey.String("PG2")),
		resource.WithAttributes(semconv.ServiceNamespaceKey.String("CUSTOM3")),
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
	router := gin.Default()
	initMetrics()
	router.Use(otelginmetrics.Middleware(
		"TEST-SERVICE",
		otelginmetrics.WithAdditionalAttributes(map[string]string{"label": "value"}),
	))

	router.GET("/test/:xxx", func(ctx *gin.Context) {
		xxx, _ := strconv.Atoi(ctx.Param("xxx"))
		time.Sleep(time.Duration(xxx) * time.Second)
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("xxx"),
		})
	})

	router.GET("/test1/:xxx", func(ctx *gin.Context) {
		xxx, _ := strconv.Atoi(ctx.Param("xxx"))
		time.Sleep(time.Duration(xxx) * 5 * time.Second)
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("xxx"),
		})
	})

	router.GET("/test2/:xxx", func(ctx *gin.Context) {
		xxx, _ := strconv.Atoi(ctx.Param("xxx"))
		time.Sleep(time.Duration(xxx) * 10 * time.Second)
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("xxx"),
		})
	})

	go func() {
		for i := 0; i < 100; i++ {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test/" + strconv.Itoa(i))
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test/" + strconv.Itoa(i))
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test/" + strconv.Itoa(i))
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test1/" + strconv.Itoa(i))
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test2/" + strconv.Itoa(i))
		}
	}()

	_ = router.Run(":9199")

}
