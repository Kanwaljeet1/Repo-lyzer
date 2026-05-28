import urllib.request
import json

url = "https://api.github.com/repos/agnivo988/Repo-lyzer/pulls/336/reviews/4383620251/comments"
req = urllib.request.Request(url)
with open("pr_comments.txt", "w", encoding="utf-8") as f:
    with urllib.request.urlopen(req) as response:
        data = json.loads(response.read().decode())
        for c in data:
            f.write(f"File: {c.get('path')}\n")
            f.write(f"Line: {c.get('line')}\n")
            f.write(f"Diff: {c.get('diff_hunk')}\n")
            f.write(f"Body: {c.get('body')}\n")
            f.write("="*40 + "\n")
