#!/bin/sh
# Entrypoint script for frontend Docker container
# Replaces environment variable placeholders in built files

# PUBLIC_API_URL is used for browser-side SSE connection
# This should be the URL accessible from the user's browser (e.g., http://localhost:8080 or https://api.yourdomain.com)
if [ -n "$PUBLIC_API_URL" ]; then
    echo "Injecting PUBLIC_API_URL for browser SSE: $PUBLIC_API_URL"
    # Find and replace in all HTML files in build directory
    find /app/build -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%|$PUBLIC_API_URL|g" {} + 2>/dev/null || true
    find /app/.svelte-kit -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%|$PUBLIC_API_URL|g" {} + 2>/dev/null || true
else
    echo "PUBLIC_API_URL not set, browser will use auto-detection"
    # Remove the placeholder so auto-detection kicks in
    find /app/build -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%||g" {} + 2>/dev/null || true
    find /app/.svelte-kit -type f -name "*.html" -exec sed -i "s|%PUBLIC_API_URL%||g" {} + 2>/dev/null || true
fi

# Log BACKEND_API_URL for debugging (used by server-side proxy)
if [ -n "$BACKEND_API_URL" ]; then
    echo "BACKEND_API_URL for server-side proxy: $BACKEND_API_URL"
fi

# Execute the main command
exec "$@"
