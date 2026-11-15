package textutil

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "words", input: "  Hello, GoPress World!  ", want: "hello-gopress-world"},
		{name: "separators", input: "go___with---gorm", want: "go-with-gorm"},
		{name: "unicode", input: "Go 后端 入门", want: "go-后端-入门"},
		{name: "punctuation only", input: "!!!", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Slugify(test.input); got != test.want {
				t.Errorf("Slugify(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
