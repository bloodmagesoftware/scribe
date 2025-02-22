package ignore

import (
	"scribe/internal/config"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

func GetMatcher(c *config.Config) gitignore.Matcher {
	ignore := strings.Split(c.Ignore, "\n")
	patterns := make([]gitignore.Pattern, 0, len(ignore)+2)
	for _, s := range ignore {
		ts := strings.TrimSpace(s)
		if len(ts) == 0 || strings.HasPrefix(ts, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(ts, nil))
	}
	patterns = append(patterns, gitignore.ParsePattern("/.scribe/", nil))
	patterns = append(patterns, gitignore.ParsePattern("/.scribe.yaml", nil))
	return gitignore.NewMatcher(patterns)
}
