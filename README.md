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
    l := zerolog.Level(-3)
    zerolog.SetGlobalLevel(l)
    zl := zerolog.New(os.Stdout).Level(l)
    log := zerologr.New(&zl)
    log.V(3).Info("v3")
    log.V(4).Info("you should NOT see this")
```

Zerolog's levels get more verbose as the number gets smaller, and more important
as the number gets larger.

The `-3` in the above snippet means that `log.V(3).Info()` calls will be active.
`-4` would enable `log.V(4).Info()`, etc.  Note that Zerolog's levels are `int8`
which means the most verbose level you can give it is -128.  The Zerologr
implementation will cap `V()` levels greater than 128 to 128, so setting the
Zerolog level to -128 really means "activate all logs".

## Implementation Details

For the most part, concepts in Zerolog correspond directly with those in logr.

Zerolog uses semantically named levels for logging (`TraceLevel`, `DebugLevel`,
`InfoLevel`, `WarningLevel`, ...). Logr uses arbitrary numeric levels.

Levels in logr correspond to named levels in Zerolog. If the given level in logr
matches a named level, it is represented by `zerologLevel = 1 - logrLevel`. If
the level in logr is more verbose than a named Zerolog level (`TraceLevel`), it
is represented by `zerologLevel = -logrLevel`.

For example logr's `V(0)` is equivalent to Zerolog's `InfoLevel`, `V(1)` is equivalent
to Zerolog's `DebugLevel`, `V(2)` is equivalent to Zerolog's `TraceLevel`, while
`V(3)` is equivalent to Zerolog handling it numerically as `-3`.
