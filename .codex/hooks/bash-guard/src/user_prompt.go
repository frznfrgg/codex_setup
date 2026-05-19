package main

import (
	"fmt"
	"regexp"
	"strings"
)

var approvalPromptRE = regexp.MustCompile(`^\s*approve-risk\s+([0-9a-fA-F]{8,64})\s*$`)

func handleUserPromptSubmit(in HookInput) {
	out := buildUserPromptSubmitOutput(in)
	if out != nil {
		emit(*out)
	}
}

func buildUserPromptSubmitOutput(in HookInput) *CodexOutput {
	if !strings.Contains(in.Prompt, "approve-risk") {
		return nil
	}

	match := approvalPromptRE.FindStringSubmatch(in.Prompt)
	if match == nil {
		return blockPrompt("Invalid risk approval prompt. Use exactly: approve-risk <command-hash>")
	}

	sessionID := in.SessionID
	if sessionID == "" {
		sessionID = "unknown-session"
	}
	result := approvePrefix(sessionID, match[1])
	if !result.OK {
		return blockPrompt(result.Reason)
	}

	shortHash := result.Entry.ShortHash
	if shortHash == "" && len(result.CommandHash) >= shortHashLen {
		shortHash = result.CommandHash[:shortHashLen]
	}
	context := fmt.Sprintf(
		"The user approved one risky Bash command for this Codex session. Retry the exact same command once; do not alter it before retrying.\n\nApproved command hash: %s\nApproved command: %s",
		shortHash,
		result.Entry.Command,
	)
	return &CodexOutput{
		SystemMessage: fmt.Sprintf("Risk approval recorded for %s.", shortHash),
		HookSpecificOutput: &HookSpecific{
			HookEventName:     "UserPromptSubmit",
			AdditionalContext: context,
		},
	}
}

func blockPrompt(reason string) *CodexOutput {
	return &CodexOutput{
		Decision: "block",
		Reason:   reason,
	}
}
