#!/usr/bin/env python3
"""Synchronize generated feature documentation for a merged pull request.

The script is intentionally deterministic so it can run in GitHub Actions without
an LLM dependency. It records each feature-bearing merged PR once and preserves
manual edits around the managed documentation region.
"""
from __future__ import annotations

import argparse
import datetime as dt
import re
import subprocess
from pathlib import Path
from typing import Iterable

ROOT = Path(__file__).resolve().parents[1]
DOC_PATH = ROOT / "docs" / "features.md"

CHANGELOG_START = "<!-- mothership-docs:changelog:start -->"
CHANGELOG_END = "<!-- mothership-docs:changelog:end -->"
FEATURES_START = "<!-- mothership-docs:features:start -->"
FEATURES_END = "<!-- mothership-docs:features:end -->"

DEFAULT_DOC = f"""# Feature Documentation

This page is the canonical feature documentation index for Mothership. It is
updated after feature-bearing pull requests merge to `main`, then refined in the
follow-up documentation PR when a generated entry needs more detail.

## What each feature entry must include

Every documented feature should include:

- a changelog entry that identifies the merged PR and the user-visible change
- a feature description that explains the capability and why it exists
- usage help that tells maintainers how to install, configure, or invoke it

## Feature Changelog

{CHANGELOG_START}
{CHANGELOG_END}

## Feature Catalog

{FEATURES_START}
{FEATURES_END}
"""

DOC_ONLY_PREFIXES = (
    "docs/",
    ".github/",
)
DOC_ONLY_FILES = {
    "README.md",
    "AGENTS.md",
}
FEATURE_PREFIXES = (
    "agents/",
    "mothership-config/",
    "skills/",
    "tools/",
)
FEATURE_FILES = {
    "install.sh",
}


