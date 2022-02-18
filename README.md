# otel-go-contrib

## Usage

### Gin Opentelemetry Metrics

```golang
import "github.com/technologize/otel-go-contrib/otelginmetrics"
router := gin.Default()
router.Use(otelginmetrics.Middleware("hello world"))
```
