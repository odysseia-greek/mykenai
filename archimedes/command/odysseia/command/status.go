package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const defaultStatusWarningsLimit = 15

type statusOptions struct {
	namespace     string
	warningsLimit int
}

type podList struct {
	Items []pod `json:"items"`
}

type pod struct {
	Metadata podMetadata `json:"metadata"`
	Status   podStatus   `json:"status"`
}

type podMetadata struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type podStatus struct {
	Phase                 string            `json:"phase"`
	Reason                string            `json:"reason"`
	Message               string            `json:"message"`
	Conditions            []podCondition    `json:"conditions"`
	InitContainerStatuses []containerStatus `json:"initContainerStatuses"`
	ContainerStatuses     []containerStatus `json:"containerStatuses"`
}

type podCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type containerStatus struct {
	Name         string         `json:"name"`
	Ready        bool           `json:"ready"`
	RestartCount int            `json:"restartCount"`
	State        containerState `json:"state"`
	LastState    containerState `json:"lastState"`
}

type containerState struct {
	Waiting    *waitingState    `json:"waiting"`
	Terminated *terminatedState `json:"terminated"`
}

type waitingState struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type terminatedState struct {
	Reason   string `json:"reason"`
	Message  string `json:"message"`
	ExitCode int    `json:"exitCode"`
}

type podIssue struct {
	Namespace string
	Name      string
	Phase     string
	Ready     string
	Restarts  int
	Issues    []string
}

