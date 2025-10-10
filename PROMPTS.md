# Prompting & Tool-Call Conventions

This file is a concise reference for writing role prompts and tool-call JSON for the ai-team CLI. It duplicates and extracts the guidance from `README.md` so non-developers can quickly follow the canonical format.

1. General rules

---

- Return exactly one JSON object when you intend to request a tool call. No extra text, no markdown, no status messages.
- Prefer snake_case for tool names and argument keys: `write_file`, `read_file`, `list_dir`, `run_command`, `apply_patch`.
- Canonical tool-call structure:

  {"tool_call": {"name": "<tool_name>", "arguments": {<arg_key>: <arg_value>}}}

2. Canonical examples

---

- Write file (preferred):
  {"tool_call": {"name": "write_file", "arguments": {"file_path": "ai-team-data/design.md", "content": "# Design\n..."}}}

- Read file:
  {"tool_call": {"name": "read_file", "arguments": {"file_path": "ai-team-data/design.md"}}}

- List directory:
  {"tool_call": {"name": "list_dir", "arguments": {"path": "."}}}

- Run a command:
  {"tool_call": {"name": "run_command", "arguments": {"command": "go test ./..."}}}

- Apply a patch:
  {"tool_call": {"name": "apply_patch", "arguments": {"file_path": "ai-team-data/design.md", "patch_content": "@@ -1 +1 @@\n-Old\n+New\n"}}}

3. lastToolResponse (what you get after a tool runs)

---

- `lastToolResponse`: the structured result of the last executed tool (map/list/string).
- `lastToolResponse_json`: JSON string of the above (useful for template rendering where a string is required).

Tips for prompts:

- Use `{{.lastToolResponse_json}}` when you expect the model to parse or reason about the previous result as text.
- To check for execution errors in templates: `{{if .lastToolResponse.error}} ... {{end}}` (the executor sets an `error` key on failures).

4. Looping and `loop_condition`

---

- Use `loop: true` and set `loop_count` when you want a role to iterate.
- Use `loop_condition` to stop early by providing a Go-template expression that will be rendered against the role context after every iteration.

Example (stop when a `write_file` is produced):

loop: true
loop_count: 10
loop_condition: "{{.tool_call.name}} == 'write_file'"

Supported evaluation forms (safe, limited):

- literal `true` / `false`
- equality: `{{.a}} == 'b'`
- inequality: `{{.a}} != 'b'`

If the rendered `loop_condition` evaluates to true, the loop exits early. For complex expressions (numeric comparisons, logical AND/OR) contact the dev team — we can add a small expression evaluator or extend the current implementation.

5. Prompt authoring tips

---

- Provide 1–2 canonical examples in the prompt; models usually follow examples closely.
- Keep the instruction "Respond ONLY with a single JSON tool_call object" early and explicit.
- If you need the model to inspect files, prefer issuing `list_dir` then `read_file` calls in separate iterations.

6. Troubleshooting

---

- If you see `No valid tool-call found in response`, enable debug logging:

  AI_TEAM_DEBUG=1 ./ai-team run-chain <chain> --input "..."

- Inspect logs for the extractor messages (it prints original vs normalized tool-call payloads when available).

That's it — use this file as a quick checklist when writing role prompts.
