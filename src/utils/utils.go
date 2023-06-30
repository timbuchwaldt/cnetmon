package utils

import (
	"cnetmon/structs"
	"os"

	"github.com/rs/zerolog/log"
)

func CheckErrorFatal(err error) {
	if err != nil {
		log.Error().Err(err).Msg("Fatal error")
		os.Exit(1)
	}
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func IPTargetInSlice(a structs.Target, list []structs.Target) bool {
	for _, b := range list {
		if b.IP == a.IP {
			return true
		}
	}
	return false
}
