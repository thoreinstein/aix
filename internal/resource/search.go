package resource

import (
	"slices"
	"strings"
)

// SearchOptions configures resource search filtering.
type SearchOptions struct {
	// Type filters by resource type. Empty string matches all types.
	Type ResourceType
	// RepoName filters by repository name. Empty string matches all repos.
	RepoName string
}

// Search finds resources matching the query and filter options.
// Matching is case-insensitive against Name and Description fields.
// An empty query returns all resources (subject to filters).
// Results are sorted by match quality (exact name > prefix > contains > description-only).
func Search(resources []Resource, query string, opts SearchOptions) []Resource {
	query = strings.ToLower(query)

	var results []Resource
	for _, r := range resources {
		if !matchesFilters(r, opts) {
			continue
		}
		if query == "" || matchesQuery(r, query) {
			results = append(results, r)
		}
	}

	// Sort by score descending (higher score = better match)
	slices.SortFunc(results, func(a, b Resource) int {
		scoreA := scoreMatch(a, query)
		scoreB := scoreMatch(b, query)
		// Descending order: higher score first
		return scoreB - scoreA
	})

	return results
}

// matchesFilters checks if a resource passes the filter criteria.
func matchesFilters(r Resource, opts SearchOptions) bool {
	if opts.Type != "" && r.Type != opts.Type {
		return false
	}
	if opts.RepoName != "" && r.RepoName != opts.RepoName {
		return false
	}
	return true
}

// matchesQuery checks if a resource matches the search query.
// Matching is case-insensitive substring matching against Name and Description.
func matchesQuery(r Resource, query string) bool {
	name := strings.ToLower(r.Name)
	desc := strings.ToLower(r.Description)
	return strings.Contains(name, query) || strings.Contains(desc, query)
}

// scoreMatch returns a score indicating match quality.
// Higher scores indicate better matches.
//
// Scoring:
//   - 100: Exact name match
//   - 75: Name starts with query (prefix match)
//   - 50: Name contains query
//   - 25: Description contains query (but name doesn't)
//   - 0: No match or empty query
func scoreMatch(r Resource, query string) int {
	if query == "" {
		return 0
	}

	name := strings.ToLower(r.Name)
	desc := strings.ToLower(r.Description)

	// Exact name match scores highest
	if name == query {
		return 100
	}

	// Name prefix match scores high
	if strings.HasPrefix(name, query) {
		return 75
	}

	// Name contains query scores medium
	if strings.Contains(name, query) {
		return 50
	}

	// Description-only match scores lowest
	if strings.Contains(desc, query) {
		return 25
	}

	return 0
}
