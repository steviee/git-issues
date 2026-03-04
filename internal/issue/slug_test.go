package issue

import "testing"

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "simple ascii title",
			title: "Fix auth bug",
			want:  "fix-auth-bug",
		},
		{
			name:  "german umlauts",
			title: "Login schlägt bei leeren Passwörtern fehl",
			want:  "login-schlaegt-bei-leeren-passwoertern-f",
		},
		{
			name:  "special characters removed",
			title: "Hello! @World #2024",
			want:  "hello-world-2024",
		},
		{
			name:  "truncation at 40 chars with trailing dash stripped",
			title: "This is an extremely long title that should be truncated at exactly forty chars",
			want:  "this-is-an-extremely-long-title-that-sho",
		},
		{
			name:  "empty string",
			title: "",
			want:  "",
		},
		{
			name:  "uppercase umlauts",
			title: "Ä Ö Ü",
			want:  "ae-oe-ue",
		},
		{
			name:  "multiple spaces and dashes collapse",
			title: "foo  --  bar",
			want:  "foo-bar",
		},
		{
			name:  "eszett conversion",
			title: "Straße",
			want:  "strasse",
		},
		{
			name:  "trailing dash stripped after truncation",
			title: "abcdefghijklmnopqrstuvwxyz-abcdefghijklm nop",
			want:  "abcdefghijklmnopqrstuvwxyz-abcdefghijklm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.title)
			if got != tt.want {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestFilename(t *testing.T) {
	tests := []struct {
		name string
		id   int
		slug string
		want string
	}{
		{
			name: "standard id and slug",
			id:   7,
			slug: "login-bug",
			want: "0007-login-bug.md",
		},
		{
			name: "id with padding",
			id:   1,
			slug: "fix-auth-bug",
			want: "0001-fix-auth-bug.md",
		},
		{
			name: "large id",
			id:   1234,
			slug: "big-id",
			want: "1234-big-id.md",
		},
		{
			name: "five digit id exceeds padding",
			id:   12345,
			slug: "overflow",
			want: "12345-overflow.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filename(tt.id, tt.slug)
			if got != tt.want {
				t.Errorf("Filename(%d, %q) = %q, want %q", tt.id, tt.slug, got, tt.want)
			}
		})
	}
}
