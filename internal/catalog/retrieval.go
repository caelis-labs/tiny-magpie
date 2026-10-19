package catalog

import "strings"

// RetrievalKind recognizes the APIs vendors already list. Name matching is
// the fallback for OpenAI-shaped catalogs that give only model IDs.
func RetrievalKind(id string, endpoints []string) string {
	for _, endpoint := range endpoints {
		switch strings.Trim(strings.TrimPrefix(endpoint, "/v1"), "/") {
		case "embeddings":
			return "embedding"
		case "rerank":
			return "rerank"
		}
	}
	id = strings.ToLower(id)
	if strings.Contains(id, "embed") {
		return "embedding"
	}
	if strings.Contains(id, "rerank") {
		return "rerank"
	}
	return ""
}
