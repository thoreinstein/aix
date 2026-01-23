package resource

import "testing"

// testResources returns a slice of test resources covering all types and multiple repos.
func testResources() []Resource {
	return []Resource{
		{Name: "code-review", Description: "Reviews code for issues", Type: TypeSkill, RepoName: "official"},
		{Name: "deploy", Description: "Deploy to production", Type: TypeCommand, RepoName: "community"},
		{Name: "security-guru", Description: "Security code review", Type: TypeAgent, RepoName: "official"},
		{Name: "github", Description: "GitHub MCP server", Type: TypeMCP, RepoName: "official"},
		{Name: "test-runner", Description: "Runs automated tests", Type: TypeSkill, RepoName: "community"},
		{Name: "codegen", Description: "Generates boilerplate code", Type: TypeCommand, RepoName: "official"},
		{Name: "reviewer", Description: "General code reviewer agent", Type: TypeAgent, RepoName: "community"},
		{Name: "database", Description: "Database MCP server", Type: TypeMCP, RepoName: "community"},
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	resources := testResources()

	tests := []struct {
		query       string
		wantMatches int
	}{
		{query: "CODE", wantMatches: 4},     // matches code-review, codegen, security-guru (desc), reviewer (desc)
		{query: "code", wantMatches: 4},     // same matches
		{query: "Code", wantMatches: 4},     // same matches
		{query: "DEPLOY", wantMatches: 1},   // matches deploy
		{query: "GITHUB", wantMatches: 1},   // matches github
		{query: "Review", wantMatches: 3},   // matches code-review, security-guru (desc), reviewer
		{query: "SECURITY", wantMatches: 1}, // matches security-guru
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := Search(resources, tt.query, SearchOptions{})
			if len(results) != tt.wantMatches {
				names := make([]string, len(results))
				for i, r := range results {
					names[i] = r.Name
				}
				t.Errorf("Search(%q) = %d results %v, want %d", tt.query, len(results), names, tt.wantMatches)
			}
		})
	}
}

func TestSearch_MatchesName(t *testing.T) {
	resources := testResources()

	tests := []struct {
		query string
		want  string
	}{
		{query: "code-review", want: "code-review"},
		{query: "deploy", want: "deploy"},
		{query: "github", want: "github"},
		{query: "security-guru", want: "security-guru"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := Search(resources, tt.query, SearchOptions{})
			if len(results) == 0 {
				t.Fatalf("Search(%q) returned no results", tt.query)
			}
			// Exact name match should be first
			if results[0].Name != tt.want {
				t.Errorf("Search(%q) first result = %q, want %q", tt.query, results[0].Name, tt.want)
			}
		})
	}
}

func TestSearch_MatchesDescription(t *testing.T) {
	resources := testResources()

	tests := []struct {
		name      string
		query     string
		wantFirst string
		wantLen   int
	}{
		{
			name:      "query in description only",
			query:     "production",
			wantFirst: "deploy",
			wantLen:   1,
		},
		{
			name:      "automated in description",
			query:     "automated",
			wantFirst: "test-runner",
			wantLen:   1,
		},
		{
			name:      "boilerplate in description",
			query:     "boilerplate",
			wantFirst: "codegen",
			wantLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Search(resources, tt.query, SearchOptions{})
			if len(results) != tt.wantLen {
				t.Errorf("Search(%q) returned %d results, want %d", tt.query, len(results), tt.wantLen)
			}
			if len(results) > 0 && results[0].Name != tt.wantFirst {
				t.Errorf("Search(%q) first result = %q, want %q", tt.query, results[0].Name, tt.wantFirst)
			}
		})
	}
}

