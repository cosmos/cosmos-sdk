#!/bin/bash
# Simple HTTP server for the reveal.js presentation
PORT=${1:-8000}

echo "Starting presentation server on http://localhost:$PORT"
echo "Press Ctrl+C to stop the server"
echo ""

# Try Python 3 first, then Python 2
if command -v python3 &> /dev/null; then
    python3 -m http.server $PORT
elif command -v python &> /dev/null; then
    python -m SimpleHTTPServer $PORT
else
    echo "Error: Python is not installed. Please install Python or use another method."
    exit 1
fi