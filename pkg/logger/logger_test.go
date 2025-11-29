package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/gopybara/httpbara"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"msg"`
	TS      string `json:"ts"`
}

func setupTestLogger() (*ZapLogger, *bytes.Buffer) {
	var buf bytes.Buffer

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	return &ZapLogger{log: zapLogger}, &buf
}

func parseLogEntry(t *testing.T, buf *bytes.Buffer) LogEntry {
	var entry LogEntry
	err := json.Unmarshal(buf.Bytes(), &entry)
	assert.NoError(t, err)
	return entry
}

func parseLogEntries(t *testing.T, buf *bytes.Buffer) []LogEntry {
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	entries := make([]LogEntry, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry LogEntry
		err := json.Unmarshal([]byte(line), &entry)
		assert.NoError(t, err)
		entries = append(entries, entry)
	}

	return entries
}

func TestZapLoggerImplementsInterface(t *testing.T) {
	logger, _ := setupTestLogger()

	var _ httpbara.Logger = logger

	assert.NotNil(t, logger)
	assert.Implements(t, (*httpbara.Logger)(nil), logger)
}

func TestNewZapLogger(t *testing.T) {
	zapLogger, _ := zap.NewProduction()

	logger := NewZapLogger(zapLogger)

	assert.NotNil(t, logger)
	assert.IsType(t, &ZapLogger{}, logger)

	zapLoggerImpl, ok := logger.(*ZapLogger)
	assert.True(t, ok)
	assert.Equal(t, zapLogger, zapLoggerImpl.log)
}

func TestZapLogger_Info(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		args     []any
		expected map[string]any
	}{
		{
			name:    "simple message",
			message: "test info message",
			args:    nil,
			expected: map[string]any{
				"level": "info",
				"msg":   "test info message",
			},
		},
		{
			name:    "with string field",
			message: "user action",
			args:    []any{"user_id", "12345"},
			expected: map[string]any{
				"level":   "info",
				"msg":     "user action",
				"user_id": "12345",
			},
		},
		{
			name:    "with multiple fields",
			message: "operation completed",
			args:    []any{"operation", "create_user", "duration", 150, "success", true},
			expected: map[string]any{
				"level":     "info",
				"msg":       "operation completed",
				"operation": "create_user",
				"duration":  float64(150),
				"success":   true,
			},
		},
		{
			name:    "with int64 field",
			message: "timestamp logged",
			args:    []any{"timestamp", int64(1640995200)},
			expected: map[string]any{
				"level":     "info",
				"msg":       "timestamp logged",
				"timestamp": float64(1640995200),
			},
		},
		{
			name:    "with uint field",
			message: "count updated",
			args:    []any{"count", uint(42)},
			expected: map[string]any{
				"level": "info",
				"msg":   "count updated",
				"count": float64(42),
			},
		},
		{
			name:    "with float64 field",
			message: "price set",
			args:    []any{"price", 99.99},
			expected: map[string]any{
				"level": "info",
				"msg":   "price set",
				"price": 99.99,
			},
		},
		{
			name:    "with error field",
			message: "error occurred",
			args:    []any{errors.New("test error")},
			expected: map[string]any{
				"level": "info",
				"msg":   "error occurred",
				"error": "test error",
			},
		},
		{
			name:    "with zap field",
			message: "custom field",
			args:    []any{zap.String("custom", "value")},
			expected: map[string]any{
				"level":  "info",
				"msg":    "custom field",
				"custom": "value",
			},
		},
		{
			name:    "with mixed field types",
			message: "complex log",
			args:    []any{"string_field", "value", zap.Int("zap_field", 100), "bool_field", true},
			expected: map[string]any{
				"level":        "info",
				"msg":          "complex log",
				"string_field": "value",
				"zap_field":    float64(100),
				"bool_field":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := setupTestLogger()

			logger.Info(tt.message, tt.args...)

			var logData map[string]any
			err := json.Unmarshal(buf.Bytes(), &logData)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expected {
				actualValue, exists := logData[key]
				assert.True(t, exists, "field %s should exist", key)
				assert.Equal(t, expectedValue, actualValue, "field %s should match", key)
			}
		})
	}
}

func TestZapLogger_Debug(t *testing.T) {
	logger, buf := setupTestLogger()

	logger.Debug("debug message", "key", "value")

	entry := parseLogEntry(t, buf)
	assert.Equal(t, "debug", entry.Level)
	assert.Equal(t, "debug message", entry.Message)

	var logData map[string]any
	err := json.Unmarshal(buf.Bytes(), &logData)
	assert.NoError(t, err)
	assert.Equal(t, "value", logData["key"])
}

