package provider

import "github.com/yetone/magpie/internal/catalog"

// RetrievalModels uses the vendor catalog already fetched by upstream and
// the plan's explicit retrieval models. It never adds retrieval models to chat
// picks or automatically mixes their vector spaces in groups.
func (p Provider) RetrievalModels() []catalog.Model {
	var out []catalog.Model
	seen := map[string]bool{}
	add := func(m catalog.Model) {
		kind := m.Serves
		if kind == "" {
			kind = catalog.RetrievalKind(m.ID, nil)
		}
		if (kind != catalog.ServesEmbedding && kind != catalog.ServesRerank) || seen[m.ID] {
			return
		}
		seen[m.ID], m.Serves = true, kind
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
	for _, m := range p.PlanRetrieval("") {
		add(m)
	}
	for _, id := range p.Models {
		add(catalog.Model{ID: id, Name: id})
	}
	return out
}
