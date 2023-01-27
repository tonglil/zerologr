# Zerologr

[![Go Reference](https://pkg.go.dev/badge/github.com/go-logr/zerologr.svg)](https://pkg.go.dev/github.com/go-logr/zerologr)
![test](https://github.com/go-logr/zerologr/workflows/test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-logr/zerologr)](https://goreportcard.com/report/github.com/go-logr/zerologr)

A [logr](https://github.com/go-logr/logr) LogSink implementation using [Zerolog](https://github.com/rs/zerolog).

## Usage

```go
import (
    "os"

    "github.com/go-logr/logr"
    "github.com/go-logr/zerologr"
    "github.com/rs/zerolog"
)

func main() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
    zerologr.NameFieldName = "logger"
    zerologr.NameSeparator = "/"

    zl := zerolog.New(os.Stderr)
    zl = zl.With().Caller().Timestamp().Logger()
    var log logr.Logger = zerologr.New(&zl)

    log.Info("Logr in action!", "the answer", 42)
}
```

## Increasing verbosity

You can do something like the following in your setup code:

```go
    l := zerolog.Level(1 - 3)
    zerolog.SetGlobalLevel(l)
    zl := zerolog.New(os.Stdout).Level(l)
    log := zerologr.New(&zl)
    log.V(3).Info("you should see this")
    log.V(4).Info("you should NOT see this")
```

Zerolog's levels get more verbose as the number gets smaller, and more important
as the number gets larger.

The `1 - 3` in the above snippet means that `log.V(3).Info()` calls will be
active. `1 - 4` would enable `log.V(4).Info()`, etc.  Note that Zerolog's levels
are `int8` which means the most verbose level you can give it is `1 - 129`
(-128). The Zerologr implementation will cap `V()` levels greater than 129 to
129, so setting the Zerolog level to -128 really means "activate all logs".

## Implementation Details

For the most part, concepts in Zerolog correspond directly with those in logr.

Zerolog uses semantically named levels for logging (`TraceLevel`, `DebugLevel`,
`InfoLevel`, `WarningLevel`, ...). Logr uses arbitrary numeric levels.

Levels in logr correspond to levels in Zerolog. Any given level in logr is
represents by `zerologLevel = 1 - logrLevel`.

For example logr's `V(0)` is Zerolog's `InfoLevel` (which is numerically 1),
`V(1)` is Zerolog's `DebugLevel` (which is numerically 0), and `V(2)` is
Zerolog's `TraceLevel` (which is numerically -1). Zerolog does not have named
levels that are more verbose than `TraceLevel`, and instead handles them
numerically where `V(3)` is equivalent to Zerolog handling it as `-2`.
