package main

import "strings"

var exactIntents = map[string]string{
	"guard.parse_error_after_trigger": "run a command containing a destructive keyword that the hook could not parse",
	"rm.inline_parse_error":           "run inline shell code containing a destructive command that the hook could not parse",
	"rm.shell_pipe_sink":              "pipe generated or file content into a shell evaluator",
	"rm.remote":                       "delete files on a remote host",
	"rm.chrooted":                     "delete files inside a chroot with different path semantics",
	"rm.stdin_args":                   "delete files whose target list comes from stdin",
	"rm.catastrophic":                 "delete a protected filesystem path",
	"rm.outside_safe_path":            "delete files outside the project or safe scratch paths",
	"rm.delete_safe_root":             "delete the root of a shared safe scratch path",
	"rm.shared_root_glob":             "delete a top-level glob inside a shared scratch path",
	"rm.unresolvable":                 "delete a path containing variables or substitutions that cannot be verified",
	"git.force_push":                  "overwrite remote Git history",
	"git.push_delete":                 "delete a remote Git ref",
	"git.reset_hard":                  "discard local uncommitted changes and reset the repo to the target ref",
	"git.clean_force":                 "remove untracked files from the working tree",
	"git.checkout_pathspec":           "discard working-tree changes through git checkout",
	"git.restore_pathspec":            "discard working-tree changes through git restore",
	"git.branch_force_delete":         "force-delete a local branch and any unmerged commits on it",
	"git.stash_loss":                  "delete stashed work",
	"git.history_rewrite":             "rewrite Git history",
	"db_client.sql_destructive":       "run destructive SQL against a database",
	"db_client.redis_destructive":     "run a destructive Redis command",
}

var prefixIntents = []struct {
	prefix string
	intent string
}{
	{"supabase.", "apply or rewrite database migration state"},
	{"infra.kubectl_", "modify Kubernetes resources"},
	{"infra.gcloud_", "modify Google Cloud resources"},
	{"infra.gcs_", "modify Google Cloud Storage objects"},
	{"infra.helm_", "modify Helm releases"},
	{"infra.docker_", "modify Docker containers, images, or system state"},
	{"infra.mongo_", "run a destructive MongoDB operation"},
	{"infra.terraform_", "modify infrastructure state"},
	{"infra.gsutil_", "modify Google Cloud Storage objects"},
	{"infra.opensearch_", "send a mutating request to OpenSearch or Elasticsearch"},
	{"infra.cloud_api_", "send a mutating request to a cloud control-plane API"},
	{"infra.graphql_", "send a mutating GraphQL request"},
	{"infra.aws_", "modify AWS resources"},
	{"infra.azure_", "modify Azure resources"},
	{"infra.oci_", "modify Oracle Cloud resources"},
	{"infra.ibmcloud_", "modify IBM Cloud resources"},
	{"paas.", "modify or delete PaaS resources"},
}

func intentForReasonCode(reasonCode string) string {
	codes := strings.Split(reasonCode, ",")
	compact := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code != "" {
			compact = append(compact, code)
		}
	}
	if len(compact) > 1 {
		return "run multiple risky operations in one Bash command"
	}
	code := ""
	if len(compact) == 1 {
		code = compact[0]
	}
	if intent, ok := exactIntents[code]; ok {
		return intent
	}
	for _, item := range prefixIntents {
		if strings.HasPrefix(code, item.prefix) {
			return item.intent
		}
	}
	return "run a risky Bash command"
}
