import os
import re

ALLOWED_PREFIXES = ["Deprecated:", "TODO:", "FIXME:"]

def find_go_files(root_dir):
    go_files = []
    for subdir, _, files in os.walk(root_dir):
        for file in files:
            if file.endswith(".go") and not (
                    file.endswith("_test.go") or
                    file.endswith(".pb.go") or
                    file.endswith(".pulsar.go")
            ):
                go_files.append(os.path.join(subdir, file))
    return go_files

def check_func_comments(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    issues = []
    i = 0
    while i < len(lines) - 1:
        # Gather consecutive comment lines
        comment_block = []
        comment_start_line = None
        while i < len(lines) and lines[i].strip().startswith('//'):
            if comment_start_line is None:
                comment_start_line = i
            comment_block.append(lines[i].strip())
            i += 1

        # Look for a function declaration
        if i < len(lines):
            func_match = re.match(r'func\s+(\(\w+\s+\*?\w+\)\s+)?(\w+)', lines[i])
            if func_match and comment_block:
                func_name = func_match.group(2)

                # Get first meaningful comment line
                for line in comment_block:
                    comment_content = line[2:].strip()  # strip `//` and spaces
                    if not comment_content:
                        continue
                    first_word = comment_content.split()[0]

                    # Accept if it matches the function name or an allowed prefix
                    if first_word != func_name and not any(comment_content.startswith(prefix) for prefix in ALLOWED_PREFIXES):
                        issues.append({
                            "file": filepath,
                            "line": comment_start_line + 1,
                            "expected": func_name,
                            "found": first_word
                        })
                    break  # Only inspect the first relevant comment line
        i += 1
    return issues

def main():
    root_dir = '.'  # Update as needed
    go_files = find_go_files(root_dir)

    all_issues = []
    for file in go_files:
        issues = check_func_comments(file)
        all_issues.extend(issues)

    for issue in all_issues:
        print(f"{issue['file']}:{issue['line']} - Comment '{issue['found']}' should be '{issue['expected']}'")

if __name__ == "__main__":
    main()
