package logger

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestJSONLogger(t *testing.T) {
	logger, err := New(
		WithField("defined_key", "defined_value"),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	err = errors.New("pkg error")
	logger.Error("err occurs", WrapMeta(nil, NewMeta("para1", "value1"), NewMeta("para2", "value2"))...)
	logger.Error("err occurs", WrapMeta(err, NewMeta("para1", "value1"), NewMeta("para2", "value2"))...)

}

func BenchmarkJsonLogger(b *testing.B) {
	b.ResetTimer()
	logger, err := New(
		WithField("defined_key", "defined_value"),
	)
	if err != nil {
		b.Fatal(err)
	}

	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

}
