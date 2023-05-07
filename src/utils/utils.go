package utils

import (
	"os"

	"github.com/rs/zerolog/log"
)

func CheckError(err error) {
	if err != nil {
		log.Error().Err(err).Msg("Fatal error")
		os.Exit(1)
	}
}
