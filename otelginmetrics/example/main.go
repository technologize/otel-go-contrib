package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/technologize/otel-go-contrib/otelginmetrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	metricGlobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type Transactions struct {
	TransactionId        string        `gorm:"primaryKey, column:transaction_id"`
	ClientCode           string        `gorm:"column:client_code"`
	BankCode             string        `gorm:"column:bank_code"`
	BankAccountNumber    string        `gorm:"column:bank_account_number"`
	BankIFSCCode         string        `gorm:"column:bank_ifsc_code"`
	Amount               float64       `gorm:"column:amount"`
	Product              string        `gorm:"column:product"`
	RequestSource        string        `gorm:"column:request_source"`
	InitiatedDatetime    int64         `gorm:"column:initiated_datetime"`
	AppName              string        `gorm:"column:app_name"`
	VPA                  string        `gorm:"column:vpa"`
	Type                 string        `gorm:"column:type"`
	SubType              string        `gorm:"column:sub_type"`
	PaymentProvider      string        `gorm:"column:provider"`
	Status               string        `gorm:"column:status"`
	BankReferenceNumber  string        `gorm:"column:bank_reference_number"`
	UPITransactionRefNum string        `gorm:"column:upi_transaction_ref_num"`
	BankErrorCode        string        `gorm:"column:bank_error_code"`
	BankErrorDescription string        `gorm:"column:bank_error_description"`
	ReconcileAttempt     int64         `gorm:"column:reconcile_attempt"`
	ReconcileDatetime    sql.NullInt64 `gorm:"column:reconcile_datetime"`
	LastUpdatedDatetime  sql.NullInt64 `gorm:"column:last_updated_datetime"`
	TransactionDatetime  string        `gorm:"column:transaction_datetime"`
}

var (
	meter               = global.Meter("github.com/angel-finoux/pg2", metric.WithInstrumentationVersion("1.0.0"))
	transactionsCounter = metric.Must(meter).NewInt64UpDownCounter("pgc.transactions.count", metric.WithDescription("Number of Transactions in PG2"), metric.WithUnit(unit.Dimensionless))
)

func RecordMetrics(ctx context.Context, transaction Transactions) {
	transactionAttributes := make([]attribute.KeyValue, 0, 1)
	transactionAttributes = append(transactionAttributes, attribute.String("status", transaction.Status))
	transactionAttributes = append(transactionAttributes, attribute.String("provider", transaction.PaymentProvider))
	transactionAttributes = append(transactionAttributes, attribute.String("bank_code", transaction.BankCode))
	transactionAttributes = append(transactionAttributes, attribute.String("type", transaction.Type))
	transactionAttributes = append(transactionAttributes, attribute.String("upi_app", transaction.AppName))
	transactionAttributes = append(transactionAttributes, attribute.String("request_source", transaction.RequestSource))
	transactionsCounter.Add(ctx, 1, transactionAttributes...)
}

func RandomValue(values []string) string {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	return values[rand.Intn(len(values))]
}

func initMetrics() {

	// metricExporter, _ := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	metricExporter, err := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithInsecure())
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
		transaction := Transactions{
			Status:          RandomValue([]string{"success", "failure"}),
			PaymentProvider: RandomValue([]string{"hdfc", "icici", "axis", "atom", "razorpay", "techprocess"}),
			BankCode:        RandomValue([]string{"hdf", "ici", "axi", "sbi"}),
			Type:            RandomValue([]string{"NetBanking", "UPI"}),
			AppName:         RandomValue([]string{"GPay", "Phonepe", "Paytm"}),
			RequestSource:   RandomValue([]string{"ABMA", "SPARK-iOS", "Spark-Android"}),
		}
		fmt.Println(transaction.Status + "==" + transaction.PaymentProvider + "==" + transaction.BankCode + "==" + transaction.Type + "==" + transaction.AppName + "==" + transaction.RequestSource)
		RecordMetrics(ctx, transaction)
	}

	router.GET("/test/:xxx", func(ctx *gin.Context) {
		logic(ctx, 1)
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("xxx"),
		})
	})

	router.GET("/test1/:xxx", func(ctx *gin.Context) {
		logic(ctx, 5)
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("xxx"),
		})
	})

	router.GET("/test2/:xxx", func(ctx *gin.Context) {
		logic(ctx, 10)
		ctx.JSON(200, map[string]string{
			"productId": ctx.Param("xxx"),
		})
	})

	go func() {
		for {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test/1")
		}
	}()
	go func() {
		for {
			_, _ = http.DefaultClient.Get("http://localhost:9199/test/1")
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
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
