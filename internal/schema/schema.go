package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lus/kratos-readonly-traits/internal/static"
	"github.com/rs/zerolog/log"
)

func ExtractReadOnlyTraits(url string) (map[string]bool, error) {
	// Retrieve the schema data
	log.Info().Msg(fmt.Sprintf("Getting schema: %s", url))
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to get schema with error: %s", err.Error()))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Error().Msg(fmt.Sprintf("Failed to get schema (%d) with body: %s", resp.StatusCode, string(body[:])))
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}

	// Extract the traits
	log.Info().Msg("Extracting traits...")
	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, err
	}
	traits, ok := extractNestedValue[map[string]any](schema, "properties.traits.properties")
	if !ok {
		return nil, errors.New("traits object missing from schema")
	}

	log.Info().Msg(fmt.Sprintf("Traits length: %d", len(traits)))
	// Extract the readonly state for every trait
	traitStates := make(map[string]bool, len(traits))
	for trait, rawValues := range traits {
		values, ok := rawValues.(map[string]any)
		if !ok {
			traitStates[trait] = false
			continue
		}
		readonly, _ := extractNestedValue[bool](values, static.IdentitySchemaExtensionKey+".readonly")
		traitStates[trait] = readonly
	}
	return traitStates, nil
}

func extractNestedValue[T any](structure map[string]any, key string) (T, bool) {
	var defaultValue T
	keys := strings.Split(key, ".")
	currentMap := structure
	for i := 0; i < len(keys)-1; i++ {
		newMap, ok := currentMap[keys[i]].(map[string]any)
		if !ok {
			return defaultValue, ok
		}
		currentMap = newMap
	}
	val, ok := currentMap[keys[len(keys)-1]].(T)
	return val, ok
}