def run_git(args: list[str]) -> str:
    result = subprocess.run(
        ["git", *args],
        cwd=ROOT,
        check=True,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return result.stdout.strip()


def changed_paths_for_commit(commit: str) -> list[str]:
    if not commit:
        return []
    output = run_git(["diff-tree", "--no-commit-id", "--name-only", "-r", commit])
    return [line for line in output.splitlines() if line]


def is_feature_path(path: str) -> bool:
    return path in FEATURE_FILES or path.startswith(FEATURE_PREFIXES)


def is_docs_only(paths: Iterable[str]) -> bool:
    paths = list(paths)
    if not paths:
        return False
    return all(path in DOC_ONLY_FILES or path.startswith(DOC_ONLY_PREFIXES) for path in paths)


def extract_heading_section(body: str, aliases: Iterable[str]) -> str:
    aliases_norm = {alias.lower() for alias in aliases}
    lines = body.replace("\r\n", "\n").split("\n")
    for index, line in enumerate(lines):
        match = re.match(r"^(#{1,6})\s+(.+?)\s*$", line)
        if not match:
            continue
        level = len(match.group(1))
        title = re.sub(r"[^a-z0-9 ]+", "", match.group(2).lower()).strip()
        if title not in aliases_norm:
            continue
        collected: list[str] = []
        for next_line in lines[index + 1 :]:
            next_match = re.match(r"^(#{1,6})\s+", next_line)
            if next_match and len(next_match.group(1)) <= level:
                break
            collected.append(next_line)
        return "\n".join(collected).strip()
    return ""


def extract_comment_block(body: str, name: str) -> str:
    pattern = re.compile(
        rf"<!--\s*mothership-docs:{re.escape(name)}:start\s*-->(.*?)<!--\s*mothership-docs:{re.escape(name)}:end\s*-->",
        re.IGNORECASE | re.DOTALL,
    )
    match = pattern.search(body)
    return match.group(1).strip() if match else ""


def first_meaningful_paragraph(body: str) -> str:
    cleaned_lines: list[str] = []
    for raw_line in body.replace("\r\n", "\n").split("\n"):
        line = raw_line.strip()
        if not line or line.startswith("<!--") or line.startswith("#"):
            if cleaned_lines:
                break
            continue
        if line.startswith(("- [ ]", "- [x]", "- [X]")):
            continue
        cleaned_lines.append(line)
    return " ".join(cleaned_lines).strip()


def normalize_block(text: str, fallback: str) -> str:
    text = text.strip()
    if not text:
        return fallback
    return text


def markdown_list(items: list[str], limit: int = 12) -> str:
    shown = items[:limit]
    lines = [f"  - `{item}`" for item in shown]
    if len(items) > limit:
        lines.append(f"  - …and {len(items) - limit} more")
    return "\n".join(lines) if lines else "  - _No changed paths were available._"


def ensure_doc() -> str:
    if not DOC_PATH.exists():
        DOC_PATH.parent.mkdir(parents=True, exist_ok=True)
        DOC_PATH.write_text(DEFAULT_DOC, encoding="utf-8")
    text = DOC_PATH.read_text(encoding="utf-8")
    for marker in (CHANGELOG_START, CHANGELOG_END, FEATURES_START, FEATURES_END):
        if marker not in text:
            raise SystemExit(f"{DOC_PATH} is missing required marker: {marker}")
    return text


def insert_after_marker(text: str, start: str, entry: str) -> str:
    return text.replace(start, f"{start}\n\n{entry.strip()}\n", 1)


def build_entries(args: argparse.Namespace, paths: list[str]) -> tuple[str, str]:
    body = args.pr_body or ""
    summary = normalize_block(
        extract_comment_block(body, "changelog")
        or extract_heading_section(body, ["feature changelog", "changelog", "summary"])
        or first_meaningful_paragraph(body),
        f"Merged `{args.pr_title}`. Review and expand this generated changelog summary if the PR body did not include release-note detail.",
    )
    description = normalize_block(
        extract_comment_block(body, "description")
        or extract_heading_section(body, ["feature description", "feature descriptions", "description", "what changed"]),
        f"`{args.pr_title}` introduced or changed feature behavior in the files listed below. Review the merged implementation and replace this generated placeholder with a maintainer-authored description before merging the documentation PR.",
    )
    usage = normalize_block(
        extract_comment_block(body, "usage")
        or extract_heading_section(body, ["usage help", "usage", "how to use", "testing"]),
        "Usage help was not provided in the merged PR body. Document the commands, setup steps, host-specific behavior, or operator workflow required to use this change before merging the documentation PR.",
    )
    merged_at = args.merged_at or dt.datetime.now(dt.UTC).strftime("%Y-%m-%dT%H:%M:%SZ")
    merge_sha = args.merge_sha or "unknown"
    source = args.pr_url or f"PR #{args.pr_number}"
    path_list = markdown_list(paths)

    changelog = f"""<!-- mothership-docs:pr:{args.pr_number}:changelog -->
### PR #{args.pr_number} — {args.pr_title}

- **Merged:** {merged_at}
- **Source:** {source}
- **Merge commit:** `{merge_sha}`
- **Change summary:** {summary}
- **Changed paths:**
{path_list}
"""

    feature = f"""<!-- mothership-docs:pr:{args.pr_number}:feature -->
### PR #{args.pr_number} — {args.pr_title}

**Feature description**

{description}

**Usage help**

{usage}
"""
    return changelog, feature


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pr-number", required=True)
    parser.add_argument("--pr-title", required=True)
    parser.add_argument("--pr-url", default="")
    parser.add_argument("--pr-body", default="")
    parser.add_argument("--pr-body-file", default="")
    parser.add_argument("--merge-sha", default="")
    parser.add_argument("--merged-at", default="")
    parser.add_argument("--changed-path", action="append", default=[])
    args = parser.parse_args()
    if args.pr_body_file:
        args.pr_body = Path(args.pr_body_file).read_text(encoding="utf-8")
    return args


def main() -> None:
    args = parse_args()
    paths = args.changed_path or changed_paths_for_commit(args.merge_sha)
    paths = sorted(dict.fromkeys(paths))

    if is_docs_only(paths):
        print("Merged PR only changed documentation or automation files; no feature documentation sync needed.")
        return
    if paths and not any(is_feature_path(path) for path in paths):
        print("Merged PR did not touch a recognized feature path; no feature documentation sync needed.")
        return

    text = ensure_doc()
    pr_marker = f"<!-- mothership-docs:pr:{args.pr_number}:"
    if pr_marker in text:
        print(f"PR #{args.pr_number} is already documented.")
        return

    changelog, feature = build_entries(args, paths)
    text = insert_after_marker(text, CHANGELOG_START, changelog)
    text = insert_after_marker(text, FEATURES_START, feature)
    DOC_PATH.write_text(text, encoding="utf-8")
    print(f"Updated {DOC_PATH.relative_to(ROOT)} for PR #{args.pr_number}.")


if __name__ == "__main__":
    main()