func TestZapLogger_Error(t *testing.T) {
	logger, buf := setupTestLogger()

	testErr := errors.New("test error")
	logger.Error("error occurred", "error_type", "validation", testErr)

	entry := parseLogEntry(t, buf)
	assert.Equal(t, "error", entry.Level)
	assert.Equal(t, "error occurred", entry.Message)

	var logData map[string]any
	err := json.Unmarshal(buf.Bytes(), &logData)
	assert.NoError(t, err)
	assert.Equal(t, "validation", logData["error_type"])
	assert.Equal(t, "test error", logData["error"])
}

func TestZapLogger_Warn(t *testing.T) {
	logger, buf := setupTestLogger()

	logger.Warn("warning message", "reason", "deprecated_api")

	entry := parseLogEntry(t, buf)
	assert.Equal(t, "warn", entry.Level)
	assert.Equal(t, "warning message", entry.Message)

	var logData map[string]any
	err := json.Unmarshal(buf.Bytes(), &logData)
	assert.NoError(t, err)
	assert.Equal(t, "deprecated_api", logData["reason"])
}

func TestZapLogger_Panic(t *testing.T) {
	logger, buf := setupTestLogger()

	assert.Panics(t, func() {
		logger.Panic("panic message", "cause", "critical_error")
	})

	entry := parseLogEntry(t, buf)
	assert.Equal(t, "panic", entry.Level)
	assert.Equal(t, "panic message", entry.Message)

	var logData map[string]any
	err := json.Unmarshal(buf.Bytes(), &logData)
	assert.NoError(t, err)
	assert.Equal(t, "critical_error", logData["cause"])
}

func TestMapFieldsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		testFunc func(*testing.T, *ZapLogger, *bytes.Buffer)
	}{
		{
			name: "empty args",
			args: []any{},
			testFunc: func(t *testing.T, logger *ZapLogger, buf *bytes.Buffer) {
				var logData map[string]any
				err := json.Unmarshal(buf.Bytes(), &logData)
				assert.NoError(t, err)
				assert.Equal(t, "info", logData["level"])
				assert.Equal(t, "test", logData["msg"])
			},
		},
		{
			name: "odd number of args (missing value)",
			args: []any{"key"},
			testFunc: func(t *testing.T, logger *ZapLogger, buf *bytes.Buffer) {
				var logData map[string]any
				err := json.Unmarshal(buf.Bytes(), &logData)
				assert.NoError(t, err)
				_, exists := logData["key"]
				assert.False(t, exists, "incomplete key-value pair should not be logged")
			},
		},
		{
			name: "nil value",
			args: []any{"null_field", nil},
			testFunc: func(t *testing.T, logger *ZapLogger, buf *bytes.Buffer) {
				var logData map[string]any
				err := json.Unmarshal(buf.Bytes(), &logData)
				assert.NoError(t, err)
				assert.Nil(t, logData["null_field"])
			},
		},
		{
			name: "complex type",
			args: []any{"complex", map[string]int{"nested": 42}},
			testFunc: func(t *testing.T, logger *ZapLogger, buf *bytes.Buffer) {
				var logData map[string]any
				err := json.Unmarshal(buf.Bytes(), &logData)
				assert.NoError(t, err)
				complexField, exists := logData["complex"]
				assert.True(t, exists)
				assert.NotNil(t, complexField)
			},
		},
		{
			name: "multiple zap fields",
			args: []any{
				zap.String("field1", "value1"),
				zap.Int("field2", 42),
				zap.Bool("field3", true),
			},
			testFunc: func(t *testing.T, logger *ZapLogger, buf *bytes.Buffer) {
				var logData map[string]any
				err := json.Unmarshal(buf.Bytes(), &logData)
				assert.NoError(t, err)
				assert.Equal(t, "value1", logData["field1"])
				assert.Equal(t, float64(42), logData["field2"])
				assert.Equal(t, true, logData["field3"])
			},
		},
		{
			name: "mixed zap and key-value",
			args: []any{
				zap.String("zap_field", "zap_value"),
				"kv_field", "kv_value",
				zap.Int("another_zap", 123),
			},
			testFunc: func(t *testing.T, logger *ZapLogger, buf *bytes.Buffer) {
				var logData map[string]any
				err := json.Unmarshal(buf.Bytes(), &logData)
				assert.NoError(t, err)
				assert.Equal(t, "zap_value", logData["zap_field"])
				assert.Equal(t, "kv_value", logData["kv_field"])
				assert.Equal(t, float64(123), logData["another_zap"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := setupTestLogger()

			logger.Info("test", tt.args...)

			tt.testFunc(t, logger, buf)
		})
	}
}

func TestMapFieldsTypeConversions(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		expected any
	}{
		{"string", "str_field", "test_string", "test_string"},
		{"int", "int_field", 42, float64(42)},
		{"int64", "int64_field", int64(9223372036854775807), float64(9223372036854775807)},
		{"uint", "uint_field", uint(123), float64(123)},
		{"float64", "float_field", 3.14159, 3.14159},
		{"bool", "bool_field", true, true},
		{"error", "error", errors.New("test error"), "test error"},
		{"nil", "nil_field", nil, nil},
		{"slice", "slice_field", []string{"a", "b", "c"}, []any{"a", "b", "c"}},
		{"map", "map_field", map[string]int{"key": 1}, map[string]any{"key": float64(1)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := setupTestLogger()

			if tt.key == "error" {
				logger.Info("test", tt.value)
			} else {
				logger.Info("test", tt.key, tt.value)
			}

			var logData map[string]any
			err := json.Unmarshal(buf.Bytes(), &logData)
			assert.NoError(t, err)

			if tt.key == "error" {
				assert.Equal(t, tt.expected, logData["error"])
			} else {
				assert.Equal(t, tt.expected, logData[tt.key])
			}
		})
	}
}

