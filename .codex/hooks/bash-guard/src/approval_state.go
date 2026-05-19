package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	approvalTTLSeconds = 30 * 60
	minHashPrefix      = 8
	shortHashLen       = 12
)

type RiskEntry struct {
	Command    string  `json:"command"`
	ShortHash  string  `json:"short_hash"`
	Reason     string  `json:"reason"`
	ReasonCode string  `json:"reason_code"`
	Intent     string  `json:"intent"`
	CreatedAt  float64 `json:"created_at,omitempty"`
	ExpiresAt  float64 `json:"expires_at,omitempty"`
	ApprovedAt float64 `json:"approved_at,omitempty"`
}

type approvalState struct {
	Version  int                             `json:"version"`
	Pending  map[string]map[string]RiskEntry `json:"pending"`
	Approved map[string]map[string]RiskEntry `json:"approved"`
}

type approvalResult struct {
	OK          bool
	Reason      string
	CommandHash string
	Entry       RiskEntry
}

func recordPending(sessionID, commandHash string, risk RiskEntry) error {
	return mutateApprovalState(func(state *approvalState, ts float64) error {
		if state.Pending[sessionID] == nil {
			state.Pending[sessionID] = map[string]RiskEntry{}
		}
		risk.CreatedAt = ts
		risk.ExpiresAt = ts + approvalTTLSeconds
		state.Pending[sessionID][commandHash] = risk
		return nil
	})
}

func approvePrefix(sessionID, hashPrefix string) approvalResult {
	prefix := strings.ToLower(hashPrefix)
	if len(prefix) < minHashPrefix {
		return approvalResult{OK: false, Reason: fmt.Sprintf("Approval hash must include at least %d hex characters.", minHashPrefix)}
	}

	var result approvalResult
	err := mutateApprovalState(func(state *approvalState, ts float64) error {
		pending := state.Pending[sessionID]
		var matches []string
		for commandHash := range pending {
			if strings.HasPrefix(commandHash, prefix) {
				matches = append(matches, commandHash)
			}
		}
		switch len(matches) {
		case 0:
			result = approvalResult{OK: false, Reason: "No pending risky command matches that hash in this Codex session."}
			return nil
		case 1:
			commandHash := matches[0]
			entry := pending[commandHash]
			delete(pending, commandHash)
			if state.Approved[sessionID] == nil {
				state.Approved[sessionID] = map[string]RiskEntry{}
			}
			entry.ApprovedAt = ts
			entry.ExpiresAt = ts + approvalTTLSeconds
			state.Approved[sessionID][commandHash] = entry
			result = approvalResult{OK: true, CommandHash: commandHash, Entry: entry}
			return nil
		default:
			result = approvalResult{OK: false, Reason: "Approval hash is ambiguous. Use more hex characters."}
			return nil
		}
	})
	if err != nil {
		return approvalResult{OK: false, Reason: err.Error()}
	}
	return result
}

func consumeApproval(sessionID, commandHash, command string) bool {
	consumed := false
	_ = mutateApprovalState(func(state *approvalState, _ float64) error {
		approved := state.Approved[sessionID]
		entry, ok := approved[commandHash]
		if !ok || entry.Command != command {
			return nil
		}
		delete(approved, commandHash)
		consumed = true
		return nil
	})
	return consumed
}

func mutateApprovalState(fn func(*approvalState, float64) error) error {
	if err := os.MkdirAll(approvalStateDir(), 0o700); err != nil {
		return err
	}
	lock, err := os.OpenFile(approvalLockPath(), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lock.Close()
	_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	ts := float64(time.Now().Unix())
	state := loadApprovalState()
	cleanupApprovalState(state, ts)
	if err := fn(state, ts); err != nil {
		return err
	}
	return saveApprovalState(state)
}

func loadApprovalState() *approvalState {
	state := emptyApprovalState()
	data, err := os.ReadFile(approvalStatePath())
	if err != nil {
		return state
	}
	if err := json.Unmarshal(data, state); err != nil {
		return emptyApprovalState()
	}
	if state.Pending == nil {
		state.Pending = map[string]map[string]RiskEntry{}
	}
	if state.Approved == nil {
		state.Approved = map[string]map[string]RiskEntry{}
	}
	if state.Version == 0 {
		state.Version = 1
	}
	return state
}

func emptyApprovalState() *approvalState {
	return &approvalState{
		Version:  1,
		Pending:  map[string]map[string]RiskEntry{},
		Approved: map[string]map[string]RiskEntry{},
	}
}

func saveApprovalState(state *approvalState) error {
	tmp, err := os.CreateTemp(approvalStateDir(), "state.*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(state); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, approvalStatePath())
}

func cleanupApprovalState(state *approvalState, ts float64) {
	cleanupBucket := func(bucket map[string]map[string]RiskEntry) {
		for sessionID, commands := range bucket {
			for commandHash, entry := range commands {
				if entry.ExpiresAt <= ts {
					delete(commands, commandHash)
				}
			}
			if len(commands) == 0 {
				delete(bucket, sessionID)
			}
		}
	}
	cleanupBucket(state.Pending)
	cleanupBucket(state.Approved)
}

func approvalStateDir() string {
	if configured := os.Getenv("BASH_GUARD_CODEX_STATE_DIR"); configured != "" {
		return expandHome(configured)
	}
	return filepath.Join(homeDir(), ".codex", "hook-state", "bash-guard")
}

func approvalStatePath() string {
	return filepath.Join(approvalStateDir(), "state.json")
}

func approvalLockPath() string {
	return filepath.Join(approvalStateDir(), "state.lock")
}

func expandHome(path string) string {
	if path == "~" {
		return homeDir()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir(), strings.TrimPrefix(path, "~/"))
	}
	return path
}
