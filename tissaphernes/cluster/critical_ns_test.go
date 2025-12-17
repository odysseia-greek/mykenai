package cluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const kubeSystem = "kube-system"

func TestPodsReadyInKubeSystem(t *testing.T) {

	f := features.New("pods are Ready in kube-system").
		WithLabel("suite", "tissaphernes").
		Assess("all pods Ready", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Run(kubeSystem, func(t *testing.T) {
				assertAllPodsReady(ctx, t, cfg, kubeSystem, 2*time.Second)
			})
			return ctx
		}).Feature()

	testenv.Test(t, f)
}

func assertAllPodsReady(ctx context.Context, t *testing.T, cfg *envconf.Config, ns string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for {
		var pods corev1.PodList
		if err := cfg.Client().Resources(ns).List(ctx, &pods); err != nil {
			t.Fatalf("list pods in %s: %v", ns, err)
		}

		notReady := []string{}
		for i := range pods.Items {
			p := &pods.Items[i]

			// ignore pods that are completed successfully
			if p.Status.Phase == corev1.PodSucceeded {
				continue
			}

			if !isPodReady(p) {
				notReady = append(notReady, fmt.Sprintf("%s/%s (%s)", ns, p.Name, p.Status.Phase))
			}
		}

		if len(notReady) == 0 {
			return
		}

		if time.Now().After(deadline) {
			t.Fatalf("not all pods ready in %s after %s:\n- %s", ns, timeout, joinLines(notReady))
		}

		// small retry loop so transient rollouts don’t flap the test
		time.Sleep(3 * time.Second)
	}
}

func isPodReady(p *corev1.Pod) bool {
	// if it’s being deleted, treat as not ready
	if p.DeletionTimestamp != nil && !p.DeletionTimestamp.IsZero() {
		return false
	}
	// Pods with no conditions are not ready
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func joinLines(xs []string) string {
	out := ""
	for i, s := range xs {
		if i > 0 {
			out += "\n- "
		}
		out += s
	}
	return out
}

// Optional: keep `metav1` import used if you later add label selectors etc.
var _ = metav1.TypeMeta{}
