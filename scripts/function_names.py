import os
import re

def find_go_files(root_dir):
    go_files = []
    for subdir, _, files in os.walk(root_dir):
        for file in files:
            # Ignore test files and generated files
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
    for i in range(len(lines) - 1):
        comment_match = re.match(r'//\s*(\w+)', lines[i])
        func_match = re.match(r'func\s+(\(\w+\s+\*?\w+\)\s+)?(\w+)', lines[i+1])
        if comment_match and func_match:
            comment_name = comment_match.group(1)
            func_name = func_match.group(2)
            if comment_name != func_name:
                issues.append({
                    "file": filepath,
                    "line": i + 1,
                    "expected": func_name,
                    "found": comment_name
                })
    return issues

def main():
    root_dir = '.'  # Change this to your Go project directory if needed
    go_files = find_go_files(root_dir)

    all_issues = []
    for file in go_files:
        issues = check_func_comments(file)
        all_issues.extend(issues)

    for issue in all_issues:
        print(f"{issue['file']}:{issue['line']} - Comment '{issue['found']}' should be '{issue['expected']}'")

if __name__ == "__main__":
    main()
