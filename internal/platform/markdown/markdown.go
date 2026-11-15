package markdown

import (
	"bytes"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var whitespace = regexp.MustCompile(`\s+`)

type Renderer struct {
	engine    goldmark.Markdown
	sanitizer *bluemonday.Policy
	plain     *bluemonday.Policy
}

func NewRenderer() *Renderer {
	policy := bluemonday.UGCPolicy()
	policy.AllowURLSchemes("http", "https", "mailto")
	policy.AllowRelativeURLs(true)
	policy.RequireNoFollowOnLinks(true)
	policy.AllowAttrs("class").Matching(regexp.MustCompile(`^language-[a-zA-Z0-9_-]+$`)).OnElements("code")

	return &Renderer{
		engine: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		),
		sanitizer: policy,
		plain:     bluemonday.StrictPolicy(),
	}
}

func (r *Renderer) Render(source string) (string, error) {
	var output bytes.Buffer
	if err := r.engine.Convert([]byte(source), &output); err != nil {
		return "", err
	}
	return r.sanitizer.Sanitize(output.String()), nil
}

func (r *Renderer) Excerpt(source string, limit int) string {
	rendered, err := r.Render(source)
	if err != nil {
		rendered = source
	}
	text := r.plain.Sanitize(rendered)
	text = html.UnescapeString(text)
	text = strings.TrimSpace(whitespace.ReplaceAllString(text, " "))
	if limit <= 0 || utf8.RuneCountInString(text) <= limit {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:limit])) + "…"
}
