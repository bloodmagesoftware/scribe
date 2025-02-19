package ignore

import (
	"scribe/internal/config"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

func GetMatcher(c *config.Config) gitignore.Matcher {
	patterns := make([]gitignore.Pattern, len(c.Ignore)+2)
	for i, s := range c.Ignore {
		patterns[i] = gitignore.ParsePattern(s, nil)
	}
	patterns[len(c.Ignore)] = gitignore.ParsePattern("/.scribe/", nil)
	patterns[len(c.Ignore)+1] = gitignore.ParsePattern("/.scribe.yaml", nil)
	return gitignore.NewMatcher(patterns)
}
