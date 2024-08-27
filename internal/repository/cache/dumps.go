package cache

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
)

func makeDump(filename string, pull any) {
	data, err := json.Marshal(pull)
	if err != nil {
		log.Error().
			Err(err).
			Str("filename", filename).
			Msg("can not marshall pull")
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Error().
			Err(err).
			Str("filename", filename).
			Msg("can not write pull")
	}
	log.Info().
		Str("filename", filename).
		Msg("dump saved successfully")
}
func loadFromDump(filename string, pull any) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &pull)
}
