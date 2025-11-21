#!/bin/bash

# Setup script for Monitor Detail Page
# This script creates the necessary directory structure and copies the monitor detail page

set -e

echo "Setting up Monitor Detail Page..."

# Navigate to frontend directory
cd "$(dirname "$0")"

# Create the [id] directory under domains
echo "Creating directory structure..."
mkdir -p src/routes/domains/[id]

# Copy the monitor detail page
echo "Installing monitor detail page..."
if [ -f "/tmp/monitor-detail-page.svelte" ]; then
    cp /tmp/monitor-detail-page.svelte src/routes/domains/[id]/+page.svelte
    echo "✓ Monitor detail page installed successfully!"
else
    echo "✗ Error: /tmp/monitor-detail-page.svelte not found"
    echo "  Please ensure the file exists before running this script"
    exit 1
fi

echo ""
echo "Setup complete! The monitor detail page is now available at:"
echo "  src/routes/domains/[id]/+page.svelte"
echo ""
echo "You can now access monitor details by navigating to /domains/{monitor-id}"