func TestSearch_TypeFilter(t *testing.T) {
	resources := testResources()

	tests := []struct {
		name        string
		query       string
		filterType  ResourceType
		wantLen     int
		wantAllType ResourceType
	}{
		{
			name:        "filter skills only",
			query:       "",
			filterType:  TypeSkill,
			wantLen:     2,
			wantAllType: TypeSkill,
		},
		{
			name:        "filter commands only",
			query:       "",
			filterType:  TypeCommand,
			wantLen:     2,
			wantAllType: TypeCommand,
		},
		{
			name:        "filter agents only",
			query:       "",
			filterType:  TypeAgent,
			wantLen:     2,
			wantAllType: TypeAgent,
		},
		{
			name:        "filter mcp only",
			query:       "",
			filterType:  TypeMCP,
			wantLen:     2,
			wantAllType: TypeMCP,
		},
		{
			name:        "query with type filter",
			query:       "code",
			filterType:  TypeSkill,
			wantLen:     1,
			wantAllType: TypeSkill,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SearchOptions{Type: tt.filterType}
			results := Search(resources, tt.query, opts)

			if len(results) != tt.wantLen {
				t.Errorf("Search with type %q returned %d results, want %d", tt.filterType, len(results), tt.wantLen)
			}

			for _, r := range results {
				if r.Type != tt.wantAllType {
					t.Errorf("Search with type %q returned resource with type %q", tt.filterType, r.Type)
				}
			}
		})
	}
}

func TestSearch_RepoFilter(t *testing.T) {
	resources := testResources()

	tests := []struct {
		name        string
		query       string
		repoName    string
		wantLen     int
		wantAllRepo string
	}{
		{
			name:        "filter official repo",
			query:       "",
			repoName:    "official",
			wantLen:     4,
			wantAllRepo: "official",
		},
		{
			name:        "filter community repo",
			query:       "",
			repoName:    "community",
			wantLen:     4,
			wantAllRepo: "community",
		},
		{
			name:        "query with repo filter",
			query:       "review",
			repoName:    "official",
			wantLen:     2, // code-review and security-guru
			wantAllRepo: "official",
		},
		{
			name:        "nonexistent repo",
			query:       "",
			repoName:    "nonexistent",
			wantLen:     0,
			wantAllRepo: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SearchOptions{RepoName: tt.repoName}
			results := Search(resources, tt.query, opts)

			if len(results) != tt.wantLen {
				names := make([]string, len(results))
				for i, r := range results {
					names[i] = r.Name
				}
				t.Errorf("Search with repo %q returned %d results %v, want %d", tt.repoName, len(results), names, tt.wantLen)
			}

			for _, r := range results {
				if tt.wantAllRepo != "" && r.RepoName != tt.wantAllRepo {
					t.Errorf("Search with repo %q returned resource from repo %q", tt.repoName, r.RepoName)
				}
			}
		})
	}
}

