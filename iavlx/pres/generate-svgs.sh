#!/bin/bash
# Generate SVG files from DOT files

for f in *.dot; do
  if [ -f "$f" ]; then
    echo "Converting $f to SVG..."
    dot -Tsvg "$f" -o "${f}.svg"
  fi
done

echo "Done! SVG files generated."