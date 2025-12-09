package allure_report

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lamoda/gonkey/models"
)

// mockInfo contains information about a mock service
type mockInfo struct {
	ServiceName   string
	Strategy      string
	Endpoints     []string
	EndpointCount int // для краткого отображения в Setup
	StepsCount    int // для sequence
}

// extractMockInfo extracts structured information about mock services
func extractMocksInfo(mocks map[string]interface{}) []mockInfo {
	if len(mocks) == 0 {
		return nil
	}

	var infos []mockInfo
	for serviceName, def := range mocks {
		info := mockInfo{
			ServiceName: serviceName,
		}

		defMap, ok := def.(map[interface{}]interface{})
		if !ok {
			info.Strategy = "unknown"
			infos = append(infos, info)
			continue
		}

		if s, ok := defMap["strategy"]; ok {
			if strategyStr, ok := s.(string); ok {
				info.Strategy = strategyStr
				info.Endpoints = extractEndpoints(strategyStr, defMap)
				info.EndpointCount = len(info.Endpoints)

				if strategyStr == "sequence" {
					if seq, ok := defMap["sequence"].([]interface{}); ok {
						info.StepsCount = len(seq)
					}
				}
			}
		}

		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ServiceName < infos[j].ServiceName
	})

	return infos
}

// extractEndpoints extracts endpoint information based on strategy
func extractEndpoints(strategy string, def map[interface{}]interface{}) []string {
	var endpoints []string

	switch strategy {
	case "constant", "template":
		if path, ok := def["path"].(string); ok {
			endpoints = []string{path}
		} else {
			endpoints = []string{"/"}
		}

	case "file":
		if filename, ok := def["filename"].(string); ok {
			endpoints = []string{fmt.Sprintf("file: %s", filepath.Base(filename))}
		}

	case "uriVary":
		if uris, ok := def["uris"].(map[interface{}]interface{}); ok {
			for uri := range uris {
				if uriStr, ok := uri.(string); ok {
					endpoints = append(endpoints, uriStr)
				}
			}
			sort.Strings(endpoints)
		}

	case "methodVary":
		if methods, ok := def["methods"].(map[interface{}]interface{}); ok {
			for method := range methods {
				if methodStr, ok := method.(string); ok {
					endpoints = append(endpoints, methodStr)
				}
			}
			sort.Strings(endpoints)
		}

	case "basedOnRequest":
		if uris, ok := def["uris"].([]interface{}); ok {
			for i := range uris {
				endpoints = append(endpoints, fmt.Sprintf("case %d", i))
			}
		}

	case "sequence":
		if seq, ok := def["sequence"].([]interface{}); ok {
			for i := range seq {
				endpoints = append(endpoints, fmt.Sprintf("step %d", i))
			}
		}

	case "nop", "dropRequest":
		endpoints = []string{"-"}
	}

	return endpoints
}

// formatFixturesInfo formats fixture information for display
func formatFixturesInfo(fixtures []string, multiDb models.FixturesMultiDb) string {
	if len(fixtures) > 0 {
		return fmt.Sprintf("%s (%d files)", strings.Join(fixtures, ", "), len(fixtures))
	}

	if len(multiDb) > 0 {
		var parts []string
		for _, fixture := range multiDb {
			parts = append(parts, fmt.Sprintf("%s: %d files", fixture.DbName, len(fixture.Files)))
		}
		return strings.Join(parts, "; ")
	}

	return ""
}

func formatMockBrief(info mockInfo) string {
	switch info.Strategy {
	case "sequence":
		if info.StepsCount > 0 {
			return fmt.Sprintf("%s (%s, %d steps)", info.ServiceName, info.Strategy, info.StepsCount)
		}
		return fmt.Sprintf("%s (%s)", info.ServiceName, info.Strategy)

	case "uriVary":
		if info.EndpointCount > 0 {
			return fmt.Sprintf("%s (%s, %d endpoints)", info.ServiceName, info.Strategy, info.EndpointCount)
		}
		return fmt.Sprintf("%s (%s)", info.ServiceName, info.Strategy)

	case "methodVary":
		if info.EndpointCount > 0 {
			return fmt.Sprintf("%s (%s, %d methods)", info.ServiceName, info.Strategy, info.EndpointCount)
		}
		return fmt.Sprintf("%s (%s)", info.ServiceName, info.Strategy)

	case "basedOnRequest":
		if info.EndpointCount > 0 {
			return fmt.Sprintf("%s (%s, %d cases)", info.ServiceName, info.Strategy, info.EndpointCount)
		}
		return fmt.Sprintf("%s (%s)", info.ServiceName, info.Strategy)

	default:
		// constant, template, file, nop, dropRequest
		return fmt.Sprintf("%s (%s)", info.ServiceName, info.Strategy)
	}
}
