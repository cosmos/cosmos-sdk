import glob
import json

# Path to the folder containing JSON files
folder_path = 'lychee'

# Get all JSON files in the specified folder
json_files = glob.glob(f"{folder_path}/__*.json")

# List to store filenames with broken links
files_with_broken_links = []

# Loop through each JSON file
for json_file in json_files:
    with open(json_file, 'r') as file:
        data = json.load(file)
        # Check if "fail_map" contains broken links
        if "fail_map" in data and data["fail_map"]:
            # Get the keys (filenames with broken links)
            filenames = list(data["fail_map"].keys())
            print(filenames)
            # Add to the list of files with broken links
            files_with_broken_links.extend(filenames)

# Remove duplicates
files_with_broken_links = list(set(files_with_broken_links))

# Save the filenames to recheck.txt
with open('recheck.txt', 'w') as output_file:
    for filename in files_with_broken_links:
        output_file.write(f"{filename}\n")
