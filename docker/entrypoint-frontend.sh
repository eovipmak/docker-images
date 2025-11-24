#!/bin/sh
# Entrypoint script for frontend Docker container
# Replaces environment variable placeholders in built files

# Replace PUBLIC_API_URL in the app HTML if it exists
if [ -n "$PUBLIC_API_URL" ]; then
    echo "Injecting PUBLIC_API_URL: $PUBLIC_API_URL"
    # Find and replace in all HTML files in build directory
    find /app/build -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%|$PUBLIC_API_URL|g" {} +
    find /app/.svelte-kit -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%|$PUBLIC_API_URL|g" {} + 2>/dev/null || true
else
    echo "PUBLIC_API_URL not set, using auto-detection"
    # Remove the placeholder so auto-detection kicks in
    find /app/build -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%||g" {} +
    find /app/.svelte-kit -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%||g" {} + 2>/dev/null || true
fi

# Execute the main command
exec "$@"
