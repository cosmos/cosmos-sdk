import json

# Load the JSON file
with open('lychee/lychee.json', 'r') as file:
    data = json.load(file)

# Extract filenames from fail_map
failed_files = list(data.get("fail_map", {}).keys())

# Save the filenames to a new text file
with open('lychee/recheck_files.txt', 'w') as file:
    for filename in failed_files:
        file.write(f"{filename}\n")
