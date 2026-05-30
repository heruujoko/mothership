const DEFAULT_ROLE_TEAM_MAP = {
  researcher: "research",
  coder: "implementation",
  qa: "review",
  pr_monkey: "default",
};

const SAFE_TEAM_TOOL_ALLOWLIST = new Set(["team"]);

function parseModelFallbackChain(raw) {
  if (Array.isArray(raw)) return raw.map((v) => `${v}`.trim()).filter(Boolean);
  if (typeof raw === "string") {
    return raw
      .split(",")
      .map((v) => v.trim())
      .filter(Boolean);
  }
  return [];
}

export function decidePiRoute({ role, config, availableTools }) {
  const pi = config?.pi || {};
  const teamTool = pi.team_tool || "team";
  const enabled = pi.team_path_enabled !== false;

  if (!enabled) {
    return { path: "legacy", reason: "team_path_disabled", teamTool };
  }

  const unsafeToolAllowed = pi.team_tool_allow_unsafe === true;
  if (!SAFE_TEAM_TOOL_ALLOWLIST.has(teamTool) && !unsafeToolAllowed) {
    return { path: "legacy", reason: "team_tool_unsafe", teamTool };
  }

  const teamName = pi.team || DEFAULT_ROLE_TEAM_MAP[role];
  if (!teamName) {
    return { path: "legacy", reason: "team_mapping_missing", teamTool };
  }

  if (!availableTools.includes(teamTool)) {
    return { path: "legacy", reason: "team_tool_unavailable", teamTool, teamName };
  }

  const modelFallbackChain = parseModelFallbackChain(pi.model_fallback_chain);
  return {
    path: "team",
    teamTool,
    teamName,
    modelHint: pi.model_hint,
    modelFallbackChain,
  };
}

export function buildTeamInvocation({ teamName, task, modelHint, modelFallbackChain }) {
  const payload = {
    action: "run",
    team: teamName,
    goal: task,
  };
  if (modelHint) payload.model = modelHint;
  if (Array.isArray(modelFallbackChain) && modelFallbackChain.length > 0) {
    payload.model_fallback_chain = modelFallbackChain;
  }
  return payload;
}

export { DEFAULT_ROLE_TEAM_MAP, SAFE_TEAM_TOOL_ALLOWLIST };
