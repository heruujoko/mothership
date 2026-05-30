/**
 * mothership_spawn - Sub-agent delegation tool for mothership on pi
 *
 * Spawns a pi subprocess to perform a delegated role (researcher, coder, qa,
 * pr_monkey) using the role contract from agents/*.md.
 *
 * Contract & registry resolution order:
 *   1. Explicit param (contract_file / registry_path)
 *   2. Repo-local <cwd>/agents/<file>
 *   3. User-scope  $PI_HOME/agents/<file>
 */

import { spawn } from "node:child_process";
import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";
import type { Message } from "@mariozechner/pi-ai";
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { Text } from "@mariozechner/pi-tui";
import { Type } from "typebox";
import { buildTeamInvocation, decidePiRoute } from "./pi-routing.js";

const VALID_ROLES = new Set(["researcher", "coder", "qa", "pr_monkey"]);

/** Resolve PI_HOME: env var or default ~/.agents */
function getPiHome(): string {
	return process.env.PI_HOME || path.join(os.homedir(), ".agents");
}

/**
 * Resolve a contract file path.
 * Order: explicit param → repo-local → user-scope ($PI_HOME/agents/)
 */
function resolveContractFile(
	role: string,
	explicitPath: string | undefined,
	cwd: string,
): { path: string; source: "explicit" | "repo-local" | "user-scope"; triedPaths: string[] } {
	const triedPaths: string[] = [];

	// 1. Explicit param
	if (explicitPath) {
		triedPaths.push(explicitPath);
		if (fs.existsSync(explicitPath)) {
			return { path: explicitPath, source: "explicit", triedPaths };
		}
	}

	// 2. Repo-local
	const repoPath = path.join(cwd, "agents", `${role}.md`);
	triedPaths.push(`repo-local: ${repoPath}`);
	if (fs.existsSync(repoPath)) {
		return { path: repoPath, source: "repo-local", triedPaths };
	}

	// 3. User-scope
	const piHome = getPiHome();
	const userPath = path.join(piHome, "agents", `${role}.md`);
	triedPaths.push(`user-scope: ${userPath}`);
	if (fs.existsSync(userPath)) {
		return { path: userPath, source: "user-scope", triedPaths };
	}

	return { path: repoPath, source: "repo-local", triedPaths };
}

/**
 * Resolve registry.yaml path.
 * Order: repo-local → user-scope ($PI_HOME/agents/)
 */
function resolveRegistryPath(cwd: string): string {
	const repoLocal = path.join(cwd, "agents", "registry.yaml");
	if (fs.existsSync(repoLocal)) return repoLocal;
	const piHome = getPiHome();
	return path.join(piHome, "agents", "registry.yaml");
}

interface MothershipSpawnResult {
	role: string;
	output: string;
	usage: {
		inputTokens: number;
		outputTokens: number;
		costTotal: number;
		turns: number;
		model?: string;
		stopReason?: string;
	};
	error?: string;
}

function getFinalOutput(messages: Message[]): string {
	for (let i = messages.length - 1; i >= 0; i--) {
		const msg = messages[i];
		if (msg.role === "assistant") {
			for (const part of msg.content) {
				if (part.type === "text") return part.text;
			}
		}
	}
	return "";
}

/**
 * Parse role config from registry.yaml using section-scoped extraction.
 * First isolates the `roles:` block, then extracts the specific role entry
 * within that block to avoid cross-section false positives.
 * Also extracts the `contract_file` field for Design Gap 3 resolution.
 */
