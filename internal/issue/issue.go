package issue

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var ValidStatuses = []string{"open", "in-progress", "closed", "wontfix"}
var ValidPriorities = []string{"low", "medium", "high", "critical"}

type Relations struct {
	Blocks     []int `yaml:"blocks,omitempty" json:"blocks,omitempty"`
	DependsOn  []int `yaml:"depends-on,omitempty" json:"depends-on,omitempty"`
	RelatedTo  []int `yaml:"related-to,omitempty" json:"related-to,omitempty"`
	Duplicates []int `yaml:"duplicates,omitempty" json:"duplicates,omitempty"`
}

type Issue struct {
	ID        int       `yaml:"id" json:"id"`
	Title     string    `yaml:"title" json:"title"`
	Status    string    `yaml:"status" json:"status"`
	Priority  string    `yaml:"priority" json:"priority"`
	Labels    []string  `yaml:"labels" json:"labels"`
	Relations Relations `yaml:"relations,omitempty" json:"relations,omitempty"`
	Created   string    `yaml:"created" json:"created"`
	Updated   string    `yaml:"updated" json:"updated"`
	Closed    string    `yaml:"closed,omitempty" json:"closed,omitempty"`

	Body     string `yaml:"-" json:"-"`
	FilePath string `yaml:"-" json:"-"`
}

func ParseFile(filePath string) (*Issue, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading issue file: %w", err)
	}
	issue, err := Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}
	issue.FilePath = filePath
	return issue, nil
}

func Parse(content string) (*Issue, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// First line must be ---
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return nil, fmt.Errorf("file must start with ---")
	}

	// Read YAML until closing ---
	var yamlLines []string
	foundClosing := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClosing = true
			break
		}
		yamlLines = append(yamlLines, line)
	}
	if !foundClosing {
		return nil, fmt.Errorf("missing closing --- delimiter")
	}

	yamlContent := strings.Join(yamlLines, "\n")

	var issue Issue
	if err := yaml.Unmarshal([]byte(yamlContent), &issue); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Ensure labels is not nil
	if issue.Labels == nil {
		issue.Labels = []string{}
	}

	// Read body (everything after closing ---)
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	if len(bodyLines) > 0 {
		body := strings.Join(bodyLines, "\n")
		// Strip single leading newline if present
		body = strings.TrimPrefix(body, "\n")
		issue.Body = body
	}

	return &issue, nil
}

func Marshal(issue *Issue) []byte {
	// Marshal frontmatter
	yamlData, _ := yaml.Marshal(issue)

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n")
	if issue.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(issue.Body)
		if !strings.HasSuffix(issue.Body, "\n") {
			buf.WriteString("\n")
		}
	}
	return []byte(buf.String())
}

func Validate(issue *Issue) error {
	if issue.ID <= 0 {
		return fmt.Errorf("id must be > 0")
	}
	if strings.TrimSpace(issue.Title) == "" {
		return fmt.Errorf("title must not be empty")
	}
	if !contains(ValidStatuses, issue.Status) {
		return fmt.Errorf("invalid status %q, must be one of: %s", issue.Status, strings.Join(ValidStatuses, ", "))
	}
	if !contains(ValidPriorities, issue.Priority) {
		return fmt.Errorf("invalid priority %q, must be one of: %s", issue.Priority, strings.Join(ValidPriorities, ", "))
	}
	if issue.Created == "" {
		return fmt.Errorf("created date must not be empty")
	}
	if _, err := time.Parse("2006-01-02", issue.Created); err != nil {
		return fmt.Errorf("invalid created date %q, must be YYYY-MM-DD", issue.Created)
	}
	return nil
}

func PriorityRank(priority string) int {
	switch priority {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func Today() string {
	return time.Now().Format("2006-01-02")
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
