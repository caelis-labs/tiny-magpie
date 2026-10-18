package provider

import (
	"fmt"

	"strings"
)

func countsLine(cs []PartnerCount) string {
	var out []string
	for _, c := range cs {
		out = append(out, fmt.Sprintf("%s %s %d", c.ID, c.What, c.Count))
	}
	return strings.Join(out, ",")
}