function parseRoleConfigFromRegistry(
	registryPath: string,
	role: string,
): { pi: Record<string, any>; contract_file?: string } {
	const result: { pi: Record<string, any>; contract_file?: string } = { pi: {} };

	if (!fs.existsSync(registryPath)) return result;

	const content = fs.readFileSync(registryPath, "utf8");

	// Extract the `roles:` top-level section first, then search within it.
	// This avoids matching role names that appear in other sections.
	const rolesSection = extractYamlSection(content, "roles");
	if (!rolesSection) return result;

	// Within the roles section, find the specific role entry (indented 2 spaces).
	const roleRegex = new RegExp(`\\n\\s{2}${role}:([\\s\\S]*?)(?=\\n\\s{2}[a-z_]\\w*:|$)`);
	const roleMatch = rolesSection.match(roleRegex);
	if (!roleMatch) return result;

	const block = roleMatch[1];

	// Extract contract_file if present (Design Gap 3)
	const cfMatch = block.match(/\n\s{4}contract_file:\s*(.+)/);
	if (cfMatch) {
		result.contract_file = cfMatch[1].trim().replace(/^["']|["']$/g, "");
	}

	// Extract pi: block (4-space indented within the role)
	const piMatch = block.match(/\n\s{4}pi:([\s\S]*?)(?=\n\s{4}[a-z_]\w*:|$)/);
	if (!piMatch) return result;

	const piBlock = piMatch[1];
	const pi: Record<string, any> = {};
	for (const rawLine of piBlock.split("\n")) {
		const line = rawLine.trim();
		if (!line || line.startsWith("#")) continue;
		const idx = line.indexOf(":");
		if (idx <= 0) continue;
		const key = line.slice(0, idx).trim();
		let value: any = line.slice(idx + 1).trim();
		if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
			value = value.slice(1, -1);
		}
		if (value === "true") value = true;
		else if (value === "false") value = false;
		pi[key] = value;
	}
	result.pi = pi;
	return result;
}

/**
 * Extract a YAML top-level section by key (e.g., "roles", "host_models").
 * Returns the content from the key line to the next top-level key or end-of-file.
 * Lines where the key appears nested under other keys are ignored.
 */
function extractYamlSection(content: string, sectionKey: string): string | null {
	// Match a top-level key (no leading whitespace) followed by ":"
	// Capture everything until the next top-level key or end of string.
	const regex = new RegExp(`^${sectionKey}:([\\s\\S]*?)(?=^[a-z_]\\w*:|\\z)`, "m");
	const match = content.match(regex);
	return match ? match[0] : null;
}

function getAvailableToolNames(ctx: any): string[] {
	const names = new Set<string>();
	const candidates = [ctx?.tools, ctx?.runtime?.tools, ctx?.availableTools];
	for (const c of candidates) {
		if (Array.isArray(c)) {
			for (const t of c) {
				if (typeof t === "string") names.add(t);
				else if (t?.name) names.add(t.name);
			}
		} else if (c && typeof c === "object") {
			for (const key of Object.keys(c)) {
				// Only include keys whose value is callable (function) or is an
				// object with callable shape — filters out prototype noise,
				// string metadata keys, and non-tool object properties.
				const val = c[key];
				if (typeof val === "function" || (val && typeof val === "object")) {
					names.add(key);
				}
			}
		}
	}
	return [...names];
}

async function runPiSubprocess(
	contractPath: string,
	task: string,
	cwd: string,
	signal: AbortSignal | undefined,
	role: string,
): Promise<MothershipSpawnResult> {
	const result: MothershipSpawnResult = {
		role,
		output: "",
		usage: { inputTokens: 0, outputTokens: 0, costTotal: 0, turns: 0 },
	};

	// Bug 4 fix: Read contract file content and pass it inline
	let contractContent: string;
	try {
		contractContent = fs.readFileSync(contractPath, "utf8");
	} catch {
		result.error = `Cannot read contract file: ${contractPath}`;
		return result;
	}

	const args: string[] = ["--mode", "json", "-p", "--no-session"];
	args.push("--append-system-prompt", contractContent);
	args.push(`Task: ${task}`);

	let wasAborted = false;
	let stderrBuffer = "";
	const messages: Message[] = [];

	const exitCode = await new Promise<number>((resolve) => {
		const proc = spawn("pi", args, {
			cwd,
			shell: false,
			stdio: ["ignore", "pipe", "pipe"],
		});

		let stdoutBuffer = "";

		const processLine = (line: string) => {
			if (!line.trim()) return;
			let event: any;
			try {
				event = JSON.parse(line);
			} catch {
				return;
			}

			if (event.type === "message_end" && event.message) {
				const msg = event.message as Message;
				messages.push(msg);

				if (msg.role === "assistant") {
					result.usage.turns++;
					const usage = msg.usage as any;
					if (usage) {
						result.usage.inputTokens += usage.input || 0;
						result.usage.outputTokens += usage.output || 0;
						result.usage.costTotal += usage.cost?.total || 0;
					}
					if (!result.usage.model && (msg as any).model) result.usage.model = (msg as any).model;
					if ((msg as any).stopReason) result.usage.stopReason = (msg as any).stopReason;
				}
			}

			if (event.type === "tool_result_end" && event.message) {
				messages.push(event.message as Message);
			}
		};

		proc.stdout.on("data", (data) => {
			stdoutBuffer += data.toString();
			const lines = stdoutBuffer.split("\n");
			stdoutBuffer = lines.pop() || "";
			for (const line of lines) processLine(line);
		});

		proc.stderr.on("data", (data) => {
			stderrBuffer += data.toString();
		});

		proc.on("close", (code) => {
			if (stdoutBuffer.trim()) processLine(stdoutBuffer);
			resolve(code ?? 0);
		});

		proc.on("error", () => {
			resolve(1);
		});

		if (signal) {
			let sigkillTimer: ReturnType<typeof setTimeout> | undefined;
			const killProc = () => {
				wasAborted = true;
				proc.kill("SIGTERM");
				sigkillTimer = setTimeout(() => {
					if (!proc.killed) proc.kill("SIGKILL");
				}, 5000);
			};
			if (signal.aborted) killProc();
			else signal.addEventListener("abort", killProc, { once: true });
			proc.on("close", () => {
				if (sigkillTimer !== undefined) clearTimeout(sigkillTimer);
			});
		}
	});

	result.output = getFinalOutput(messages);

	if (wasAborted) {
		result.error = "Aborted by user";
	} else if (exitCode !== 0) {
		const stderr = stderrBuffer.trim();
		result.error = `Sub-process exited with code ${exitCode}${stderr ? `: ${stderr}` : ""}`;
	}

	return result;
}

export default function (pi: ExtensionAPI) {
	pi.registerTool({
		name: "mothership_spawn",
		label: "Mothership Spawn",
		description:
			"Delegate a task to a mothership role (researcher, coder, qa, pr_monkey) by spawning a new pi subprocess with the role contract. Use when mothership commander needs to delegate work to a sub-agent. Returns the sub-agent's final text output.",
		parameters: Type.Object({
			role: Type.String({
				description:
					"Mothership role to invoke: researcher, coder, qa, or pr_monkey. Determines which agent contract to use.",
			}),
			task: Type.String({
				description:
					"The task or question to delegate to the sub-agent. Should be a clear, self-contained request.",
			}),
			contract_file: Type.Optional(
				Type.String({
					description:
						"Explicit path to the role contract file. If not provided, resolved from repo-local agents/<role>.md or user-scope $PI_HOME/agents/<role>.md.",
				}),
			),
			registry_path: Type.Optional(
				Type.String({
					description:
						"Explicit path to registry.yaml. If not provided, resolved from repo-local or user-scope agents/registry.yaml.",
				}),
			),
			context: Type.Optional(
				Type.String({
					description:
						"Additional context to pass to the sub-agent. Can include file paths, issue numbers, branch name, or other relevant details.",
				}),
			),
			cwd: Type.Optional(
				Type.String({
					description: "Working directory for the subprocess. Defaults to the current pi working directory.",
				}),
			),
		}),

		async execute(_toolCallId, params, signal, _onUpdate, ctx) {
			const role = params.role;
			const task = params.task ?? "";
			const cwd = params.cwd ?? ctx.cwd;

			// Validate role before any work
			if (!VALID_ROLES.has(role)) {
				return {
					content: [
						{
							type: "text",
							text: `mothership_spawn failed: unknown role "${role}". Valid roles: ${[...VALID_ROLES].join(", ")}`,
						},
					],
					details: { role, error: "invalid_role" },
					isError: true,
				};
			}

			// --- Bug 1 fix: Resolve contract file with user-scope fallback ---
			const resolved = resolveContractFile(role, params.contract_file, cwd);
			const contractFile = resolved.path;

			if (!fs.existsSync(contractFile)) {
				const triedList = resolved.triedPaths.join("\n  - ");
				return {
					content: [
						{
							type: "text",
							text: `mothership_spawn failed: contract file not found for role "${role}".\n\nTried:\n  - ${triedList}\n\nEnsure the installer ran (\`./install.sh --target pi\`) and the contract file exists at one of those paths.`,
						},
					],
					details: { role, error: "contract_not_found", tried_paths: resolved.triedPaths },
					isError: true,
				};
			}

			const contextNote = params.context
				? `\n\n---\n\nAdditional context:\n${params.context}\n\n---\n\n`
				: "";

			const enrichedTask = `${contextNote}${task}`;

			try {
				// --- Design Gap 3 fix: Read contract_file from registry as source of truth ---
				const registryPath = params.registry_path || resolveRegistryPath(cwd);
				const roleConfig = parseRoleConfigFromRegistry(registryPath, role);

				// If registry specifies a contract_file, prefer that as primary contract path
				// (with same fallback logic). This solves Bug 2 (pr-monkey vs pr_monkey) since
				// the registry can point to whatever filename exists.
				let effectiveContractFile = contractFile;
				if (roleConfig.contract_file) {
					const registryContractFile = path.isAbsolute(roleConfig.contract_file)
						? roleConfig.contract_file
						: path.resolve(path.dirname(registryPath), roleConfig.contract_file);
					if (fs.existsSync(registryContractFile)) {
						effectiveContractFile = registryContractFile;
					}
				}

				const route = decidePiRoute({ role, config: roleConfig, availableTools: getAvailableToolNames(ctx as any) });

				// --- Bug 3 guard: Detect known pi-crew CLI flag issues ---
				// The `--exclude-tools` flag is unsupported on some pi CLI versions.
				// When team tool is potentially affected, log a clear diagnostic.
				const knownUnsupportedFlags = ["--exclude-tools"];
				let teamPathSafe = route.path === "team";
				if (teamPathSafe) {
					const callTool = (ctx as any)?.callTool || (ctx as any)?.executeTool || (pi as any)?.callTool;
					if (typeof callTool !== "function") {
						console.warn(
							`[mothership_spawn] fallback=legacy reason=team_tool_unavailable tool=${route.teamTool} role=${role}`,
						);
						teamPathSafe = false;
					}
				}

				if (teamPathSafe && route.path === "team") {
					const payload = buildTeamInvocation({
						teamName: route.teamName,
						task: enrichedTask,
						modelHint: route.modelHint,
						modelFallbackChain: route.modelFallbackChain,
					});
					try {
						const callTool =
							(ctx as any)?.callTool || (ctx as any)?.executeTool || (pi as any)?.callTool;
						const teamResult = await callTool(route.teamTool, payload, signal);
						if (teamResult?.isError) {
							const errText = Array.isArray(teamResult.content)
								? teamResult.content.map((c: any) => c?.text || "").join(" ")
								: "";
							const isExcludeTools =
								errText.includes("--exclude-tools") || errText.includes("Unknown option");
							console.warn(
								`[mothership_spawn] fallback=legacy reason=team_tool_unhealthy tool=${route.teamTool} role=${role}` +
									(isExcludeTools
										? ` (detected: pi-crew CLI flag issue — check pi/pi-crew compatibility for ${knownUnsupportedFlags})`
										: ""),
							);
						} else {
							const output = Array.isArray(teamResult?.content)
								? teamResult.content.map((c: any) => c?.text || "").join("\n").trim()
								: "";
							return {
								content: [{ type: "text", text: output || "(no output)" }],
								details: {
									role,
									route: "team",
									team: route.teamName,
									team_tool: route.teamTool,
									contract_source: resolved.source,
									model_hint: route.modelHint,
									model_fallback_chain: route.modelFallbackChain,
								},
							};
						}
					} catch (err: any) {
						const errMsg = err?.message || "";
						const isExcludeTools =
							errMsg.includes("--exclude-tools") || errMsg.includes("Unknown option");
						console.warn(
							`[mothership_spawn] fallback=legacy reason=team_tool_unhealthy tool=${route.teamTool} role=${role}` +
								(isExcludeTools
									? ` (detected: pi-crew CLI flag issue — pi CLI may not support ${knownUnsupportedFlags}. Try updating pi or using --no-team-fallback.)`
									: ""),
						);
					}
				} else if (route.path !== "team") {
					console.warn(
						`[mothership_spawn] fallback=legacy reason=${route.reason} role=${role}`,
					);
				}

				const result = await runPiSubprocess(
					effectiveContractFile,
					enrichedTask,
					cwd,
					signal,
					role,
				);

				if (result.error) {
					return {
						content: [
							{
								type: "text",
								text: `[${role}] ${result.error}${result.output ? `\n\nOutput:\n${result.output}` : ""}`,
							},
						],
						details: { role, ...result.usage, error: result.error, contract_source: resolved.source },
						isError: true,
					};
				}

				return {
					content: [{ type: "text", text: result.output || "(no output)" }],
					details: { role, ...result.usage, contract_source: resolved.source },
				};
			} catch (err: any) {
				return {
					content: [{ type: "text", text: `mothership_spawn failed: ${err.message}` }],
					details: { role, error: err.message },
					isError: true,
				};
			}
		},

		renderCall(args, theme, _context) {
			const roleName = args.role || "...";
			const taskPreview = args.task
				? args.task.length > 60
					? args.task.slice(0, 60) + "..."
					: args.task
				: "...";
			const label =
				theme.fg("toolTitle", theme.bold("mothership_spawn ")) +
				theme.fg("accent", roleName);
			return new Text(label + "\n  " + theme.fg("dim", taskPreview), 0, 0);
		},
	});
}
