package queue

import "github.com/rs/zerolog/log"

// BackliteLogger adapts zerolog to backlite.Logger
type BackliteLogger struct{}

func (l *BackliteLogger) Debug(msg string, args ...any) {
	log.Debug().Fields(args).Msg(msg)
}

func (l *BackliteLogger) Info(msg string, args ...any) {
	log.Info().Fields(args).Msg(msg)
}

func (l *BackliteLogger) Warn(msg string, args ...any) {
	log.Warn().Fields(args).Msg(msg)
}

func (l *BackliteLogger) Error(msg string, args ...any) {
	log.Error().Fields(args).Msg(msg)
}
