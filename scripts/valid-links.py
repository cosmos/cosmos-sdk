import re
import requests
import json
from pathlib import Path
from urllib.parse import urlparse
import time

# Constants
TIMEOUT = 5
CACHE_FILE = ".link_cache.json"
CHECKED = set()
CACHE = {}
REPORT = []
WEB_TASKS = []
LINK_RE = re.compile(r'\[.*?\]\((.*?)\)')
HEADER_RE = re.compile(r'^#+\s+(.*)')

# Load cache if exists
if Path(CACHE_FILE).exists():
    with open(CACHE_FILE, 'r') as f:
        CACHE = json.load(f)

# Convert markdown heading text to GitHub-style anchor slug
def slugify(header):
    return re.sub(r'[^a-z0-9\- ]', '', header.lower()).replace(' ', '-').strip('-')

# Check a web URL; fallback to GET if HEAD returns 429 (rate-limited), cache successes
def check_web_link(link):
    if link in CACHE and CACHE[link]["status"].startswith("✅"):
        return link, CACHE[link]["status"], CACHE[link]["detail"]

    try:
        # Try HEAD first
        response = requests.head(link, allow_redirects=True, timeout=TIMEOUT)
        if response.status_code in (405, 429):
            retries = 0
            backoff = 1
            while retries < 3:
                print(f"⏳ Fallback to GET (status {response.status_code}), retry {retries + 1} after {backoff}s: {link}")
                time.sleep(backoff)
                response = requests.get(link, allow_redirects=True, timeout=TIMEOUT)
                if response.status_code not in (405, 429):
                    break
                backoff *= 2
                retries += 1

        status = response.status_code
        if status == 999 or link.startswith("https://www.linkedin.com"):
            result = (link, "⚠️ Possibly valid (blocked)", status)
        elif 200 <= status < 400 or status in (403, 405):
            result = (link, "✅ Valid (web)", status)
        else:
            result = (link, "❌ Invalid (web)", status)
        CACHE[link] = {"status": result[1], "detail": result[2]}
        return result
    except Exception as e:
        return (link, "⚠️ Error (web)", str(e))

# Extract all slugified headings from a markdown file
def extract_headings(file_path):
    headings = set()
    with open(file_path, 'r', encoding='utf-8') as f:
        for line in f:
            match = HEADER_RE.match(line)
            if match:
                headings.add(slugify(match.group(1)))
    return headings

# Validate local file paths and intra-doc anchor references
def check_local_path(link, base_path):
    if link.startswith('#'):
        headings = extract_headings(base_path)
        slug = link[1:]
        if slug in headings:
            return (link, "✅ Exists (heading)", "✓")
        else:
            return (link, "❌ Missing (heading)", "anchor not found")
    elif '#' in link:
        file_part, anchor = link.split('#', 1)
        target = (base_path.parent / file_part).resolve()
        if not target.exists():
            return (link, "❌ Missing (file)", "file not found")
        headings = extract_headings(target)
        if slugify(anchor) in headings:
            return (link, "✅ Exists (file+heading)", "✓")
        else:
            return (link, "❌ Missing (heading)", "anchor not found")
    else:
        resolved = (base_path.parent / link).resolve()
        if resolved.exists():
            return (link, "✅ Exists (local)", "✓")
        else:
            return (link, "❌ Missing (local)", "file not found")

# Extract and check all links in a Markdown file, tracking line number
def check_links_in_file(path):
    print(f"🔗 Checking: {path}")
    with open(path, 'r', encoding='utf-8') as f:
        in_code_block = False
        for line_num, line in enumerate(f, 1):
            if line.strip().startswith('```'):
                in_code_block = not in_code_block
                continue
            if in_code_block:
                continue
            if '`' in line:
                continue

            for match in LINK_RE.finditer(line):
                link = match.group(1)
                # Ignore non-links like `cdc` or `address` (no slashes or dots and short)
                if not ('/' in link or '.' in link or link.startswith('#')):
                    continue
                if link in CHECKED:
                    continue
                CHECKED.add(link)
                if link.startswith("http://") or link.startswith("https://"):
                    WEB_TASKS.append((link, str(path), line_num))
                else:
                    result = check_local_path(link, path)
                    if result:
                        REPORT.append((str(path), line_num, *result))

# Main entry point for walking markdown files and checking links
def main():
    print("🔍 Checking links in markdown files...")
    md_files = list(Path('.').rglob('*.md'))
    for idx, file_path in enumerate(md_files, 1):
        print(f"📄 Progress: {idx}/{len(md_files)}")
        check_links_in_file(file_path)

    if WEB_TASKS:
        print(f"🌐 Checking {len(WEB_TASKS)} web links serially...")
        for url, src, line in WEB_TASKS:
            try:
                link, status, detail = check_web_link(url)
                REPORT.append((src, line, link, status, detail))
            except Exception as e:
                REPORT.append((src, line, url, "⚠️ Thread error", str(e)))

    passed = sum(1 for r in REPORT if r[3].startswith("✅"))
    failed = len(REPORT) - passed
    print("\n✅ Passed: {} | ❌ Failed: {}\n".format(passed, failed))

    with open("docs_links_status.md", "w") as f:
        f.write("# ❌ Broken Links\n\n")
        f.write("| File | Line | Link | Status | Detail |\n")
        f.write("|------|------|------|--------|--------|\n")
        for file, line, link, status, detail in REPORT:
            if not status.startswith("✅"):
                f.write(f"| `{file}` | `{line}` | `{link}` | {status} | {detail} |\n")

        f.write("\n# ✅ Working Links\n\n")
        f.write("| File | Line | Link | Status | Detail |\n")
        f.write("|------|------|------|--------|--------|\n")
        for file, line, link, status, detail in REPORT:
            if status.startswith("✅"):
                f.write(f"| `{file}` | `{line}` | `{link}` | {status} | {detail} |\n")

    with open(CACHE_FILE, 'w') as f:
        json.dump(CACHE, f, indent=2)
    print("🧾 Wrote link report to docs_links_status.md and updated cache")

if __name__ == "__main__":
    main()
