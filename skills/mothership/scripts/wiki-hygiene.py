#!/usr/bin/env python3
import argparse
import datetime as dt
import pathlib
import re
import sys
from typing import List, Tuple


def strip_quotes(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {'"', "'"}:
        return value[1:-1]
    return value


def yaml_safe_string(value: str) -> str:
    """Return value in YAML-safe quoted form if it contains characters that could corrupt YAML structure."""
    if not value:
        return '""'
    if any(ch in value for ch in (":", "#", "[", "]", "{", "}", "!", "&", "*", "?", "|", ">", '"', "'", "%", "@", "`")):
        escaped = value.replace("\\", "\\\\").replace('"', '\\"')
        return f'"{escaped}"'
    if value[0] in {"-", " ", "\t", "~"}:
        escaped = value.replace("\\", "\\\\").replace('"', '\\"')
        return f'"{escaped}"'
    return value


def strip_inline_comment(value: str) -> str:
    in_single = False
    in_double = False
    for i, ch in enumerate(value):
        if ch == "'" and not in_double:
            in_single = not in_single
        elif ch == '"' and not in_single:
            in_double = not in_double
        elif ch == "#" and not in_single and not in_double:
            if i == 0 or value[i - 1].isspace():
                return value[:i].rstrip()
    return value.rstrip()


def parse_scalar(value: str):
    value = strip_quotes(strip_inline_comment(value.strip()))
    lowered = value.lower()
    if lowered == "true":
        return True
    if lowered == "false":
        return False
    if re.fullmatch(r"-?\d+", value):
        return int(value)
    return value


def parse_simple_yaml(path: pathlib.Path):
    data = {}
    stack = [(0, data)]
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        if not raw_line.strip() or raw_line.lstrip().startswith("#"):
            continue
        indent = len(raw_line) - len(raw_line.lstrip(" "))
        line = raw_line.strip()
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        while len(stack) > 1 and indent <= stack[-1][0]:
            stack.pop()
        current = stack[-1][1]
        if value.strip() == "":
            child = {}
            current[key.strip()] = child
            stack.append((indent, child))
        else:
            current[key.strip()] = parse_scalar(value)
    return data


def parse_frontmatter_freshness(path: pathlib.Path) -> str:
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0].strip() != "---":
        return ""
    for line in lines[1:]:
        if line.strip() == "---":
            break
        match = re.match(r"^freshness:\s*(.+?)\s*$", line)
        if match:
            return strip_quotes(strip_inline_comment(match.group(1)))
    return ""


def parse_iso_date(value: str):
    value = value.strip()
    if not value:
        return None
    try:
        return dt.date.fromisoformat(value)
    except ValueError:
        pass
    try:
        return dt.datetime.fromisoformat(value.replace("Z", "+00:00")).date()
    except ValueError:
        return None


def scan_pages(config_path: pathlib.Path):
    config = parse_simple_yaml(config_path)
    wiki = config.get("wiki", {})
    root = str(wiki.get("root", "")).strip()
    project = str(wiki.get("project", "")).strip()
    if not root or not project:
        raise ValueError(f"wiki.root or wiki.project not set in {config_path}")

    stale_threshold = int(wiki.get("stale_threshold_days", 90))
    archive_threshold = int(wiki.get("archive_after_days", 120))
    root_path = pathlib.Path(root).expanduser()
    project_dir = root_path / "projects" / project
    today = dt.datetime.now(dt.timezone.utc).date()

    stale_pages = []
    archive_pages = []
    for bucket in ("insights", "decisions"):
        for page in sorted((project_dir / bucket).glob("*.md")):
            freshness = parse_frontmatter_freshness(page)
            freshness_date = parse_iso_date(freshness)
            if not freshness_date:
                continue
            days_old = (today - freshness_date).days
            rel_path = page.relative_to(root_path).as_posix()
            entry = {
                "path": rel_path,
                "freshness": freshness,
                "days_old": days_old,
            }
            if days_old > archive_threshold:
                archive_pages.append(entry)
            elif days_old > stale_threshold:
                stale_pages.append(entry)

    return {
        "root": root_path,
        "project": project,
        "stale_threshold": stale_threshold,
        "archive_threshold": archive_threshold,
        "archive_dir": root_path / "archive",
        "project_dir": project_dir,
        "stale_pages": stale_pages,
        "archive_pages": archive_pages,
    }


