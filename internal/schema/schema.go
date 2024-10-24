package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lus/kratos-readonly-traits/internal/static"
	"github.com/rs/zerolog/log"
)

func ExtractReadOnlyTraits(url string) (map[string]bool, error) {
	// Retrieve the schema data
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Failed to get schema: %s\nWith error: %s", url, err.Error()))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}

	// Extract the traits
	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		return nil, err
	}
	traits, ok := extractNestedValue[map[string]any](schema, "properties.traits.properties")
	if !ok {
		return nil, errors.New("traits object missing from schema")
	}

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

func AppendDotToDomain(rawURL string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return "", err
	}

	// Append a dot to the host
	parsedURL.Host += "."

	// Rebuild the modified URL
	return parsedURL.String(), nil
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