func TestSearch_CombinedFilters(t *testing.T) {
	resources := testResources()

	tests := []struct {
		name     string
		query    string
		opts     SearchOptions
		wantLen  int
		wantName string // if wantLen == 1
	}{
		{
			name:     "type and repo filter - official skills",
			query:    "",
			opts:     SearchOptions{Type: TypeSkill, RepoName: "official"},
			wantLen:  1,
			wantName: "code-review",
		},
		{
			name:     "type and repo filter - community commands",
			query:    "",
			opts:     SearchOptions{Type: TypeCommand, RepoName: "community"},
			wantLen:  1,
			wantName: "deploy",
		},
		{
			name:    "query, type, and repo",
			query:   "code",
			opts:    SearchOptions{Type: TypeSkill, RepoName: "official"},
			wantLen: 1,
		},
		{
			name:    "no matches with combined filters",
			query:   "deploy",
			opts:    SearchOptions{Type: TypeSkill, RepoName: "official"},
			wantLen: 0,
		},
		{
			name:     "official agents",
			query:    "",
			opts:     SearchOptions{Type: TypeAgent, RepoName: "official"},
			wantLen:  1,
			wantName: "security-guru",
		},
		{
			name:     "community mcp",
			query:    "",
			opts:     SearchOptions{Type: TypeMCP, RepoName: "community"},
			wantLen:  1,
			wantName: "database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Search(resources, tt.query, tt.opts)

			if len(results) != tt.wantLen {
				names := make([]string, len(results))
				for i, r := range results {
					names[i] = r.Name
				}
				t.Errorf("Search(%q, %+v) returned %d results %v, want %d", tt.query, tt.opts, len(results), names, tt.wantLen)
			}

			if tt.wantLen == 1 && tt.wantName != "" && results[0].Name != tt.wantName {
				t.Errorf("Search(%q, %+v) first result = %q, want %q", tt.query, tt.opts, results[0].Name, tt.wantName)
			}
		})
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	resources := testResources()

	t.Run("empty query returns all resources", func(t *testing.T) {
		results := Search(resources, "", SearchOptions{})
		if len(results) != len(resources) {
			t.Errorf("Search(\"\") returned %d results, want %d", len(results), len(resources))
		}
	})

	t.Run("empty query with type filter", func(t *testing.T) {
		results := Search(resources, "", SearchOptions{Type: TypeSkill})
		for _, r := range results {
			if r.Type != TypeSkill {
				t.Errorf("Search with TypeSkill filter returned %q type", r.Type)
			}
		}
	})

	t.Run("empty query with repo filter", func(t *testing.T) {
		results := Search(resources, "", SearchOptions{RepoName: "official"})
		for _, r := range results {
			if r.RepoName != "official" {
				t.Errorf("Search with official repo filter returned %q repo", r.RepoName)
			}
		}
	})

	t.Run("whitespace-only query treated as empty", func(t *testing.T) {
		// Note: The current implementation lowercases but doesn't trim whitespace
		// This test documents the actual behavior
		results := Search(resources, "   ", SearchOptions{})
		// Whitespace query won't match anything since no name/description contains just spaces
		if len(results) != 0 {
			t.Errorf("Search(\"   \") returned %d results, want 0 (whitespace doesn't match)", len(results))
		}
	})
}

func TestSearch_NoResults(t *testing.T) {
	resources := testResources()

	tests := []struct {
		name  string
		query string
		opts  SearchOptions
	}{
		{
			name:  "query matches nothing",
			query: "zzzznonexistent",
			opts:  SearchOptions{},
		},
		{
			name:  "query exists but filtered out by type",
			query: "deploy",
			opts:  SearchOptions{Type: TypeSkill},
		},
		{
			name:  "query exists but filtered out by repo",
			query: "github",
			opts:  SearchOptions{RepoName: "community"},
		},
		{
			name:  "valid query, invalid type and repo combination",
			query: "code-review",
			opts:  SearchOptions{Type: TypeMCP, RepoName: "community"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Search(resources, tt.query, tt.opts)
			if len(results) != 0 {
				names := make([]string, len(results))
				for i, r := range results {
					names[i] = r.Name
				}
				t.Errorf("Search(%q, %+v) returned %d results %v, want 0", tt.query, tt.opts, len(results), names)
			}
		})
	}
}

