# For mass processing resource files

import json

stuff = []
newstuff = {}

# Load txt file line by line
with open("chars.txt") as file:
	for line in file:
		chara = line.strip()
		print(f"{chara}")
		stuff.append(chara)

# Load json file
# with open("t2.json", "r") as file:
# stuff = json.load(file)

# for key, val in stuff.items():
# newstuff[key] = f"{val}th"

# Write json file
with open("data.json", "w", encoding="utf-8") as f:
	json.dump(stuff, f, ensure_ascii=False, indent=4)
