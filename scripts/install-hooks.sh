#!/bin/bash

# Install git hooks for the project
# This script copies pre-commit and commit-msg hooks from scripts/hooks/ to .git/hooks/

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
SOURCE_HOOKS_DIR="$SCRIPT_DIR/hooks"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Installing Git Hooks${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if .git directory exists
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo -e "${RED}ERROR: .git directory not found!${NC}"
    echo -e "${YELLOW}Make sure you're running this from a git repository${NC}"
    exit 1
fi

# Check if source hooks directory exists
if [ ! -d "$SOURCE_HOOKS_DIR" ]; then
    echo -e "${RED}ERROR: Source hooks directory not found!${NC}"
    echo -e "${YELLOW}Expected: $SOURCE_HOOKS_DIR${NC}"
    exit 1
fi

# Install pre-commit hook
if [ -f "$SOURCE_HOOKS_DIR/pre-commit" ]; then
    echo -e "\n${GREEN}Installing pre-commit hook...${NC}"
    cp "$SOURCE_HOOKS_DIR/pre-commit" "$HOOKS_DIR/pre-commit"
    chmod +x "$HOOKS_DIR/pre-commit"
    echo -e "${GREEN}✓ Pre-commit hook installed${NC}"
else
    echo -e "${YELLOW}⚠ Pre-commit hook not found, skipping...${NC}"
fi

# Install commit-msg hook
if [ -f "$SOURCE_HOOKS_DIR/commit-msg" ]; then
    echo -e "\n${GREEN}Installing commit-msg hook...${NC}"
    cp "$SOURCE_HOOKS_DIR/commit-msg" "$HOOKS_DIR/commit-msg"
    chmod +x "$HOOKS_DIR/commit-msg"
    echo -e "${GREEN}✓ Commit-msg hook installed${NC}"
else
    echo -e "${YELLOW}⚠ Commit-msg hook not found, skipping...${NC}"
fi

# Summary
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✓ Git hooks installed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${YELLOW}What happens now:${NC}"
echo "1. Before each commit, pre-commit hook will run checks:"
echo "   - Check for debugging statements"
echo "   - Check for sensitive data"
echo "   - Verify code formatting (go fmt)"
echo "   - Check imports (goimports)"
echo "   - Run static analysis (go vet)"
echo "   - Run linter (golangci-lint)"
echo "   - Ensure go.mod is tidy"
echo ""
echo "2. Commit messages will be validated to follow conventional format:"
echo "   - Format: type(scope): message"
echo "   - Example: feat(auth): add login functionality"
echo ""
echo -e "${YELLOW}Optional tools (install for full functionality):${NC}"
echo "  • goimports: go install golang.org/x/tools/cmd/goimports@latest"
echo "  • golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"
