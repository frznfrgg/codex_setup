// Command bash-guard is a Codex Bash safety hook.
// It reads a JSON envelope on stdin, evaluates the wrapped Bash command
// against a set of rules, and emits Codex-native hook responses.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// HookInput is the JSON envelope sent by Codex-compatible hook runtimes.
type HookInput struct {
	HookEventName  string    `json:"hook_event_name"`
	ToolName       string    `json:"tool_name"`
	ToolInput      ToolInput `json:"tool_input"`
	SessionID      string    `json:"session_id"`
	TurnID         string    `json:"turn_id"`
	Cwd            string    `json:"cwd"`
	PermissionMode string    `json:"permission_mode"`
	Model          string    `json:"model"`
	Prompt         string    `json:"prompt"`
	Source         string    `json:"source"`
	AgentID        string    `json:"agent_id"`
	AgentType      string    `json:"agent_type"`
	StopHookActive bool      `json:"stop_hook_active"`
}

type ToolInput struct {
	Command string `json:"command"`
}

// CodexOutput is a generic Codex hook response envelope.
type CodexOutput struct {
	Decision           string        `json:"decision,omitempty"`
	Reason             string        `json:"reason,omitempty"`
	SystemMessage      string        `json:"systemMessage,omitempty"`
	HookSpecificOutput *HookSpecific `json:"hookSpecificOutput,omitempty"`
}

type HookSpecific struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
	AdditionalContext        string `json:"additionalContext,omitempty"`
}

func main() {
	start := time.Now()

	// --- self-test mode ---
	if len(os.Args) > 1 && os.Args[1] == "--selftest" {
		runSelfTest()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(versionString())
		return
	}

	// --- read input ---
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bash-guard: read stdin: %v\n", err)
		return
	}
	if strings.TrimSpace(string(rawInput)) == "" {
		return
	}
	var in HookInput
	if err := json.Unmarshal(rawInput, &in); err != nil {
		fmt.Fprintf(os.Stderr, "bash-guard: malformed JSON input: %v\n", err)
		return
	}

	switch in.HookEventName {
	case "SessionStart", "SubagentStart", "SubagentStop", "Stop":
		emit(*buildLifecycleOutput(in))
	case "UserPromptSubmit":
		handleUserPromptSubmit(in)
	case "PreToolUse", "":
		handlePreToolUse(in, start)
	default:
		if in.ToolName == "Bash" {
			handlePreToolUse(in, start)
		}
		return
	}
}

// buildLifecycleOutput adds compact, programmatic runtime context only where
// Codex supports model-visible additionalContext. Stop events deliberately
// return an empty JSON object; injecting a block decision would create a
// continuation prompt (and risks a stop-hook loop).
func buildLifecycleOutput(in HookInput) *CodexOutput {
	switch in.HookEventName {
	case "SessionStart", "SubagentStart":
		return &CodexOutput{
			HookSpecificOutput: &HookSpecific{
				HookEventName:     in.HookEventName,
				AdditionalContext: lifecycleContext(in),
			},
		}
	case "SubagentStop", "Stop":
		return &CodexOutput{}
	default:
		return &CodexOutput{}
	}
}

func lifecycleContext(in HookInput) string {
	fields := []string{"event=" + quoteContextValue(in.HookEventName)}
	appendField := func(name, value string) {
		if value != "" {
			fields = append(fields, name+"="+quoteContextValue(value))
		}
	}

	appendField("source", in.Source)
	appendField("agent_type", in.AgentType)
	appendField("agent_id", in.AgentID)
	appendField("cwd", in.Cwd)
	appendField("permission_mode", in.PermissionMode)
	appendField("model", in.Model)
	return "Codex runtime context: " + strings.Join(fields, " ")
}

func quoteContextValue(value string) string {
	// Keep hook-provided metadata on one short line. In particular, never pass
	// transcript or assistant-message contents through lifecycle context.
	value = strings.Join(strings.Fields(value), " ")
	runes := []rune(value)
	if len(runes) > 240 {
		value = string(runes[:237]) + "..."
	}
	return fmt.Sprintf("%q", value)
}

func handlePreToolUse(in HookInput, start time.Time) {
	out, audit := buildPreToolUseOutput(in, start)
	if out != nil {
		emit(*out)
	}
	if audit != nil {
		writeAudit(*audit)
	}
}

