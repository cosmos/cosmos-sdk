import re
import subprocess
import sys
from pathlib import Path
from difflib import SequenceMatcher
from operator import itemgetter

TARGET_VERSION = "v0.53.0"
REPO = "cosmos/cosmos-sdk"
FUZZY_THRESHOLD = 0.80

LINK_RE = re.compile(
    rf"https://github.com/{REPO}/blob/([^/]+)/(.+?)#L(\d+)(?:-L(\d+))?"
)

ERROR_LOG = []
DRY_RUN = '--dry-run' in sys.argv

def get_file_at_version(version, filepath):
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

def get_lines_range(lines, start, end):
    if not lines or end > len(lines):
        return None
    return [line.strip() for line in lines[start-1:end]]

def ordered_contains(old_lines, window):
    old_idx = 0
    for line in window:
        if line == old_lines[old_idx]:
            old_idx += 1
            if old_idx == len(old_lines):
                return True
    return False

def find_range_in_new_version(old_lines, new_lines):
    max_window = len(old_lines) + 10
    for start in range(0, len(new_lines) - len(old_lines) + 1):
        for window_size in range(len(old_lines), max_window + 1):
            end = start + window_size
            if end > len(new_lines):
                break
            window = [line.strip() for line in new_lines[start:end]]
            if ordered_contains(old_lines, window):
                return start + 1, end
    return None, None

def fuzzy_match_range(old_lines, new_lines, threshold=FUZZY_THRESHOLD):
    best_score = 0.0
    best_range = None
    old_str = "\n".join(old_lines)
    old_len = len(old_lines)
    window_range = range(old_len - 2, old_len + 3)  # ±2 lines

    for window_size in window_range:
        if window_size <= 0:
            continue
        for i in range(len(new_lines) - window_size + 1):
            window = new_lines[i:i + window_size]
            window_str = "\n".join([line.strip() for line in window])
            score = SequenceMatcher(None, old_str, window_str).ratio()
            if score > best_score:
                best_score = score
                best_range = (i + 1, i + window_size)

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
        if old_version == TARGET_VERSION:
            return match.group(0)

        start_line = int(start_line_str)
        end_line = int(end_line_str) if end_line_str else start_line

        old_lines = get_file_at_version(old_version, path)
        new_lines = get_file_at_version(TARGET_VERSION, path)

        if not old_lines or not new_lines:
            reason = f"Could not fetch {path} at {old_version} or {TARGET_VERSION}"
            ERROR_LOG.append({
                "url": match.group(0),
                "file": str(file_path),
                "reason": reason,
                "score": None
            })
            print(f"⚠️  {reason}")
            return match.group(0)

        old_segment = get_lines_range(old_lines, start_line, end_line)
        if not old_segment:
            reason = f"Invalid range {start_line}-{end_line} in {path} at {old_version}"
            ERROR_LOG.append({
                "url": match.group(0),
                "file": str(file_path),
                "reason": reason,
                "score": None
            })
            print(f"⚠️  {reason}")
            return match.group(0)

        # Try ordered match
        new_start, new_end = find_range_in_new_version(old_segment, new_lines)
        if new_start and new_end:
            suffix = f"#L{new_start}-L{new_end}" if new_start != new_end else f"#L{new_start}"
            new_url = f"https://github.com/{REPO}/blob/{TARGET_VERSION}/{path}{suffix}"
            print(f"✅ Updated (windowed): {match.group(0)} → {new_url}")
            changed = True
            return new_url

        # Fuzzy match fallback
        fuzzy_result, score = fuzzy_match_range(old_segment, new_lines)
        if fuzzy_result:
            new_start, new_end = fuzzy_result
            suffix = f"#L{new_start}-L{new_end}" if new_start != new_end else f"#L{new_start}"
            new_url = f"https://github.com/{REPO}/blob/{TARGET_VERSION}/{path}{suffix}"
            print(f"✅ Updated (fuzzy {score:.2f}): {match.group(0)} → {new_url}")
            changed = True
            return new_url

        reason = f"Fuzzy match failed in {path} — best score: {score:.2f}"
        ERROR_LOG.append({
            "url": match.group(0),
            "file": str(file_path),
            "reason": reason,
            "score": round(score, 2)
        })
        print(f"⚠️  {reason}")
        return match.group(0)

    new_content = LINK_RE.sub(replacer, content)

    if changed and not DRY_RUN:
        with open(file_path, 'w') as f:
            f.write(new_content)
        print(f"📝 File updated: {file_path}")

def summarize_errors(errors):
    summary = {
        "Fuzzy Fail": 0,
        "Invalid Range": 0,
        "File Not Found": 0,
        "Other": 0
    }

    for e in errors:
        r = e["reason"]
        if "Invalid range" in r:
            summary["Invalid Range"] += 1
        elif "Could not fetch" in r:
            summary["File Not Found"] += 1
        elif "Fuzzy match failed" in r:
            summary["Fuzzy Fail"] += 1
        else:
            summary["Other"] += 1

    print("\n📊 Summary of unresolved references:")
    for k, v in summary.items():
        print(f"  {k}: {v}")
    return summary

def write_markdown_report(errors):
    with open("docs_report.md", "w") as f:
        f.write("| Type | File | URL | Score | Fix Hint |\n")
        f.write("|------|------|-----|-------|-----------|\n")

        for e in errors:
            reason = e["reason"]
            score = e["score"]
            if "Invalid range" in reason:
                category = "Invalid Range"
                hint = "Check file length"
            elif "Could not fetch" in reason:
                category = "File Not Found"
                hint = "Check git ref or file path"
            elif "Fuzzy match failed" in reason:
                category = "Fuzzy Fail"
                hint = "Lower threshold or manual fix"
            else:
                category = "Other"
                hint = "Manual review"

            f.write(f"| {category} | `{e['file']}` | [{e['url']}]({e['url']}) | {score if score else ''} | {hint} |\n")

def main():
    print("🔍 Scanning for outdated GitHub links...")
    for md_file in Path('.').rglob('*.md'):
        update_links_in_file(md_file)

    if ERROR_LOG:
        sorted_errors = sorted(ERROR_LOG, key=lambda x: x["score"] if x["score"] is not None else -1, reverse=True)

        with open("docs_todo.log", "w") as f:
            for e in sorted_errors:
                line = f"{e['url']} | {e['file']} | {e['reason']}"
                if e['score'] is not None:
                    line += f" | score={e['score']}"
                f.write(line + "\n")

        write_markdown_report(sorted_errors)
        summarize_errors(sorted_errors)
        print("\n🧾 Wrote failure log to `docs_todo.log` and markdown report to `docs_report.md`")

if __name__ == "__main__":
    main()
