package logging

import (
  "github.com/rs/zerolog"
  "os"
  "testing"
)

func TestNew(t *testing.T) {
  logger := New("test-logger")
  logger.Info().Msg("test message")
}

func TestSetupLogger(t *testing.T) {
  defer SetupLogger("DEBUG")

  t.Log("DEBUG")
  {
    SetupLogger("DEBUG")
    if zerolog.GlobalLevel() != zerolog.DebugLevel {
      t.Errorf("failed to set DEBUG log level")
    }
  }

  t.Log("INFO")
  {
    SetupLogger("INFO")
    if zerolog.GlobalLevel() != zerolog.InfoLevel {
      t.Errorf("failed to set INFO log level")
    }
  }

  t.Log("ERROR")
  {
    SetupLogger("ERROR")
    if zerolog.GlobalLevel() != zerolog.ErrorLevel {
      t.Errorf("failed to set ERROR log level")
    }
  }

  t.Log("UNKNOWN")
  {
    SetupLogger("UNKNOWN")
    if zerolog.GlobalLevel() != zerolog.InfoLevel {
      t.Errorf("failed to set fallback INFO log level")
    }
  }

}
