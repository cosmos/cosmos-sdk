import re
import subprocess
import sys
from pathlib import Path
from difflib import SequenceMatcher
from functools import lru_cache

TARGET_VERSION = "v0.53.0"
REPO = "cosmos/cosmos-sdk"
FUZZY_THRESHOLD = 0.70

LINK_RE = re.compile(
    rf"https://github.com/{REPO}/blob/([^/]+)/(.+?)#L(\d+)(?:-L(\d+))?"
)

ERROR_LOG = []
DRY_RUN = '--dry-run' in sys.argv

@lru_cache(maxsize=1024)
def get_file_at_version(version, filepath):
    # Fetch file content at a given Git version and path
    try:
        result = subprocess.run(
            ["git", "show", f"{version}:{filepath}"],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.splitlines()
    except subprocess.CalledProcessError:
        return None

def normalize_lines(lines):
    # Normalize whitespace in lines
    return [re.sub(r'\s+', ' ', line.strip()) for line in lines if line.strip()]

def get_lines_range(lines, start, end):
    # Extract and normalize a line range
    if not lines or end > len(lines):
        return None
    return normalize_lines(lines[start-1:end])

def fuzzy_match_range(old_lines, new_lines, threshold=FUZZY_THRESHOLD):
    # Try to find a matching block in the new version using a fixed-size window and fuzzy similarity
    old_str = "\n".join(normalize_lines(old_lines))
    old_len = len(old_lines)
    best_score = 0.0
    best_range = None
    candidates = []

    for i in range(len(new_lines) - old_len + 1):
        window = new_lines[i:i + old_len]
        window_str = "\n".join(normalize_lines(window))
        score = SequenceMatcher(None, old_str, window_str).ratio()
        candidates.append((i + 1, i + old_len, score))
        if score > best_score:
            best_score = score
            best_range = (i + 1, i + old_len)

    # Optional: Print top few candidates for debugging/review
    top_candidates = sorted(candidates, key=lambda x: -x[2])[:3]
    for start, end, score in top_candidates:
        print(f"🔎 Candidate range L{start}-L{end} scored {score:.2f}")

    if best_score >= threshold:
        return best_range, best_score
    return None, best_score

def update_links_in_file(file_path):
    with open(file_path, 'r') as f:
        content = f.read()

    changed = False

    def replacer(match):
        nonlocal changed
        old_version, path, start_line_str, end_line_str = match.groups()

        # Skip already-correct version
        if old_version == TARGET_VERSION:
            return match.group(0)

        # Skip links that reference exact Git commit hashes
        if re.fullmatch(r"[a-f0-9]{7,40}", old_version):
            return match.group(0)

        start_line = int(start_line_str)
        end_line = int(end_line_str) if end_line_str else start_line

        old_lines = get_file_at_version(old_version, path)
        new_lines = get_file_at_version(TARGET_VERSION, path)

        if not old_lines or not new_lines:
            reason = f"Could not fetch {path} at {old_version} or {TARGET_VERSION}"
            ERROR_LOG.append({"url": match.group(0), "file": str(file_path), "reason": reason, "score": None})
            print(f"⚠️  {reason}")
            return match.group(0)

        old_segment = get_lines_range(old_lines, start_line, end_line)
        if not old_segment:
            reason = f"Invalid range {start_line}-{end_line} in {path} at {old_version}"
            ERROR_LOG.append({"url": match.group(0), "file": str(file_path), "reason": reason, "score": None})
            print(f"⚠️  {reason}")
            return match.group(0)

        result, score = fuzzy_match_range(old_segment, new_lines)
        if result:
            new_start, new_end = sorted(result)
            suffix = f"#L{new_start}-L{new_end}" if new_start != new_end else f"#L{new_start}"
            new_url = f"https://github.com/{REPO}/blob/{TARGET_VERSION}/{path}{suffix}"
            print(f"✅ Updated (fuzzy {score:.2f}): {match.group(0)} → {new_url}")
            changed = True
            return new_url

        reason = f"Fuzzy match failed in {path} — best score: {score:.2f}"
        ERROR_LOG.append({"url": match.group(0), "file": str(file_path), "reason": reason, "score": round(score, 2)})
        print(f"⚠️  {reason}")
        return match.group(0)

    new_content = LINK_RE.sub(replacer, content)

    if changed and not DRY_RUN:
        with open(file_path, 'w') as f:
            f.write(new_content)
        print(f"📝 File updated: {file_path}")

def main():
    print("🔍 Scanning for outdated GitHub links...")
    for md_file in Path('.').rglob('*.md'):
        update_links_in_file(md_file)

    if ERROR_LOG:
        with open("docs_todo.log", "w") as f:
            for e in ERROR_LOG:
                line = f"{e['url']} | {e['file']} | {e['reason']}"
                if e['score'] is not None:
                    line += f" | score={e['score']}"
                f.write(line + "\n")
        print("🧾 Wrote failure log to docs_todo.log")

if __name__ == "__main__":
    main()
