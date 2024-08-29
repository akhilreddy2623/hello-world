package logging

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Logger API
type Logger interface {

	// Info messages logging. There should preferably always be a valid context to be passed. Pass nil but that is discouraged because you will lose
	// tracing context. 'message' can be simple string message with or without format strings. 'message' can also be ComplexMessage if you want a text message
	// wit supporting metadata. 'message' can also be any Go plain object, although that should be avoided. 'args' should be argument that you want replaced in format strings
	// of the message text.
	Info(ctx context.Context, message any, args ...any)

	// Warn messages logging. There should preferably always be a valid context to be passed. Pass nil but that is discouraged because you will lose
	// tracing context. 'message' can be simple string message with or without format strings. 'message' can also be ComplexMessage if you want a text message
	// wit supporting metadata. 'message' can also be any Go plain object, although that should be avoided. 'args' should be argument that you want replaced in format strings
	// of the message text.
	Warn(ctx context.Context, message any, args ...any)

	// Error messages logging. There should preferably always be a valid context to be passed. Pass nil but that is discouraged because you will lose
	// tracing context. 'message' can be simple string message with or without format strings. 'message' can also be ComplexMessage if you want a text message
	// wit supporting metadata. 'message' can also be any Go plain object, although that should be avoided. 'args' should be argument that you want replaced in format strings
	// of the message text. Pass error as nil if there is no error
	Error(ctx context.Context, error error, message any, args ...any)
}

// Factory method to create a logger
func GetLogger(name string) Logger {
	zlog := zerolog.
		New(os.Stdout).
		Hook(logGuidHook{}).
		Hook(threadNameHook{}).
		With().
		Timestamp().
		Str("LoggerName", name).
		Logger()

	stdoutlog := stdoutJsonLogger{zlog: zlog}

	return &stdoutlog
}

type ComplexMessage struct {
	message          string
	messageObject    any
	placeholderProps map[string]any
}

func NewMessage(message string, messageObject any, args ...any) *ComplexMessage {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	msg := &ComplexMessage{message: message, messageObject: messageObject, placeholderProps: make(map[string]any)}

	return msg
}

func (complexMesssage *ComplexMessage) Prop(name string, value any) *ComplexMessage {
	complexMesssage.placeholderProps[name] = value
	complexMesssage.messageObject = &complexMesssage.placeholderProps
	return complexMesssage
}

// initilization
func init() {
	zerolog.TimestampFieldName = "Timestamp"
	zerolog.LevelFieldName = "Level"
	zerolog.MessageFieldName = "Message"
	zerolog.CallerFieldName = "Caller"
	zerolog.ErrorFieldName = "Error"
}

// Custom Hooks
type logGuidHook struct{}

func (h logGuidHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("LogGuid", uuid.New().String())
}

type threadNameHook struct{}

func (h threadNameHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx != nil {
		threadName := ctx.Value("ThreadName")
		e.Any("ThreadName", threadName)
	}
}

// Standard Output Logger
type stdoutJsonLogger struct {
	zlog zerolog.Logger
}

func (logger stdoutJsonLogger) Info(ctx context.Context, message any, args ...any) {
	ctx, messageText, messageObject := extractLogParams(ctx, message)

	log := logger.
		zlog.
		With().
		Ctx(ctx).
		Interface("MessageObject", messageObject).
		Logger()

	if len(args) <= 0 {
		log.Info().Msg(messageText)
	} else {
		log.Info().Msgf(messageText, args...)
	}
}

func (logger stdoutJsonLogger) Warn(ctx context.Context, message any, args ...any) {
	ctx, messageText, messageObject := extractLogParams(ctx, message)

	log := logger.
		zlog.
		With().
		Ctx(ctx).
		Interface("MessageObject", messageObject).
		Logger()

	if len(args) <= 0 {
		log.Warn().Msg(messageText)
	} else {
		log.Warn().Msgf(messageText, args...)
	}
}

func (logger stdoutJsonLogger) Error(ctx context.Context, err error, message any, args ...any) {
	ctx, messageText, messageObject := extractLogParams(ctx, message)

	log := logger.
		zlog.
		With().
		Ctx(ctx).
		Interface("MessageObject", messageObject).
		Logger()

	if len(args) <= 0 {
		log.Error().Err(err).Msg(messageText)
	} else {
		log.Error().Err(err).Msgf(messageText, args...)
	}
}

func extractLogParams(ctx context.Context, message any) (context.Context, string, any) {
	var messageObject any = nil
	var messageText string = ""

	switch messageVal := message.(type) {
	case ComplexMessage:
		messageObject = messageVal.messageObject
		messageText = messageVal.message
	case *ComplexMessage:
		messageObject = messageVal.messageObject
		messageText = messageVal.message
	case string:
		messageText = messageVal
	case error:
		messageText = messageVal.Error()

	default:
		messageObject = message
		if stringer, ok := message.(fmt.Stringer); ok {
			messageText = stringer.String()
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return ctx, messageText, messageObject
}
