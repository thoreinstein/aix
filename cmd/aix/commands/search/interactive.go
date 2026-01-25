package search

import (
	"fmt"
	"io"

	"github.com/ktr0731/go-fuzzyfinder"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/resource"
)

func runInteractiveSearch(w io.Writer, resources []resource.Resource) error {
	if len(resources) == 0 {
		fmt.Fprintln(w, "No resources found.")
		return nil
	}

	idx, err := fuzzyfinder.Find(
		resources,
		func(i int) string {
			return fmt.Sprintf("%s: %s (%s)", resources[i].Type, resources[i].Name, resources[i].RepoName)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			r := resources[i]
			return fmt.Sprintf("Type: %s\nRepo: %s\nName: %s\n\nDescription:\n%s",
				r.Type,
				r.RepoName,
				r.Name,
				r.Description,
			)
		}),
	)

	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil
		}
		return errors.Wrap(err, "interactive search failed")
	}

	// Output the selected item in a nice format
	r := resources[idx]
	fmt.Fprintf(w, "Selected: %s (%s)\n", r.Name, r.Type)
	fmt.Fprintf(w, "Repo: %s\n", r.RepoName)
	fmt.Fprintf(w, "Description: %s\n", r.Description)

	return nil
}
