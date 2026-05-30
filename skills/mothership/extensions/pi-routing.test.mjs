import test from "node:test";
import assert from "node:assert/strict";
import { decidePiRoute, buildTeamInvocation } from "./pi-routing.js";

test("routes to team when enabled and available", () => {
  const route = decidePiRoute({
    role: "coder",
    config: { pi: { team_path_enabled: true, team: "implementation", team_tool: "team" } },
    availableTools: ["team"],
  });
  assert.equal(route.path, "team");
  assert.equal(route.teamName, "implementation");
});

test("falls back to legacy when team tool unavailable", () => {
  const route = decidePiRoute({
    role: "qa",
    config: { pi: { team_path_enabled: true, team: "review", team_tool: "team" } },
    availableTools: [],
  });
  assert.equal(route.path, "legacy");
  assert.equal(route.reason, "team_tool_unavailable");
});

test("uses default team mapping", () => {
  const route = decidePiRoute({
    role: "researcher",
    config: { pi: { team_path_enabled: true, team_tool: "team" } },
    availableTools: ["team"],
  });
  assert.equal(route.path, "team");
  assert.equal(route.teamName, "research");
});

test("disallow config forces legacy with explicit reason", () => {
  const route = decidePiRoute({
    role: "pr_monkey",
    config: { pi: { team_path_enabled: false, team_tool: "team" } },
    availableTools: ["team"],
  });
  assert.equal(route.path, "legacy");
  assert.equal(route.reason, "team_path_disabled");
});

test("passes optional model hint", () => {
  const route = decidePiRoute({
    role: "coder",
    config: { pi: { team_path_enabled: true, team: "implementation", team_tool: "team", model_hint: "gpt-5" } },
    availableTools: ["team"],
  });
  assert.equal(route.path, "team");
  assert.equal(route.modelHint, "gpt-5");
});

test("supports configurable model fallback chain", () => {
  const route = decidePiRoute({
    role: "coder",
    config: {
      pi: {
        team_path_enabled: true,
        team: "implementation",
        team_tool: "team",
        model_hint: "gpt-5",
        model_fallback_chain: "gpt-5,gpt-4.1-mini",
      },
    },
    availableTools: ["team"],
  });
  assert.equal(route.path, "team");
  assert.deepEqual(route.modelFallbackChain, ["gpt-5", "gpt-4.1-mini"]);
});

test("rejects unsafe team tool unless explicitly sanctioned", () => {
  const unsafe = decidePiRoute({
    role: "coder",
    config: { pi: { team_path_enabled: true, team: "implementation", team_tool: "team --danger" } },
    availableTools: ["team --danger"],
  });
  assert.equal(unsafe.path, "legacy");
  assert.equal(unsafe.reason, "team_tool_unsafe");

  const sanctioned = decidePiRoute({
    role: "coder",
    config: {
      pi: {
        team_path_enabled: true,
        team: "implementation",
        team_tool: "team --danger",
        team_tool_allow_unsafe: true,
      },
    },
    availableTools: ["team --danger"],
  });
  assert.equal(sanctioned.path, "team");
});

test("builds team invocation with action run and model fallback chain", () => {
  const args = buildTeamInvocation({
    teamName: "implementation",
    task: "Implement feature X",
    modelHint: "gpt-5",
    modelFallbackChain: ["gpt-5", "gpt-4.1-mini"],
  });
  assert.deepEqual(args, {
    action: "run",
    team: "implementation",
    goal: "Implement feature X",
    model: "gpt-5",
    model_fallback_chain: ["gpt-5", "gpt-4.1-mini"],
  });
});