func Status() *cobra.Command {
	opts := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "show odysseia cluster status",
		Long: `Show a concise odysseia cluster status summary with infrastructure health,
Flux reconciliation status, pod failures, restart counts, and recent warning events.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.namespace, "namespace", "n", "", "namespace to inspect; defaults to all namespaces")
	cmd.PersistentFlags().IntVar(&opts.warningsLimit, "warnings", defaultStatusWarningsLimit, "number of recent warning events to show")

	return cmd
}

func runStatus(cmd *cobra.Command, opts *statusOptions) error {
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "ODYSSEIA STATUS")
	fmt.Fprintln(out, "===============")

	renderSection(out, "Context", statusContextSummary())
	renderSection(out, "Nodes", statusCommandSummary("kubectl", "get", "nodes"))

	if _, err := exec.LookPath("cilium"); err == nil {
		renderSection(out, "Cilium", statusCommandSummary("cilium", "status"))
	} else {
		renderSection(out, "Cilium", "Skipped: `cilium` CLI not found")
	}

	fluxArgs := []string{"get", "all"}
	if opts.namespace == "" {
		fluxArgs = append(fluxArgs, "-A")
	} else {
		fluxArgs = append(fluxArgs, "-n", opts.namespace)
	}
	if _, err := exec.LookPath("flux"); err == nil {
		renderSection(out, "Flux", statusCommandSummary("flux", fluxArgs...))
	} else {
		renderSection(out, "Flux", "Skipped: `flux` CLI not found")
	}

	podSummary, err := podStatusSummary(opts.namespace)
	if err != nil {
		return err
	}
	renderSection(out, "Problem Pods", podSummary)

	warningSummary := warningEventsSummary(opts.namespace, opts.warningsLimit)
	if warningSummary != "" {
		renderSection(out, "Recent Warnings", warningSummary)
	}

	return nil
}

func renderSection(out io.Writer, title, body string) {
	fmt.Fprintf(out, "\n%s\n%s\n%s\n", title, strings.Repeat("-", len(title)), strings.TrimSpace(body))
}

func statusContextSummary() string {
	context, err := runStatusCommand("kubectl", "config", "current-context")
	if err != nil {
		return fmt.Sprintf("Failed to read current context: %v", err)
	}

	return strings.TrimSpace(context)
}

func statusCommandSummary(name string, args ...string) string {
	output, err := runStatusCommand(name, args...)
	if err != nil {
		return fmt.Sprintf("Command failed: %v\n%s", err, strings.TrimSpace(output))
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return "No output"
	}

	return trimmed
}

func podStatusSummary(namespace string) (string, error) {
	args := []string{"get", "pods"}
	if namespace == "" {
		args = append(args, "-A")
	} else {
		args = append(args, "-n", namespace)
	}
	args = append(args, "-o", "json")

	output, err := runStatusCommand("kubectl", args...)
	if err != nil {
		return "", fmt.Errorf("get pods: %w", err)
	}

	var pods podList
	if err := json.Unmarshal([]byte(output), &pods); err != nil {
		return "", fmt.Errorf("decode pod status: %w", err)
	}

	issues := collectPodIssues(pods.Items)
	if len(issues) == 0 {
		return "No problematic pods found", nil
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAMESPACE\tNAME\tPHASE\tREADY\tRESTARTS\tISSUES")
	for _, issue := range issues {
		fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\t%d\t%s\n",
			issue.Namespace,
			issue.Name,
			issue.Phase,
			issue.Ready,
			issue.Restarts,
			strings.Join(issue.Issues, "; "),
		)
	}
	_ = tw.Flush()

	return strings.TrimSpace(buf.String()), nil
}

func collectPodIssues(pods []pod) []podIssue {
	issues := make([]podIssue, 0)

	for _, item := range pods {
		totalContainers := len(item.Status.InitContainerStatuses) + len(item.Status.ContainerStatuses)
		readyContainers := countReadyContainers(item.Status.InitContainerStatuses) + countReadyContainers(item.Status.ContainerStatuses)
		restarts := countRestarts(item.Status.InitContainerStatuses) + countRestarts(item.Status.ContainerStatuses)

		var details []string
		if item.Status.Phase != "" && item.Status.Phase != "Running" && item.Status.Phase != "Succeeded" {
			details = append(details, "phase="+item.Status.Phase)
		}
		if item.Status.Reason != "" {
			details = append(details, "pod="+item.Status.Reason)
		}

		details = append(details, collectContainerIssues("init", item.Status.InitContainerStatuses)...)
		details = append(details, collectContainerIssues("container", item.Status.ContainerStatuses)...)

		if readyProblem := readyConditionIssue(item.Status.Conditions); readyProblem != "" {
			details = append(details, readyProblem)
		}
		if restarts > 0 {
			details = append(details, fmt.Sprintf("restarts=%d", restarts))
		}

		if len(details) == 0 {
			continue
		}

		issues = append(issues, podIssue{
			Namespace: item.Metadata.Namespace,
			Name:      item.Metadata.Name,
			Phase:     defaultString(item.Status.Phase, "Unknown"),
			Ready:     fmt.Sprintf("%d/%d", readyContainers, totalContainers),
			Restarts:  restarts,
			Issues:    dedupeStrings(details),
		})
	}

	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Restarts != issues[j].Restarts {
			return issues[i].Restarts > issues[j].Restarts
		}
		if issues[i].Namespace != issues[j].Namespace {
			return issues[i].Namespace < issues[j].Namespace
		}
		return issues[i].Name < issues[j].Name
	})

	return issues
}

func collectContainerIssues(prefix string, statuses []containerStatus) []string {
	var issues []string

	for _, status := range statuses {
		if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
			issues = append(issues, fmt.Sprintf("%s/%s waiting=%s", prefix, status.Name, status.State.Waiting.Reason))
		}
		if status.State.Terminated != nil && (status.State.Terminated.ExitCode != 0 || status.State.Terminated.Reason != "") {
			reason := defaultString(status.State.Terminated.Reason, "terminated")
			issues = append(issues, fmt.Sprintf("%s/%s terminated=%s(%d)", prefix, status.Name, reason, status.State.Terminated.ExitCode))
		}
		if status.LastState.Terminated != nil && status.LastState.Terminated.ExitCode != 0 {
			reason := defaultString(status.LastState.Terminated.Reason, "terminated")
			issues = append(issues, fmt.Sprintf("%s/%s last=%s(%d)", prefix, status.Name, reason, status.LastState.Terminated.ExitCode))
		}
	}

	return issues
}

func readyConditionIssue(conditions []podCondition) string {
	for _, condition := range conditions {
		if condition.Type == "Ready" && condition.Status != "True" {
			reason := defaultString(condition.Reason, "NotReady")
			return "ready=" + reason
		}
	}

	return ""
}

func countReadyContainers(statuses []containerStatus) int {
	ready := 0
	for _, status := range statuses {
		if status.Ready {
			ready++
		}
	}

	return ready
}

func countRestarts(statuses []containerStatus) int {
	restarts := 0
	for _, status := range statuses {
		restarts += status.RestartCount
	}

	return restarts
}

func warningEventsSummary(namespace string, limit int) string {
	args := []string{"get", "events"}
	if namespace == "" {
		args = append(args, "-A")
	} else {
		args = append(args, "-n", namespace)
	}
	args = append(args, "--field-selector", "type=Warning", "--sort-by=.lastTimestamp")

	output, err := runStatusCommand("kubectl", args...)
	if err != nil {
		return fmt.Sprintf("Failed to read warning events: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= 1 {
		return "No warning events found"
	}

	if limit > 0 && len(lines) > limit+1 {
		lines = append(lines[:1], lines[len(lines)-limit:]...)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func runStatusCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	return out
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