func TestSearch_Scoring(t *testing.T) {
	// Create resources specifically to test scoring priorities
	resources := []Resource{
		{Name: "code", Description: "Exact name match", Type: TypeSkill, RepoName: "test"},
		{Name: "code-review", Description: "Prefix match", Type: TypeSkill, RepoName: "test"},
		{Name: "my-code-tool", Description: "Contains in name", Type: TypeSkill, RepoName: "test"},
		{Name: "formatter", Description: "Formats code nicely", Type: TypeSkill, RepoName: "test"},
	}

	t.Run("exact match scores highest", func(t *testing.T) {
		results := Search(resources, "code", SearchOptions{})
		if len(results) == 0 {
			t.Fatal("Search returned no results")
		}
		if results[0].Name != "code" {
			t.Errorf("Search(\"code\") first result = %q, want %q (exact match)", results[0].Name, "code")
		}
	})

	t.Run("prefix match scores higher than contains", func(t *testing.T) {
		// Search for "code" - code-review should come before my-code-tool
		results := Search(resources, "code", SearchOptions{})
		if len(results) < 3 {
			t.Fatalf("Search returned %d results, want at least 3", len(results))
		}

		codeReviewIdx := -1
		myCodeToolIdx := -1
		for i, r := range results {
			if r.Name == "code-review" {
				codeReviewIdx = i
			}
			if r.Name == "my-code-tool" {
				myCodeToolIdx = i
			}
		}

		if codeReviewIdx == -1 {
			t.Fatal("code-review not found in results")
		}
		if myCodeToolIdx == -1 {
			t.Fatal("my-code-tool not found in results")
		}
		if codeReviewIdx > myCodeToolIdx {
			t.Errorf("code-review (prefix match) at index %d, my-code-tool (contains) at index %d; prefix should come first",
				codeReviewIdx, myCodeToolIdx)
		}
	})

	t.Run("contains in name scores higher than description only", func(t *testing.T) {
		results := Search(resources, "code", SearchOptions{})
		if len(results) < 4 {
			t.Fatalf("Search returned %d results, want 4", len(results))
		}

		myCodeToolIdx := -1
		formatterIdx := -1
		for i, r := range results {
			if r.Name == "my-code-tool" {
				myCodeToolIdx = i
			}
			if r.Name == "formatter" {
				formatterIdx = i
			}
		}

		if myCodeToolIdx == -1 {
			t.Fatal("my-code-tool not found in results")
		}
		if formatterIdx == -1 {
			t.Fatal("formatter not found in results")
		}
		if myCodeToolIdx > formatterIdx {
			t.Errorf("my-code-tool (name contains) at index %d, formatter (description only) at index %d; name match should come first",
				myCodeToolIdx, formatterIdx)
		}
	})

	t.Run("description only match is last", func(t *testing.T) {
		results := Search(resources, "code", SearchOptions{})
		if len(results) != 4 {
			t.Fatalf("Search returned %d results, want 4", len(results))
		}
		if results[3].Name != "formatter" {
			t.Errorf("Search(\"code\") last result = %q, want %q (description only match)", results[3].Name, "formatter")
		}
	})

	t.Run("scoring order is stable", func(t *testing.T) {
		// Run search multiple times to ensure consistent ordering
		for i := range 10 {
			results := Search(resources, "code", SearchOptions{})
			expected := []string{"code", "code-review", "my-code-tool", "formatter"}
			if len(results) != len(expected) {
				t.Fatalf("iteration %d: Search returned %d results, want %d", i, len(results), len(expected))
			}
			for j, r := range results {
				if r.Name != expected[j] {
					t.Errorf("iteration %d, position %d: got %q, want %q", i, j, r.Name, expected[j])
				}
			}
		}
	})
}

