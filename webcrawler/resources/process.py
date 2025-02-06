# For mass processing resource files

import json

stuff = {}
newstuff = {}

# Load txt file line by line
# with open("nums.txt") as file:
# for line in file:
# number, text = line.strip().split(",")
# print(f"{number}, {text}")
# stuff[text] = number

# Load json file
with open("t2.json", "r") as file:
	stuff = json.load(file)

for key, val in stuff.items():
	newstuff[key] = f"{val}th"

# Write json file
with open("data.json", "w", encoding="utf-8") as f:
	json.dump(newstuff, f, ensure_ascii=False, indent=4)