func buildPreToolUseOutput(in HookInput, start time.Time) (*CodexOutput, *AuditEntry) {
	if in.ToolName != "Bash" {
		return nil, nil
	}
	cmd := in.ToolInput.Command
	if cmd == "" {
		return nil, nil
	}

	// --- mode resolution ---
	binDir := selfDir()
	cfg := LoadGlobalConfig(filepath.Join(binDir, "config.toml"))
	mode := resolveCodexMode(cfg.Mode.Default)
	if mode == "off" || mode == "disabled" {
		return nil, nil
	}

	sessionID := in.SessionID
	if sessionID == "" {
		sessionID = "unknown-session"
	}
	fullCommandHash := fullHashCommand(cmd)

	if mode == "live" && consumeApproval(sessionID, fullCommandHash, cmd) {
		return nil, makeAuditEntry("approved_allow", in, start, mode, Decision{
			Level:      LevelAllow,
			Rule:       "approval",
			ReasonCode: "approval.consumed",
		}, cmd)
	}

	// --- registry ---
	reg := newRegistry(RmRule{}, SupabaseRule{}, InfraRule{}, PaasRule{}, DbClientRule{}, GitRule{})
	triggers := reg.triggerSet()

	// --- safe paths ---
	tp := LoadTrustedProjects(filepath.Join(binDir, "trusted-projects.toml"))
	projectExtras, _, untrustedNotice := LoadAndMergeProjectConfig(in.Cwd, tp)
	if untrustedNotice != "" {
		fmt.Fprintln(os.Stderr, untrustedNotice)
	}
	allExtras := append([]string(nil), cfg.SafePaths.Extra...)
	allExtras = append(allExtras, projectExtras...)
	sp := NewSafePaths(in.Cwd, allExtras)

	// --- pipeline ---
	decision := evaluate(cmd, triggers, reg, &RuleEnv{HookCwd: in.Cwd, SafePaths: sp})
	audit := makeAuditEntry("allow", in, start, mode, decision, cmd)
	if decision.Level == LevelAllow {
		return nil, audit
	}

	reason := decision.Reason
	if reason == "" {
		reason = "Risky Bash command."
	}
	intent := intentForReasonCode(decision.ReasonCode)
	if err := recordPending(sessionID, fullCommandHash, RiskEntry{
		Command:    cmd,
		ShortHash:  fullCommandHash[:shortHashLen],
		Reason:     reason,
		ReasonCode: decision.ReasonCode,
		Intent:     intent,
	}); err != nil {
		reason = reason + " Approval state could not be recorded: " + err.Error()
	}

	message := blockMessage(cmd, fullCommandHash, intent, reason)
	if mode == "shadow" || mode == "dry-run" {
		audit.Action = "would_block"
		audit.WouldDecide = decision.Level.String()
		return &CodexOutput{
			SystemMessage: "bash-guard shadow mode: risky Bash command would be blocked.",
			HookSpecificOutput: &HookSpecific{
				HookEventName:     "PreToolUse",
				AdditionalContext: message,
			},
		}, audit
	}

	audit.Action = "block"
	return &CodexOutput{
		HookSpecificOutput: &HookSpecific{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: message,
		},
	}, audit
}

func resolveCodexMode(defaultMode string) string {
	mode := defaultMode
	if mode == "" {
		mode = "live"
	}
	if v := os.Getenv("BASH_GUARD_CODEX_MODE"); v != "" {
		mode = strings.ToLower(strings.TrimSpace(v))
	}
	if v := os.Getenv("BASH_GUARD_SHADOW"); v != "" && v != "0" {
		mode = "shadow"
	}
	if v := os.Getenv("BASH_GUARD_DRY_RUN"); v != "" && v != "0" {
		mode = "dry-run"
	}
	return mode
}

func makeAuditEntry(action string, in HookInput, start time.Time, mode string, decision Decision, cmd string) *AuditEntry {
	entry := &AuditEntry{
		TS:          nowISO(),
		Action:      action,
		Mode:        mode,
		Decision:    decision.Level.String(),
		Rule:        decision.Rule,
		ReasonCode:  decision.ReasonCode,
		LatencyMS:   float64(time.Since(start).Microseconds()) / 1000.0,
		CommandHash: hashCommand(cmd),
		CommandLen:  len(cmd),
		SessionID:   in.SessionID,
		TurnID:      in.TurnID,
	}
	if os.Getenv("BASH_GUARD_LOG_COMMANDS") != "" {
		entry.Command = cmd
	}
	return entry
}

