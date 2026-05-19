package main

import (
	"strings"
	"testing"
	"time"
)

func testHookInput(t *testing.T, sessionID, command string) HookInput {
	t.Helper()
	return HookInput{
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     ToolInput{Command: command},
		SessionID:     sessionID,
		TurnID:        "turn-test",
		Cwd:           t.TempDir(),
	}
}

func configureApprovalState(t *testing.T) {
	t.Helper()
	t.Setenv("BASH_GUARD_CODEX_STATE_DIR", t.TempDir())
	t.Setenv("BASH_GUARD_CODEX_MODE", "live")
	t.Setenv("BASH_GUARD_SHADOW", "0")
	t.Setenv("BASH_GUARD_DRY_RUN", "0")
}

func TestCodexPreToolUseAllowsSafeCommandSilently(t *testing.T) {
	configureApprovalState(t)
	out, audit := buildPreToolUseOutput(testHookInput(t, "sess-safe", "echo ok"), time.Now())
	if out != nil {
		t.Fatalf("safe command produced output: %#v", out)
	}
	if audit == nil || audit.Decision != "allow" {
		t.Fatalf("safe command audit = %#v, want allow audit", audit)
	}
}

func TestCodexPreToolUseBlocksRiskyCommand(t *testing.T) {
	configureApprovalState(t)
	out, audit := buildPreToolUseOutput(testHookInput(t, "sess-risk", "git reset --hard"), time.Now())
	if out == nil || out.HookSpecificOutput == nil {
		t.Fatalf("risky command produced no output")
	}
	hook := out.HookSpecificOutput
	if hook.PermissionDecision != "deny" {
		t.Fatalf("permissionDecision = %q, want deny", hook.PermissionDecision)
	}
	reason := hook.PermissionDecisionReason
	for _, want := range []string{
		"Intent: discard local uncommitted changes",
		"Command: git reset --hard",
		"approve-risk",
	} {
		if !strings.Contains(reason, want) {
			t.Fatalf("reason missing %q:\n%s", want, reason)
		}
	}
	if audit == nil || audit.Action != "block" || audit.ReasonCode != "git.reset_hard" {
		t.Fatalf("audit = %#v, want git.reset_hard block", audit)
	}
}

func TestCodexApprovalIsSessionScopedAndOneShot(t *testing.T) {
	configureApprovalState(t)
	command := "git reset --hard"
	sessionID := "sess-approval"
	input := testHookInput(t, sessionID, command)
	input.Cwd = t.TempDir()

	blocked, _ := buildPreToolUseOutput(input, time.Now())
	if blocked == nil {
		t.Fatal("first risky command was not blocked")
	}

	hashPrefix := fullHashCommand(command)[:shortHashLen]
	approval := buildUserPromptSubmitOutput(HookInput{
		HookEventName: "UserPromptSubmit",
		SessionID:     sessionID,
		Prompt:        "approve-risk " + hashPrefix,
	})
	if approval == nil || approval.HookSpecificOutput == nil {
		t.Fatalf("approval output = %#v, want additional context", approval)
	}
	if !strings.Contains(approval.HookSpecificOutput.AdditionalContext, command) {
		t.Fatalf("approval context missing command: %#v", approval.HookSpecificOutput.AdditionalContext)
	}

	allowed, _ := buildPreToolUseOutput(input, time.Now())
	if allowed != nil {
		t.Fatalf("approved exact command produced output: %#v", allowed)
	}

	blockedAgain, _ := buildPreToolUseOutput(input, time.Now())
	if blockedAgain == nil || blockedAgain.HookSpecificOutput == nil ||
		blockedAgain.HookSpecificOutput.PermissionDecision != "deny" {
		t.Fatalf("second retry = %#v, want blocked again", blockedAgain)
	}
}

func TestCodexApprovalRejectsInvalidOrWrongSessionPrompt(t *testing.T) {
	configureApprovalState(t)
	command := "git reset --hard"
	buildPreToolUseOutput(testHookInput(t, "sess-original", command), time.Now())

	invalid := buildUserPromptSubmitOutput(HookInput{
		HookEventName: "UserPromptSubmit",
		SessionID:     "sess-original",
		Prompt:        "approve-risk nope",
	})
	if invalid == nil || invalid.Decision != "block" {
		t.Fatalf("invalid approval = %#v, want block", invalid)
	}

	wrongSession := buildUserPromptSubmitOutput(HookInput{
		HookEventName: "UserPromptSubmit",
		SessionID:     "sess-other",
		Prompt:        "approve-risk " + fullHashCommand(command)[:shortHashLen],
	})
	if wrongSession == nil || wrongSession.Decision != "block" {
		t.Fatalf("wrong-session approval = %#v, want block", wrongSession)
	}
}

func TestCodexShadowModeAllowsWithContext(t *testing.T) {
	configureApprovalState(t)
	t.Setenv("BASH_GUARD_CODEX_MODE", "shadow")
	out, audit := buildPreToolUseOutput(testHookInput(t, "sess-shadow", "git reset --hard"), time.Now())
	if out == nil || out.HookSpecificOutput == nil {
		t.Fatalf("shadow output = %#v, want context", out)
	}
	if out.HookSpecificOutput.PermissionDecision != "" {
		t.Fatalf("shadow permissionDecision = %q, want empty", out.HookSpecificOutput.PermissionDecision)
	}
	if !strings.Contains(out.HookSpecificOutput.AdditionalContext, "Risky command requires explicit approval") {
		t.Fatalf("shadow context = %q", out.HookSpecificOutput.AdditionalContext)
	}
	if audit == nil || audit.Action != "would_block" {
		t.Fatalf("shadow audit = %#v, want would_block", audit)
	}
}

func TestCodexUserPromptIgnoresUnrelatedPrompt(t *testing.T) {
	configureApprovalState(t)
	out := buildUserPromptSubmitOutput(HookInput{
		HookEventName: "UserPromptSubmit",
		SessionID:     "sess-prompt",
		Prompt:        "please continue",
	})
	if out != nil {
		t.Fatalf("unrelated prompt output = %#v, want nil", out)
	}
}
