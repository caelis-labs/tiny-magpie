package provider

import "github.com/yetone/magpie/internal/catalog"

// RetrievalModels uses the vendor catalog already fetched by upstream and
// the plan's explicit embeddings. It never adds retrieval models to chat
// picks or automatically mixes their vector spaces in groups.
func (p Provider) RetrievalModels() []catalog.Model {
	var out []catalog.Model
	seen := map[string]bool{}
	add := func(m catalog.Model) {
		kind := m.Retrieval
		if kind == "" {
			kind = catalog.RetrievalKind(m.ID, nil)
		}
		if kind == "" || seen[m.ID] {
			return
		}
		seen[m.ID], m.Retrieval = true, kind
		out = append(out, m)
	}
	if p.Account != nil {
		for _, m := range p.Available() {
			add(m)
		}
	}
	if ms, ok := p.liveAll(); ok {
		for _, m := range ms {
			add(m)
		}
	}
	for _, m := range p.PlanEmbeddings() {
		add(m)
	}
	for _, id := range p.Models {
		add(catalog.Model{ID: id, Name: id})
	}
	return out
}
