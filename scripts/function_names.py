import os
import re
from typing import List, Dict, Optional

ALLOWED_PREFIXES = ["Deprecated:", "TODO:", "FIXME:"]

def find_go_files(root_dir: str) -> List[str]:
    go_files: List[str] = []
    for subdir, _, files in os.walk(root_dir):
        for file in files:
            if file.endswith(".go") and not (
                    file.endswith("_test.go") or
                    file.endswith(".pb.go") or
                    file.endswith(".pulsar.go")
            ):
                go_files.append(os.path.join(subdir, file))
    return go_files

def check_func_comments(filepath: str) -> List[Dict[str, object]]:
    with open(filepath, 'r') as f:
        lines = f.readlines()

    issues: List[Dict[str, object]] = []
    i = 0
    while i < len(lines) - 1:
        comment_block: List[str] = []
        comment_start_line: Optional[int] = None

        while i < len(lines) and lines[i].strip().startswith('//'):
            if comment_start_line is None:
                comment_start_line = i
            comment_block.append(lines[i].strip())
            i += 1

        if i < len(lines):
            func_match = re.match(r'func\s+(\(\w+\s+\*?\w+\)\s+)?(\w+)', lines[i])
            if func_match and comment_block:
                func_name = func_match.group(2)

                for line in comment_block:
                    comment_content = line[2:].strip()
                    if not comment_content:
                        continue
                    first_word = comment_content.split()[0]

                    if first_word != func_name and not any(comment_content.startswith(prefix) for prefix in ALLOWED_PREFIXES):
                        issues.append({
                            "file": filepath,
                            "line": comment_start_line + 1,
                            "expected": func_name,
                            "found": first_word
                        })
                    break
        i += 1
    return issues

def main() -> None:
    root_dir: str = '.'  # or any path
    go_files = find_go_files(root_dir)

    all_issues: List[Dict[str, object]] = []
    for file in go_files:
        issues = check_func_comments(file)
        all_issues.extend(issues)

    for issue in all_issues:
        print(f"{issue['file']}:{issue['line']} - Comment '{issue['found']}' should be '{issue['expected']}'")

if __name__ == "__main__":
    main()
