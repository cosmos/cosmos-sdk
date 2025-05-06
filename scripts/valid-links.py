import re
import requests
from pathlib import Path
from urllib.parse import urlparse
from concurrent.futures import ThreadPoolExecutor, as_completed

# Constants
TIMEOUT = 5
CHECKED = set()
REPORT = []
WEB_TASKS = []
LINK_RE = re.compile(r'\[.*?\]\((.*?)\)')
HEADER_RE = re.compile(r'^#+\s+(.*)')

# Convert Markdown heading text to GitHub-style anchor slug
def slugify(header):
    return re.sub(r'[^a-z0-9\- ]', '', header.lower()).replace(' ', '-').strip('-')

# Check a web URL; fallback to GET if HEAD returns 429 (rate-limited)
def check_web_link(link):
    try:
        response = requests.head(link, allow_redirects=True, timeout=TIMEOUT)
        if response.status_code == 429:
            response = requests.get(link, allow_redirects=True, timeout=TIMEOUT)
        status = response.status_code
        if 200 <= status < 400:
            return link, "✅ Valid (web)", status
        else:
            return link, "❌ Invalid (web)", status
    except Exception as e:
        return link, "⚠️ Error (web)", str(e)

# Extract all slugified headings from a Markdown file
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
            return link, "✅ Exists (heading)", "✓"
        else:
            return link, "❌ Missing (heading)", "anchor not found"
    elif '#' in link:
        file_part, anchor = link.split('#', 1)
        target = (base_path.parent / file_part).resolve()
        if not target.exists():
            return link, "❌ Missing (file)", "file not found"
        headings = extract_headings(target)
        if slugify(anchor) in headings:
            return link, "✅ Exists (file+heading)", "✓"
        else:
            return link, "❌ Missing (heading)", "anchor not found"
    else:
        resolved = (base_path.parent / link).resolve()
        if resolved.exists():
            return link, "✅ Exists (local)", "✓"
        else:
            return link, "❌ Missing (local)", "file not found"

# Extract and check all links in a Markdown file
def check_links_in_file(path):
    print(f"🔗 Checking: {path}")
    with open(path, 'r', encoding='utf-8') as f:
        content = f.read()
        links = LINK_RE.findall(content)
        for link in links:
            if link in CHECKED:
                continue
            CHECKED.add(link)
            if link.startswith("https://") or link.startswith("https://"):
                WEB_TASKS.append((link, str(path)))
            else:
                result = check_local_path(link, path)
                if result:
                    REPORT.append((str(path), *result))

# Main entry point for walking markdown files and checking links
def main():
    print("🔍 Checking links in markdown files...")
    md_files = list(Path('.').rglob('*.md'))
    for idx, file_path in enumerate(md_files, 1):
        print(f"📄 Progress: {idx}/{len(md_files)}")
        check_links_in_file(file_path)

    if WEB_TASKS:
        print(f"🌐 Checking {len(WEB_TASKS)} web links in parallel...")
        with ThreadPoolExecutor(max_workers=4) as executor:
            futures = {executor.submit(check_web_link, url): (url, src) for url, src in WEB_TASKS}
            for future in as_completed(futures):
                url, src = futures[future]
                try:
                    link, status, detail = future.result()
                    REPORT.append((src, link, status, detail))
                except Exception as e:
                    REPORT.append((src, url, "⚠️ Thread error", str(e)))

    if REPORT:
        passed = sum(1 for r in REPORT if r[2].startswith("✅"))
        failed = len(REPORT) - passed
        print("\n✅ Passed: {} | ❌ Failed: {}\n".format(passed, failed))

        with open("docs_links_status.md", "w") as f:
            f.write("| File | Link | Status | Detail |\n")
            f.write("|------|------|--------|--------|\n")
            for file, link, status, detail in REPORT:
                f.write(f"| `{file}` | `{link}` | {status} | {detail} |\n")
        print("🧾 Wrote link report to docs_links_status.md")

if __name__ == "__main__":
    main()