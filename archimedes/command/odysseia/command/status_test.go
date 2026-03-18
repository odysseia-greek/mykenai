package command

import "testing"

func TestCollectPodIssues(t *testing.T) {
	pods := []pod{
		{
			Metadata: podMetadata{
				Namespace: "flux-system",
				Name:      "source-controller-abc",
			},
			Status: podStatus{
				Phase: "Running",
				Conditions: []podCondition{
					{Type: "Ready", Status: "False", Reason: "ContainersNotReady"},
				},
				ContainerStatuses: []containerStatus{
					{
						Name:         "manager",
						Ready:        false,
						RestartCount: 3,
						State: containerState{
							Waiting: &waitingState{Reason: "CrashLoopBackOff"},
						},
						LastState: containerState{
							Terminated: &terminatedState{Reason: "Error", ExitCode: 1},
						},
					},
				},
			},
		},
		{
			Metadata: podMetadata{
				Namespace: "default",
				Name:      "healthy",
			},
			Status: podStatus{
				Phase: "Running",
				Conditions: []podCondition{
					{Type: "Ready", Status: "True"},
				},
				ContainerStatuses: []containerStatus{
					{Name: "app", Ready: true},
				},
			},
		},
	}

	issues := collectPodIssues(pods)
	if len(issues) != 1 {
		t.Fatalf("expected 1 problematic pod, got %d", len(issues))
	}

	got := issues[0]
	if got.Namespace != "flux-system" || got.Name != "source-controller-abc" {
		t.Fatalf("unexpected pod issue target: %+v", got)
	}
	if got.Restarts != 3 {
		t.Fatalf("expected restart count 3, got %d", got.Restarts)
	}
	if got.Ready != "0/1" {
		t.Fatalf("expected ready 0/1, got %s", got.Ready)
	}

	expectedIssues := map[string]bool{
		"container/manager waiting=CrashLoopBackOff": true,
		"container/manager last=Error(1)":            true,
		"ready=ContainersNotReady":                   true,
		"restarts=3":                                 true,
	}
	for _, issue := range got.Issues {
		delete(expectedIssues, issue)
	}
	if len(expectedIssues) != 0 {
		t.Fatalf("missing expected issues: %+v", expectedIssues)
	}
}