def replace_stale_pages_in_wiki_block(block_lines: List[str], stale_paths: List[str]) -> List[str]:
    result = []
    i = 0
    while i < len(block_lines):
        line = block_lines[i]
        if re.match(r"^  stale_pages:\s*(\[\])?\s*$", line):
            i += 1
            while i < len(block_lines) and (block_lines[i].startswith("    ") or not block_lines[i].strip()):
                i += 1
            continue
        result.append(line)
        i += 1

    if len(result) > 1 and result[-1].strip():
        result.append("")
    if stale_paths:
        result.append("  stale_pages:")
        result.extend(f"    - {yaml_safe_string(path)}" for path in stale_paths)
    else:
        result.append("  stale_pages: []")
    return result


def update_hub_state(state_path: pathlib.Path, stale_paths: List[str]) -> None:
    state_path.parent.mkdir(parents=True, exist_ok=True)

    lines = []
    if state_path.exists():
        lines = state_path.read_text(encoding="utf-8").splitlines()
    wiki_start = None
    wiki_end = len(lines)
    for i, line in enumerate(lines):
        if re.match(r"^wiki:\s*$", line):
            wiki_start = i
            for j in range(i + 1, len(lines)):
                if re.match(r"^[A-Za-z0-9_]+:\s*(?:$|[^\s].*)", lines[j]):
                    wiki_end = j
                    break
            break

    wiki_block = ["wiki:"]
    if wiki_start is not None:
        wiki_block = lines[wiki_start:wiki_end]
    wiki_block = replace_stale_pages_in_wiki_block(wiki_block, stale_paths)

    if wiki_start is None:
        if lines and lines[-1].strip():
            lines.append("")
        lines.extend(wiki_block)
    else:
        lines = lines[:wiki_start] + wiki_block + lines[wiki_end:]

    state_path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def format_kb_stale(scan) -> str:
    lines = [
        "=== Stale Wiki Pages ===",
        f"Project: {scan['project']}",
        f"Root: {scan['root']}",
        f"Stale threshold: {scan['stale_threshold']} days",
        f"Archive threshold: {scan['archive_threshold']} days",
        "",
    ]
    for entry in scan["archive_pages"]:
        lines.append(
            f"[ARCHIVE] {entry['path']} — last updated {entry['freshness']} ({entry['days_old']} days ago, exceeds {scan['archive_threshold']}d)"
        )
    for entry in scan["stale_pages"]:
        lines.append(
            f"[STALE]   {entry['path']} — last updated {entry['freshness']} ({entry['days_old']} days ago)"
        )
    lines.extend(
        [
            "",
            f"Summary: {len(scan['stale_pages'])} stale, {len(scan['archive_pages'])} ready for archival",
            "",
            "To archive stale pages:",
            f"  mkdir -p \"{scan['archive_dir']}\"",
            f"  mv \"{scan['root']}/<page-path>\" \"{scan['archive_dir']}/\"",
            "  # Then update projects/<project>/index.md",
        ]
    )
    return "\n".join(lines)


def append_warmup_report(report_path: pathlib.Path, scan) -> None:
    with report_path.open("a", encoding="utf-8") as handle:
        for entry in scan["archive_pages"]:
            handle.write(
                f"[archive] {entry['path']} — last updated {entry['freshness']} ({entry['days_old']} days ago)\n"
            )
        for entry in scan["stale_pages"]:
            handle.write(
                f"[stale] {entry['path']} — last updated {entry['freshness']} ({entry['days_old']} days ago)\n"
            )
        handle.write(f"  stale_pages: {len(scan['stale_pages']) + len(scan['archive_pages'])}\n")
        handle.write(f"  archive_dir: {scan['archive_dir']}\n")


def main(argv: List[str]) -> int:
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest="command", required=True)

    kb_parser = subparsers.add_parser("kb-stale")
    kb_parser.add_argument("--config", required=True)

    warmup_parser = subparsers.add_parser("warmup")
    warmup_parser.add_argument("--config", required=True)
    warmup_parser.add_argument("--report-file", required=True)
    warmup_parser.add_argument("--hub-state")

    args = parser.parse_args(argv)
    try:
        scan = scan_pages(pathlib.Path(args.config))
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 1

    stale_paths = [entry["path"] for entry in (scan["stale_pages"] + scan["archive_pages"])]

    if args.command == "kb-stale":
        print(format_kb_stale(scan))
        return 0

    append_warmup_report(pathlib.Path(args.report_file), scan)
    if args.hub_state:
        update_hub_state(pathlib.Path(args.hub_state), stale_paths)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
