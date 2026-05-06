/**
 * mothership_spawn - Sub-agent delegation tool for mothership on pi
 *
 * Spawns a pi subprocess to perform a delegated role (researcher, coder, qa,
 * pr_monkey) using the role contract from agents/*.md.
 */

import { spawn } from "node:child_process";
import * as fs from "node:fs";
import * as path from "node:path";
import type { Message } from "@mariozechner/pi-ai";
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";
import { Text } from "@mariozechner/pi-tui";
import { Type } from "typebox";

const VALID_ROLES = new Set(["researcher", "coder", "qa", "pr_monkey"]);

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

	const args: string[] = ["--mode", "json", "-p", "--no-session"];
	args.push("--append-system-prompt", contractPath);
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
			// Clear the SIGKILL timer if the process exits normally before it fires
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
						"Path to the role contract file (agents/<role>.md). If not provided, derived from the role name.",
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

			const contractFile =
				params.contract_file ||
				path.join(ctx.cwd, "agents", `${role}.md`);

			const contextNote = params.context
				? `\n\n---\n\nAdditional context:\n${params.context}\n\n---\n\n`
				: "";

			const enrichedTask = `${contextNote}${task}`;

			if (!fs.existsSync(contractFile)) {
				return {
					content: [
						{
							type: "text",
							text: `mothership_spawn failed: contract file not found: ${contractFile}`,
						},
					],
					details: { role, error: "contract_not_found" },
					isError: true,
				};
			}

			try {
				const result = await runPiSubprocess(
					contractFile,
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
						details: { role, ...result.usage, error: result.error },
						isError: true,
					};
				}

				return {
					content: [{ type: "text", text: result.output || "(no output)" }],
					details: { role, ...result.usage },
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
