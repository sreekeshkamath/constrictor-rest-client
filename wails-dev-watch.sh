#!/bin/bash

# Watch script for Wails development
# This watches the web/src directory and rebuilds the frontend when files change

echo "🔍 Watching web/src for changes..."
echo "💡 Run 'make wails-dev' in another terminal to start Wails"
echo ""

# Watch for changes in web/src and rebuild
while inotifywait -r -e modify,create,delete,move web/src 2>/dev/null || fswatch -o web/src 2>/dev/null; do
    echo "📝 Frontend files changed, rebuilding..."
    cd web && npm run build && cd ..
    rm -rf wails/frontend
    cp -r web/dist wails/frontend
    echo "✅ Frontend rebuilt! Refresh the Wails app to see changes."
    echo ""
done
