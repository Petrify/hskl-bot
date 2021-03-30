package schooldiscord

import "github.com/sahilm/fuzzy"

type moduleList []modelFinalSearchable

func (lst moduleList) Len() int {
	return len(lst)
}

func (lst moduleList) String(i int) string {
	//concatenate abbreviation and name for fuzzy searching algo
	return lst[i].name + " " + lst[i].abbr
}

func fuzzySearch(list []modelFinalSearchable, pattern string, max int) []modelFinalSearchable {
	lst := moduleList(list)
	matches := fuzzy.FindFrom(pattern, lst)
	n := min(matches.Len(), max)

	out := make([]modelFinalSearchable, n)
	for i := 0; i < n; i++ {
		out[i] = list[matches[i].Index]
	}

	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
