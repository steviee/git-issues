build:
	go build -o issues .

install: build
	cp issues /usr/local/bin/issues

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f issues

setup-agent:
	@echo "Installing git-issues binary..."
	go install github.com/steviee/git-issues@latest
	@echo "Installing Claude Code skill..."
	@mkdir -p ~/.claude/skills/git-issues
	@cp skills/git-issues/SKILL.md ~/.claude/skills/git-issues/SKILL.md
	@echo "Done. Skill available as /git-issues in Claude Code."

.PHONY: build install test lint clean setup-agent
