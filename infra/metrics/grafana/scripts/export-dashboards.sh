#!/bin/bash
# Helper script to export Grafana dashboards

set -e

GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
GRAFANA_USER="${GRAFANA_USER:-admin}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-admin}"
OUTPUT_DIR="./infra/metrics/grafana/dashboards"

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
read -r -p "Enter dashboard number to export (or 'all' for all dashboards): " SELECTION

# Validate selection
if [ "$SELECTION" != "all" ]; then
    # Check if selection is a valid number
    if ! [[ "$SELECTION" =~ ^[0-9]+$ ]]; then
        echo "Error: Please enter a valid number or 'all'"
        exit 1
    fi

    # Check if selection is within valid range
    DASHBOARD_COUNT=$(echo "$DASHBOARDS" | wc -l)
    if [ "$SELECTION" -lt 1 ] || [ "$SELECTION" -gt "$DASHBOARD_COUNT" ]; then
        echo "Error: Selection out of range (1-$DASHBOARD_COUNT)"
        exit 1
    fi
fi

if [ "$SELECTION" = "all" ]; then
    # Export all dashboards
    echo ""
    echo "Exporting all dashboards..."

    EXPORT_COUNT=0
    FAIL_COUNT=0

    while IFS= read -r line; do
        UID=$(echo "$line" | awk '{print $1}')
        TITLE=$(echo "$line" | cut -d' ' -f2-)
        FILENAME="${UID}-$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd '[:alnum:]-').json"

        echo -n "Exporting: $TITLE -> $FILENAME ... "

        # Temporarily disable errexit for this operation
        set +e
        if curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
            "$GRAFANA_URL/api/dashboards/uid/$UID" | \
            jq 'del(.meta) | .dashboard.id = null | .overwrite = true' > \
            "$OUTPUT_DIR/$FILENAME" 2>/dev/null; then

            # Verify the file is valid JSON and not empty
            if [ -s "$OUTPUT_DIR/$FILENAME" ] && jq empty "$OUTPUT_DIR/$FILENAME" 2>/dev/null; then
                echo "✓ Success"
                ((EXPORT_COUNT++))
            else
                echo "✗ Failed (invalid JSON)"
                rm -f "$OUTPUT_DIR/$FILENAME"
                ((FAIL_COUNT++))
            fi
        else
            echo "✗ Failed (export error)"
            rm -f "$OUTPUT_DIR/$FILENAME"
            ((FAIL_COUNT++))
        fi
        set -e
    done <<< "$DASHBOARDS"

    echo ""
    echo "Export summary: $EXPORT_COUNT succeeded, $FAIL_COUNT failed"

    if [ $FAIL_COUNT -gt 0 ]; then
        exit 1
    fi

else
    # Export single dashboard
    SELECTED_LINE=$(echo "$DASHBOARDS" | sed -n "${SELECTION}p")

    if [ -z "$SELECTED_LINE" ]; then
        echo "Error: Invalid selection"
        exit 1
    fi

    UID=$(echo "$SELECTED_LINE" | awk '{print $1}')
    TITLE=$(echo "$SELECTED_LINE" | cut -d' ' -f2-)
    FILENAME="${UID}-$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd '[:alnum:]-').json"

    echo ""
    echo "Exporting: $TITLE"

    # Temporarily disable errexit for this operation
    set +e
    if curl -s -u "$GRAFANA_USER:$GRAFANA_PASSWORD" \
        "$GRAFANA_URL/api/dashboards/uid/$UID" | \
        jq 'del(.meta) | .dashboard.id = null | .overwrite = true' > \
        "$OUTPUT_DIR/$FILENAME" 2>/dev/null; then

        # Verify the file is valid JSON and not empty
        if [ -s "$OUTPUT_DIR/$FILENAME" ] && jq empty "$OUTPUT_DIR/$FILENAME" 2>/dev/null; then
            echo "✓ Saved to: $OUTPUT_DIR/$FILENAME"
        else
            echo "✗ Error: Export produced invalid JSON"
            rm -f "$OUTPUT_DIR/$FILENAME"
            exit 1
        fi
    else
        echo "✗ Error: Failed to export dashboard"
        rm -f "$OUTPUT_DIR/$FILENAME"
        exit 1
    fi
    set -e
fi

echo ""
echo "Export complete!"
echo ""
echo "To reload dashboards:"
echo "  make monitor-restart"