func blockMessage(command, commandHash, intent, reason string) string {
	shortHash := commandHash[:shortHashLen]
	return fmt.Sprintf(
		"Risky command requires explicit approval.\n\nIntent: %s.\nReason: %s\nCommand: %s\nCommand hash: %s\n\nTo approve this exact command once, reply:\napprove-risk %s",
		intent,
		reason,
		truncate(command, 1200),
		shortHash,
		shortHash,
	)
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

// evaluate runs the full pipeline: quick reject → parse → unwrap → rule eval → aggregate.
// Returns a Decision (always Allow or Ask).
func evaluate(cmd string, triggers []string, reg *registry, env *RuleEnv) Decision {
	if !quickReject(cmd, triggers) {
		// Quick reject: no trigger keyword. Allow.
		// (False negatives here are OK — parser would classify them safely.)
	} else {
		// Has at least one trigger keyword somewhere. Parse + classify.
		parsed, err := ParseCommand(cmd)
		if err != nil {
			// Asymmetric fail-open: trigger keyword present + parse error
			// → ask. Without a trigger we'd allow, but we already saw one.
			return Decision{
				Level:      LevelAsk,
				Rule:       "guard",
				Reason:     "could not parse command containing potentially destructive keyword",
				ReasonCode: "guard.parse_error_after_trigger",
				Context:    fmt.Sprintf("parse_error=%v", err),
			}
		}
		decisions := reg.evaluate(parsed, env)
		return Aggregate(decisions)
	}
	return Aggregate(nil)
}

// quickReject: false → no trigger keyword anywhere, allow without parsing.
//
// The trigger set is the union of:
//   - rule keywords (rm, unlink, rmdir, shred — actual rule targets)
//   - executor keywords (sudo, env, bash, sh, eval, xargs, find, ...) — words
//     that the parser must descend into to discover an underlying rule keyword.
//
// Without executor keywords here, `find /etc -delete` would skip parsing
// and silently allow.
func quickReject(cmd string, triggers []string) bool {
	all := append([]string(nil), triggers...)
	all = append(all, parserDescentKeywords...)
	if len(all) == 0 {
		return false
	}
	parts := make([]string, 0, len(all))
	for _, t := range all {
		parts = append(parts, regexp.QuoteMeta(t))
	}
	re := regexp.MustCompile(`\b(?:` + strings.Join(parts, "|") + `)\b`)
	return re.MatchString(cmd)
}

// parserDescentKeywords lists every command name that the parser must
// descend into (executors / shell evaluators / find / xargs) — these are
// the commands whose unwrap can SURFACE a rule trigger that wasn't visible
// at the top level.
//
// Filing them here (vs in the unwrap.go switch) keeps quickReject alone
// authoritative about "should we even parse?".
var parserDescentKeywords = []string{
	"sudo", "doas", "env", "command", "builtin", "exec",
	"time", "nice", "nohup", "ionice", "setsid",
	"timeout", "gtimeout", "chroot",
	"ssh",
	"bash", "sh", "zsh", "dash", "fish", "ksh", "ash", "busybox",
	"eval",
	"xargs", "find", "parallel", "watch", "flock",
}

func buildAdditionalContext(d Decision) string {
	// Allow decisions don't need additionalContext — the user never sees it
	// and it just bloats audit logs. Only attach context to ask decisions
	// where the user benefits from "why".
	if d.Level != LevelAsk {
		return ""
	}
	if d.ReasonCode == "" && d.Context == "" {
		return ""
	}
	parts := []string{"code:" + d.ReasonCode}
	if d.Context != "" {
		parts = append(parts, d.Context)
	}
	return strings.Join(parts, " ")
}

func emit(out CodexOutput) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(out)
}

func fullHashCommand(cmd string) string {
	h := sha256.Sum256([]byte(cmd))
	return hex.EncodeToString(h[:])
}

// selfDir returns the directory containing the running binary (or src dir
// during development). Used to locate config.toml and trusted-projects.toml.
func selfDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	return filepath.Dir(resolved)
}

func versionString() string {
	return "bash-guard 0.2.0"
}

// runSelfTest is a manual smoke check: it runs the parser on a couple of
// fixed commands and prints the resulting decision summary. Real coverage
// lives in *_test.go.
func runSelfTest() {
	cases := []struct{ name, cmd, cwd string }{
		{"FP-1 heredoc with apostrophes", "cat > /tmp/x <<'EOF'\nWe use find and rm a lot. Don't break.\nEOF", "/tmp"},
		{"FP-2 rm in /tmp", "cd /tmp && rm -rf ci-results && mkdir ci-results", "/home/example-user/myproject"},
		{"catastrophic /etc", "rm -rf /etc/nginx", "/home/example-user/myproject"},
		{"safe cwd subdir", "rm -rf node_modules", "/home/example-user/myproject"},
	}
	reg := newRegistry(RmRule{})
	triggers := reg.triggerSet()
	for _, c := range cases {
		sp := NewSafePaths(c.cwd, nil)
		d := evaluate(c.cmd, triggers, reg, &RuleEnv{HookCwd: c.cwd, SafePaths: sp})
		fmt.Printf("[%s] cwd=%s  ->  %s (rule=%s code=%s)\n", c.name, c.cwd, d.Level, d.Rule, d.ReasonCode)
	}
}
