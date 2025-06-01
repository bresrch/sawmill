#!/bin/bash

# Sawmill Example Runner
# Builds and executes all reference implementations

DIM='\033[2m'
RESET='\033[0m'

echo "Sawmill Examples - Reference Implementation Runner"
echo "=================================================="
echo

if [ ! -f "go.mod" ]; then
    echo -e "${DIM}Run from project root${RESET}"
    exit 1
fi

examples=(
    "basic:slog compatibility and basic operations"
    "nested-attributes:Structured data organization patterns"
    "colors:Terminal output optimization"
    "marks:Process flow tracking"
    "options-pattern:Configuration management"
    "key-value:Machine-parseable output"
    "multi-output:Fan-out logging architecture"
    "callbacks:Runtime context injection"
    "as-method:Explicit output based logging"
)

for example in "${examples[@]}"; do
    IFS=':' read -r name description <<< "$example"
    
    echo "[$name] $description"
    echo -e "${DIM}----------------------------------------${RESET}"
    
    if [ -f "examples/$name/main.go" ]; then
        echo -e "${DIM}Building examples/$name/main.go${RESET}"
        if go build "examples/$name/main.go" 2>/dev/null; then
            echo -e "${DIM}Executing${RESET}"
            echo
            
            go run "examples/$name/main.go" 2>&1 | head -12
            
            echo
            echo -e "${DIM}Complete${RESET}"
        else
            echo -e "${DIM}Build failed${RESET}"
        fi
    else
        echo -e "${DIM}File not found${RESET}"
    fi
    
    echo
    echo -e "${DIM}========================================${RESET}"
    echo
done

echo -e "${DIM}All implementations verified${RESET}"
echo
echo -e "${DIM}Usage:${RESET}"
echo -e "${DIM}  go run examples/basic/main.go${RESET}"
echo -e "${DIM}  go run examples/key-value/main.go${RESET}"
echo
echo -e "${DIM}Architecture patterns: examples/README.md${RESET}"