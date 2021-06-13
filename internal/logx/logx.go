package logx

import (
	"context"
	"fmt"
	"github.com/blendle/zapdriver"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

type AppLogger struct {
	zap       *zap.Logger
	projectID string
}

func NewDevLogger(projectID string) (*AppLogger, error) {
	config := zapdriver.NewDevelopmentConfig()
	config.Encoding = "console"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	clientLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("config.Build(): %v", err)
	}
	return &AppLogger{zap: clientLogger, projectID: projectID}, nil
}

func NewProdLogger(projectID string) (*AppLogger, error) {
	config := zapdriver.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	clientLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("config.Build(): %v", err)
	}
	return &AppLogger{zap: clientLogger, projectID: projectID}, nil
}

func NewTesterLogger(t *testing.T) *AppLogger {
	development, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("zap.NewDevelopment() err = %v; want nil", err)
	}
	return &AppLogger{zap: development, projectID: "fake"}
}

func (i *AppLogger) WrapTraceContext(ctx context.Context) *zap.SugaredLogger {
	sc := trace.SpanContextFromContext(ctx)
	fields := zapdriver.TraceContext(sc.TraceID().String(), sc.SpanID().String(), sc.IsSampled(), i.projectID)
	setFields := i.zap.With(fields...)
	return setFields.Sugar()
}

func (i *AppLogger) Sync() {
	i.zap.Sync()
}

func (i *AppLogger) Infof(template string, args ...interface{}) {
	i.zap.Info(fmt.Sprintf(template, args...))
}

func (i *AppLogger) Info(s string) {
	i.zap.Info(s)
}
