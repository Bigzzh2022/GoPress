package markdown

import (
	"strings"
	"testing"
)

func TestRenderSanitizesXSSAndKeepsMarkdown(t *testing.T) {
	renderer := NewRenderer()
	source := `# Title

<script>alert('xss')</script>
<img src="x" onerror="alert(1)">
[bad](javascript:alert(1))

` + "```go\nfmt.Println(\"safe\")\n```" + `
| A | B |
| - | - |
| 1 | 2 |
`
	output, err := renderer.Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	lower := strings.ToLower(output)
	for _, forbidden := range []string{"<script", "onerror", "javascript:"} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("Render() contains forbidden %q: %s", forbidden, output)
		}
	}
	for _, expected := range []string{"<h1", "<table>", "language-go", "fmt.Println"} {
		if !strings.Contains(output, expected) {
			t.Errorf("Render() missing %q: %s", expected, output)
		}
	}
}

func TestExcerptReturnsPlainText(t *testing.T) {
	renderer := NewRenderer()
	excerpt := renderer.Excerpt("# Hello\n\nThis is **Markdown** with [a link](https://example.com).", 200)
	if strings.Contains(excerpt, "<") || strings.Contains(excerpt, "# Hello") {
		t.Fatalf("Excerpt() = %q, want plain text", excerpt)
	}
	if !strings.Contains(excerpt, "Hello") || !strings.Contains(excerpt, "Markdown") {
		t.Fatalf("Excerpt() lost text: %q", excerpt)
	}
}
