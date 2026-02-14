#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEMPLATE_DIR="$PROJECT_ROOT/services/workers/_templates/helper-worker"
WORKERS_DIR="$PROJECT_ROOT/services/workers"
INVENTORY_FILE="$PROJECT_ROOT/helpers-inventory.json"

# Check if inventory file exists
if [ ! -f "$INVENTORY_FILE" ]; then
    echo "Error: Helpers inventory not found at $INVENTORY_FILE"
    exit 1
fi

# Helper function to convert snake_case to kebab-case
snake_to_kebab() {
    echo "$1" | tr '_' '-'
}

# Helper function to convert snake_case to PascalCase
snake_to_pascal() {
    echo "$1" | awk -F_ '{for(i=1;i<=NF;i++){$i=toupper(substr($i,1,1)) substr($i,2)}} 1' OFS=""
}

# Helper function to scaffold a single worker
scaffold_worker() {
    local helper_type="$1"
    local category="$2"

    local helper_kebab=$(snake_to_kebab "$helper_type")
    local helper_pascal=$(snake_to_pascal "$helper_type")
    local worker_dir="$WORKERS_DIR/${helper_kebab}-worker"

    echo -e "${BLUE}Scaffolding $helper_kebab-worker (category: $category)${NC}"

    # Create worker directory
    mkdir -p "$worker_dir"

    # Copy and process serverless.yml template
    sed -e "s/{{HELPER_NAME_KEBAB}}/$helper_kebab/g" \
        -e "s/{{HELPER_NAME_PASCAL}}/$helper_pascal/g" \
        -e "s/{{HELPER_TYPE_SNAKE}}/$helper_type/g" \
        -e "s/{{CATEGORY}}/$category/g" \
        "$TEMPLATE_DIR/serverless.yml.template" > "$worker_dir/serverless.yml"

    # Copy and process main.go template
    sed -e "s/{{HELPER_NAME_KEBAB}}/$helper_kebab/g" \
        -e "s/{{HELPER_NAME_PASCAL}}/$helper_pascal/g" \
        -e "s/{{HELPER_TYPE_SNAKE}}/$helper_type/g" \
        -e "s/{{CATEGORY}}/$category/g" \
        "$TEMPLATE_DIR/main.go.template" > "$worker_dir/main.go"

    echo -e "${GREEN}✓ Created $worker_dir${NC}"
}

# Main scaffolding logic
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Helper Worker Scaffolding Script${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Parse JSON and scaffold workers
total_workers=0

for category in $(jq -r 'keys[]' "$INVENTORY_FILE"); do
    echo -e "${YELLOW}Processing category: $category${NC}"

    helpers=$(jq -r ".\"$category\"[]" "$INVENTORY_FILE")

    for helper in $helpers; do
        scaffold_worker "$helper" "$category"
        ((total_workers++))
    done

    echo ""
done

echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}✓ Successfully scaffolded $total_workers worker services!${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Review generated workers in $WORKERS_DIR"
echo "2. Commit changes to git"
echo "3. Update Teamwork task 44556529"
