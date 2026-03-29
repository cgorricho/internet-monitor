#!/bin/bash

# Internet Monitor - Dashboard Generation Script
# Generates static HTML dashboard on-demand

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY_NAME="internet-monitor"

# Find the appropriate binary
find_binary() {
    # Check if we have a local build
    if [[ -f "$PROJECT_DIR/dist/${BINARY_NAME}-linux-amd64" ]]; then
        echo "$PROJECT_DIR/dist/${BINARY_NAME}-linux-amd64"
        return
    fi
    
    # Check for binary in current directory
    if [[ -f "$PROJECT_DIR/$BINARY_NAME" ]]; then
        echo "$PROJECT_DIR/$BINARY_NAME"
        return
    fi
    
    # Check if binary is in PATH
    if command -v "$BINARY_NAME" &> /dev/null; then
        echo "$BINARY_NAME"
        return
    fi
    
    echo ""
}

# Show usage
usage() {
    echo "📊 Internet Monitor Dashboard Generator"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --hours N        Hours of data to include (default: 24)"
    echo "  -c, --compare        Generate comparative dashboard with paired machines"
    echo "  -o, --output FILE    Output filename (default: dashboard.html)"
    echo "  --no-browser         Don't open browser automatically"
    echo "  --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                              # Generate 24-hour dashboard"
    echo "  $0 --hours 7 --compare          # Generate 7-day comparative dashboard"
    echo "  $0 --output weekly.html -h 168  # Generate weekly report"
    echo ""
}

# Parse command line arguments
HOURS=""
COMPARE=""
OUTPUT=""
NO_BROWSER=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--hours)
            HOURS="--hours $2"
            shift 2
            ;;
        -c|--compare)
            COMPARE="--compare"
            shift
            ;;
        -o|--output)
            OUTPUT="--output $2"
            shift 2
            ;;
        --no-browser)
            NO_BROWSER="--no-browser"
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "❌ Unknown option: $1"
            echo ""
            usage
            exit 1
            ;;
    esac
done

# Find binary
BINARY=$(find_binary)
if [[ -z "$BINARY" ]]; then
    echo "❌ Internet Monitor binary not found!"
    echo ""
    echo "Please ensure the binary is available by:"
    echo "1. Running './scripts/build.sh' to build the binary"
    echo "2. Or installing the binary to your PATH"
    echo ""
    exit 1
fi

echo "📊 Generating Internet Monitor Dashboard..."
echo "🔧 Using binary: $BINARY"

# Change to project directory
cd "$PROJECT_DIR"

# Build command
CMD="$BINARY dashboard $HOURS $COMPARE $OUTPUT $NO_BROWSER"

echo "▶️  Running: $CMD"
echo ""

# Execute dashboard generation
if eval "$CMD"; then
    echo ""
    echo "✅ Dashboard generated successfully!"
    
    # Determine output filename
    if [[ -n "$OUTPUT" ]]; then
        DASHBOARD_FILE=$(echo "$OUTPUT" | sed 's/--output //')
    else
        DASHBOARD_FILE="dashboard.html"
    fi
    
    echo "📄 Dashboard saved as: $DASHBOARD_FILE"
    echo "📂 Location: $(pwd)/$DASHBOARD_FILE"
    
    # Show file size
    if [[ -f "$DASHBOARD_FILE" ]]; then
        FILE_SIZE=$(du -h "$DASHBOARD_FILE" | cut -f1)
        echo "📦 File size: $FILE_SIZE"
    fi
    
    echo ""
    echo "🌐 To view the dashboard:"
    echo "   - Open $DASHBOARD_FILE in your web browser"
    echo "   - Or run: xdg-open $DASHBOARD_FILE"
    
else
    echo ""
    echo "❌ Dashboard generation failed!"
    echo "Please check the error messages above."
    exit 1
fi