func TestSearch_EmptyResources(t *testing.T) {
	t.Run("nil resources", func(t *testing.T) {
		results := Search(nil, "test", SearchOptions{})
		if len(results) != 0 {
			t.Errorf("Search(nil, ...) returned %v, want nil or empty", results)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		results := Search([]Resource{}, "test", SearchOptions{})
		if len(results) != 0 {
			t.Errorf("Search([], ...) returned %d results, want 0", len(results))
		}
	})

	t.Run("empty query on empty resources", func(t *testing.T) {
		results := Search([]Resource{}, "", SearchOptions{})
		if len(results) != 0 {
			t.Errorf("Search([], \"\", ...) returned %d results, want 0", len(results))
		}
	})

	t.Run("empty resources with filters", func(t *testing.T) {
		opts := SearchOptions{Type: TypeSkill, RepoName: "test"}
		results := Search([]Resource{}, "query", opts)
		if len(results) != 0 {
			t.Errorf("Search([], ..., filters) returned %d results, want 0", len(results))
		}
	})
}

func TestSearch_EdgeCases(t *testing.T) {
	resources := testResources()

	t.Run("partial name match", func(t *testing.T) {
		results := Search(resources, "run", SearchOptions{})
		// Should match test-runner
		found := false
		for _, r := range results {
			if r.Name == "test-runner" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Search(\"run\") should match test-runner")
		}
	})

	t.Run("hyphenated query", func(t *testing.T) {
		results := Search(resources, "code-review", SearchOptions{})
		if len(results) == 0 || results[0].Name != "code-review" {
			t.Error("Search(\"code-review\") should find code-review as first result")
		}
	})

	t.Run("single character query", func(t *testing.T) {
		results := Search(resources, "g", SearchOptions{})
		// Should match github, codegen, and others with 'g'
		if len(results) == 0 {
			t.Error("Search(\"g\") should return results")
		}
	})

	t.Run("query longer than any name", func(t *testing.T) {
		results := Search(resources, "this-is-a-very-long-query-that-wont-match-anything", SearchOptions{})
		if len(results) != 0 {
			t.Errorf("Search with very long query returned %d results, want 0", len(results))
		}
	})
}

func TestSearch_ScoreMatch(t *testing.T) {
	// Test the internal scoreMatch function behavior through Search ordering
	resources := []Resource{
		{Name: "test", Description: "A test resource", Type: TypeSkill, RepoName: "r"},
		{Name: "testing", Description: "Another one", Type: TypeSkill, RepoName: "r"},
		{Name: "my-test", Description: "Contains test", Type: TypeSkill, RepoName: "r"},
		{Name: "other", Description: "Has test in description", Type: TypeSkill, RepoName: "r"},
	}

	results := Search(resources, "test", SearchOptions{})
	if len(results) != 4 {
		t.Fatalf("Search returned %d results, want 4", len(results))
	}

	// Expected order: test (exact=100), testing (prefix=75), my-test (contains=50), other (desc=25)
	expected := []string{"test", "testing", "my-test", "other"}
	for i, exp := range expected {
		if results[i].Name != exp {
			t.Errorf("position %d: got %q, want %q", i, results[i].Name, exp)
		}
	}
}

func TestMatchesFilters(t *testing.T) {
	r := Resource{
		Name:     "test",
		Type:     TypeSkill,
		RepoName: "official",
	}

	tests := []struct {
		name string
		opts SearchOptions
		want bool
	}{
		{
			name: "no filters",
			opts: SearchOptions{},
			want: true,
		},
		{
			name: "matching type filter",
			opts: SearchOptions{Type: TypeSkill},
			want: true,
		},
		{
			name: "non-matching type filter",
			opts: SearchOptions{Type: TypeCommand},
			want: false,
		},
		{
			name: "matching repo filter",
			opts: SearchOptions{RepoName: "official"},
			want: true,
		},
		{
			name: "non-matching repo filter",
			opts: SearchOptions{RepoName: "community"},
			want: false,
		},
		{
			name: "both filters matching",
			opts: SearchOptions{Type: TypeSkill, RepoName: "official"},
			want: true,
		},
		{
			name: "type matches but repo doesn't",
			opts: SearchOptions{Type: TypeSkill, RepoName: "community"},
			want: false,
		},
		{
			name: "repo matches but type doesn't",
			opts: SearchOptions{Type: TypeCommand, RepoName: "official"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesFilters(r, tt.opts)
			if got != tt.want {
				t.Errorf("matchesFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesQuery(t *testing.T) {
	tests := []struct {
		name  string
		r     Resource
		query string
		want  bool
	}{
		{
			name:  "matches name exactly",
			r:     Resource{Name: "test", Description: "desc"},
			query: "test",
			want:  true,
		},
		{
			name:  "matches name substring",
			r:     Resource{Name: "testing", Description: "desc"},
			query: "test",
			want:  true,
		},
		{
			name:  "matches description",
			r:     Resource{Name: "other", Description: "this is a test"},
			query: "test",
			want:  true,
		},
		{
			name:  "no match",
			r:     Resource{Name: "other", Description: "something else"},
			query: "test",
			want:  false,
		},
		{
			name:  "case insensitive name",
			r:     Resource{Name: "Test", Description: "desc"},
			query: "test",
			want:  true,
		},
		{
			name:  "case insensitive description",
			r:     Resource{Name: "other", Description: "This is a TEST"},
			query: "test",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesQuery(tt.r, tt.query)
			if got != tt.want {
				t.Errorf("matchesQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
