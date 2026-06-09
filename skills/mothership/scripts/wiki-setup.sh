#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
project_name="$(basename "${repo_root}")"
wiki_root_default="${HOME}/.mothership/wiki"
pref_file="${repo_root}/.mothership/wiki.yaml"

echo "Mothership Wiki Setup"
echo "====================="
echo "Project: ${project_name}"
echo

read -rp "Wiki root [${wiki_root_default}]: " wiki_root
wiki_root="${wiki_root:-${wiki_root_default}}"
case "${wiki_root}" in
  "~")
    wiki_root="${HOME}"
    ;;
  "~/"*)
    wiki_root="${HOME}/${wiki_root#~/}"
    ;;
esac

yaml_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

wiki_root_yaml="$(yaml_escape "${wiki_root}")"
project_name_yaml="$(yaml_escape "${project_name}")"

project_dir="${wiki_root}/projects/${project_name}"
today="$(date +%Y-%m-%d)"

mkdir -p "${project_dir}/.draft" "${project_dir}/insights" "${project_dir}/decisions"

mkdir -p "$(dirname "${pref_file}")"
cat > "${pref_file}" <<EOF
wiki:
  root: "${wiki_root_yaml}"
  project: "${project_name_yaml}"
  auto_ingest: true
  draft_ttl: 2
EOF

schema_file="${wiki_root}/schema.md"
if [ ! -f "${schema_file}" ]; then
  schema_template="${repo_root}/mothership-config/wiki-schema-default.md"
  if [ -f "${schema_template}" ]; then
    cp "${schema_template}" "${schema_file}"
  else
    cat > "${schema_file}" <<'SCHEMA'
# Wiki Schema

This file defines conventions for reading and writing the mothership wiki.

## Writing Rules

- Insights pages use `## <topic>` sections within a single file
- Decisions use one file per decision with frontmatter
- Every L2 claim must cite a source: `> Source: PR #N (url)`
- Use `[[wikilinks]]` for cross-references between pages
- Append to existing pages, do not create new files for existing topics
- Every wiki page MUST include a YAML metadata header with:
  - sources (issue/PR/commit URLs)
  - confidence (draft|observed|inferred|verified|deprecated)
  - freshness (date last verified)
  - tags (categorical keywords)

## Metadata Template

Every new wiki page should be prepended with:
```
---
sources: []
confidence: draft
freshness: $(date +%Y-%m-%d)
tags: []
---
```

## L1 to L2 Promotion Rules

- An item must have at least one citation to promote
- Contradictions with existing L2 content are flagged, not silently overwritten
- If a new insight challenges an existing one, create a `## Contradiction` section in both pages

## Index Format

Each project `index.md` lists all insight pages and decision files with one-line summaries.

## Log Format

Entries use: `## [YYYY-MM-DD] ingest | <summary>` followed by promotion counts and citations.
SCHEMA
  fi
fi

global_index="${wiki_root}/index.md"
if [ ! -f "${global_index}" ]; then
  cat > "${global_index}" <<EOF
# Mothership Wiki Index

## Projects
- [[${project_name}]] — ${project_name} project wiki
EOF
fi

global_log="${wiki_root}/log.md"
if [ ! -f "${global_log}" ]; then
  echo "# Mothership Wiki Log" > "${global_log}"
fi

project_index="${project_dir}/index.md"
if [ ! -f "${project_index}" ]; then
  # Design Gap 2: Copy from wiki seed file if present
  wiki_seed="${repo_root}/docs/mothership-wiki-seed.md"
  if [ -f "${wiki_seed}" ]; then
    while IFS= read -r line || [ -n "${line}" ]; do
      line="${line//__PROJECT_NAME__/${project_name}}"
      line="${line//__TODAY__/${today}}"
      printf '%s\n' "${line}"
    done < "${wiki_seed}" > "${project_index}"
    echo "Initialized wiki from seed: ${wiki_seed}" >&2
  else
    cat > "${project_index}" <<EOF
---
sources: []
confidence: draft
freshness: ${today}
tags: [project-index]
---

# ${project_name} Index

## Insights
_(no insights yet)_

## Decisions
_(no decisions yet)_
EOF
  fi
fi

project_log="${project_dir}/log.md"
if [ ! -f "${project_log}" ]; then
  echo "# ${project_name} Log" > "${project_log}"
fi

timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
cat >> "${project_log}" <<EOF

## [${timestamp}] setup | Wiki initialized
Root: ${wiki_root}
Project: ${project_name}
EOF

cat >> "${global_log}" <<EOF

## [${timestamp}] setup | ${project_name} wiki created
Root: ${wiki_root}
EOF

echo
echo "Wiki setup complete:"
echo "  root: ${wiki_root}"
echo "  project: ${project_name}"
echo "  project dir: ${project_dir}"
echo "  preferences: ${pref_file}"
echo "  schema: ${schema_file}"
