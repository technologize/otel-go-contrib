# otel-go-contrib

## Usage

### Gin Opentelemetry Metrics

```golang
router := gin.Default()
router.Use(otelginmetrics.Middleware("hello world"))
```
