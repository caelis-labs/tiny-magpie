package catalog

import "testing"

func TestRetrievalModelsStayOutOfChat(t *testing.T) {
	for _, c := range []struct {
		id        string
		endpoints []string
		kind      string
	}{{"text-embedding-3-small", nil, "embedding"}, {"bge-reranker-v2", nil, "rerank"}, {"private-vector", []string{"/v1/embeddings"}, "embedding"}, {"rank", []string{"/rerank"}, "rerank"}, {"chat", []string{"/chat/completions"}, ""}} {
		if got := RetrievalKind(c.id, c.endpoints); got != c.kind {
			t.Errorf("%s: %q", c.id, got)
		}
	}
	got := Chat([]Model{{ID: "chat"}, {ID: "vector", Retrieval: "embedding"}, {ID: "rank", Retrieval: "rerank"}})
	if len(got) != 1 || got[0].ID != "chat" {
		t.Fatalf("chat catalog includes retrieval: %+v", got)
	}
}