func TestLoggerWithDifferentLevels(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:     "ts",
		LevelKey:    "level",
		MessageKey:  "msg",
		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	})

	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	zapLogger := zap.New(core)
	logger := &ZapLogger{log: zapLogger}

	logger.Debug("debug message")
	logger.Info("info message")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	assert.Len(t, lines, 1, "only info level should be logged")

	var logData map[string]any
	err := json.Unmarshal([]byte(lines[0]), &logData)
	assert.NoError(t, err)
	assert.Equal(t, "info", logData["level"])
	assert.Equal(t, "info message", logData["msg"])
}

func TestConcurrentLogging(t *testing.T) {
	logger, buf := setupTestLogger()

	done := make(chan bool)
	numGoroutines := 10
	logsPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Info("concurrent log", "goroutine", id, "iteration", j)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	entries := parseLogEntries(t, buf)
	assert.GreaterOrEqual(t, len(entries), numGoroutines*logsPerGoroutine-5)
	assert.LessOrEqual(t, len(entries), numGoroutines*logsPerGoroutine)

	for _, entry := range entries {
		assert.Equal(t, "info", entry.Level)
		assert.Equal(t, "concurrent log", entry.Message)
	}
}

func TestLargeLogMessage(t *testing.T) {
	logger, buf := setupTestLogger()

	longMessage := strings.Repeat("A", 10000)
	longValue := strings.Repeat("B", 5000)

	logger.Info(longMessage, "large_field", longValue)

	entry := parseLogEntry(t, buf)
	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, longMessage, entry.Message)

	var logData map[string]any
	err := json.Unmarshal(buf.Bytes(), &logData)
	assert.NoError(t, err)
	assert.Equal(t, longValue, logData["large_field"])
}

func TestSpecialCharacters(t *testing.T) {
	logger, buf := setupTestLogger()

	specialChars := "Special chars: \n\t\r\"'\\/"
	unicodeText := "Unicode: 🚀 测试 नमस्ते"

	logger.Info("special message", "special", specialChars, "unicode", unicodeText)

	var logData map[string]any
	err := json.Unmarshal(buf.Bytes(), &logData)
	assert.NoError(t, err)

	assert.Equal(t, specialChars, logData["special"])
	assert.Equal(t, unicodeText, logData["unicode"])
}

func BenchmarkZapLogger_Info(b *testing.B) {
	logger, _ := setupTestLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "test", "value")
	}
}

func BenchmarkZapLogger_InfoWithManyFields(b *testing.B) {
	testLogger, _ := setupTestLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLogger.Info("benchmark message",
			"field1", "value1",
			"field2", i,
			"field3", 3.14,
			"field4", true,
			"field5", "value5",
			"field6", int64(123456789),
			"field7", uint(987654321),
			"field8", "value8",
		)
	}
}

func BenchmarkMapFields(b *testing.B) {
	testLogger, _ := setupTestLogger()
	args := []any{
		"string", "value",
		"int", 42,
		"float", 3.14,
		"bool", true,
		zap.String("zap_field", "zap_value"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = testLogger.mapFields(args...)
	}
}

func TestNewZapInstanceWithLifecycle(t *testing.T) {
	lc := &mockLifecycle{}

	zapLogger, err := NewZapInstance(lc)

	assert.NoError(t, err)
	assert.NotNil(t, zapLogger)
	assert.Len(t, lc.hooks, 1)

	hook := lc.hooks[0]
	err = hook.OnStop(context.Background())
	assert.NoError(t, err)
}

type mockLifecycle struct {
	hooks []fx.Hook
}

func (m *mockLifecycle) Append(hook fx.Hook) {
	m.hooks = append(m.hooks, hook)
}

func TestModules(t *testing.T) {
	assert.NotNil(t, ZapModule)
	assert.NotNil(t, HttpbaraLoggerModule)
	assert.NotNil(t, Module)
	assert.Equal(t, ZapModule, Module)
}

func TestLoggerIntegration(t *testing.T) {
	lc := &mockLifecycle{}

	zapLogger, err := NewZapInstance(lc)
	assert.NoError(t, err)

	httpbaraLogger := NewZapLogger(zapLogger)
	assert.NotNil(t, httpbaraLogger)

	var _ httpbara.Logger = httpbaraLogger
}
