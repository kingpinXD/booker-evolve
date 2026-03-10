#!/usr/bin/env python3
"""Format GitHub issues for safe injection into Claude prompts.

Reads JSON from stdin (gh issue list --json output), sorts by
engagement, sanitizes content, and outputs a bounded text block
with nonce-based boundaries to prevent prompt injection.
"""

import json
import sys
import hashlib
import time
import re

MAX_BODY_CHARS = 500
MAX_ISSUES = 10
MAX_COMMENTS = 3


def generate_nonce():
    """Generate a unique boundary nonce to prevent prompt injection."""
    raw = f"{time.time()}-{id(sys)}"
    return hashlib.sha256(raw.encode()).hexdigest()[:16]


def sanitize(text, nonce):
    """Strip potential prompt injection markers from issue text."""
    if not text:
        return ""
    # Remove common injection patterns
    text = re.sub(r"<\|.*?\|>", "", text)
    text = re.sub(r"```system.*?```", "", text, flags=re.DOTALL)
    text = re.sub(r"IDENTITY|PERSONALITY|SKILL\.md", "[REDACTED]", text)
    # Remove the nonce itself if someone tries to inject it
    text = text.replace(nonce, "[REDACTED]")
    # Truncate
    if len(text) > MAX_BODY_CHARS:
        text = text[:MAX_BODY_CHARS] + "... [truncated]"
    return text.strip()


def net_votes(issue):
    """Calculate engagement score for sorting."""
    reactions = issue.get("reactionGroups", [])
    thumbs_up = sum(
        r.get("users", {}).get("totalCount", 0)
        for r in reactions
        if r.get("content") == "THUMBS_UP"
    )
    thumbs_down = sum(
        r.get("users", {}).get("totalCount", 0)
        for r in reactions
        if r.get("content") == "THUMBS_DOWN"
    )
    comments = len(issue.get("comments", []))
    return thumbs_up - thumbs_down + comments


def format_issues(issues):
    """Format issues into a bounded, sanitized text block."""
    nonce = generate_nonce()

    # Sort by engagement, limit count
    issues.sort(key=net_votes, reverse=True)
    issues = issues[:MAX_ISSUES]

    if not issues:
        print("No open issues.")
        return

    print(f"--- BEGIN ISSUES [{nonce}] ---")

    for issue in issues:
        number = issue.get("number", "?")
        title = sanitize(issue.get("title", ""), nonce)
        body = sanitize(issue.get("body", ""), nonce)
        labels = ", ".join(l.get("name", "") for l in issue.get("labels", []))

        print(f"\n### Issue #{number}: {title}")
        if labels:
            print(f"Labels: {labels}")
        if body:
            print(f"Body: {body}")

        # Include first few comments
        comments = issue.get("comments", [])[:MAX_COMMENTS]
        for c in comments:
            author = c.get("author", {}).get("login", "unknown")
            cbody = sanitize(c.get("body", ""), nonce)
            if cbody:
                print(f"  Comment by @{author}: {cbody}")

    print(f"\n--- END ISSUES [{nonce}] ---")


if __name__ == "__main__":
    try:
        data = json.load(sys.stdin)
        format_issues(data)
    except json.JSONDecodeError:
        print("Error: Invalid JSON input", file=sys.stderr)
        sys.exit(1)
