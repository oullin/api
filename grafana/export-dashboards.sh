#!/bin/bash
# Helper script to export Grafana dashboards

set -e

GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
GRAFANA_USER="${GRAFANA_USER:-admin}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-admin}"
OUTPUT_DIR="./grafana/dashboards"

echo "================================"
echo "Grafana Dashboard Export Tool"
echo "================================"
echo ""

# Check if Grafana is running
if ! curl -s "$GRAFANA_URL/api/health" > /dev/null 2>&1; then
    echo "Error: Grafana is not accessible at $GRAFANA_URL"
    echo "Please start Grafana with: make monitor-up"
    exit 1
fi

# List all dashboards
echo "Fetching dashboard list..."
DASHBOARDS=$(curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
    "$GRAFANA_URL/api/search?type=dash-db" | jq -r '.[] | "\(.uid) \(.title)"')

if [ -z "$DASHBOARDS" ]; then
    echo "No dashboards found in Grafana"
    exit 0
fi

echo ""
echo "Available dashboards:"
echo "---------------------"
echo "$DASHBOARDS" | nl
echo ""

# Ask user which dashboard to export
read -p "Enter dashboard number to export (or 'all' for all dashboards): " SELECTION

if [ "$SELECTION" = "all" ]; then
    # Export all dashboards
    echo ""
    echo "Exporting all dashboards..."

    while IFS= read -r line; do
        UID=$(echo "$line" | awk '{print $1}')
        TITLE=$(echo "$line" | cut -d' ' -f2-)
        FILENAME=$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd '[:alnum:]-').json

        echo "Exporting: $TITLE -> $FILENAME"

        curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
            "$GRAFANA_URL/api/dashboards/uid/$UID" | \
            jq 'del(.meta) | .dashboard.id = null | .overwrite = true' > \
            "$OUTPUT_DIR/$FILENAME"
    done <<< "$DASHBOARDS"

else
    # Export single dashboard
    SELECTED_LINE=$(echo "$DASHBOARDS" | sed -n "${SELECTION}p")
    UID=$(echo "$SELECTED_LINE" | awk '{print $1}')
    TITLE=$(echo "$SELECTED_LINE" | cut -d' ' -f2-)
    FILENAME=$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd '[:alnum:]-').json

    echo ""
    echo "Exporting: $TITLE"

    curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
        "$GRAFANA_URL/api/dashboards/uid/$UID" | \
        jq 'del(.meta) | .dashboard.id = null | .overwrite = true' > \
        "$OUTPUT_DIR/$FILENAME"

    echo "Saved to: $OUTPUT_DIR/$FILENAME"
fi

echo ""
echo "Export complete!"
echo ""
echo "To reload dashboards:"
echo "  make monitor-restart"
