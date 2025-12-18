package cluster

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestFluxSystemNamespace(t *testing.T) {
	f := features.New("flux-system namespace health").
		WithLabel("suite", "flux").
		Assess("namespace flux-system exists", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var ns corev1.Namespace
			if err := cfg.Client().Resources().Get(ctx, fluxNs, fluxNs, &ns); err != nil {
				t.Fatalf("namespace flux-system does not exist: %v", err)
			}
			return ctx
		}).
		Assess("all pods in flux-system are Ready", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			assertAllPodsReady(ctx, t, cfg, fluxNs, 2*time.Second)
			return ctx
		}).
		Assess("deleted pod gets recreated", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			podName := "flux-test-pod"

			var pod corev1.Pod
			if err := cfg.Client().Resources().Get(ctx, podName, lydiaNs, &pod); err != nil {
				t.Fatalf("test pod %s/%s not found: %v", lydiaNs, podName, err)
			}

			if err := cfg.Client().Resources().Delete(ctx, &pod); err != nil {
				t.Fatalf("failed to delete test pod: %v", err)
			}

			// Wait for Flux to notice and recreate it
			// We poll every 2s for up to 3 minutes
			err := wait.For(conditions.New(cfg.Client().Resources()).PodReady(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: lydiaNs},
			}), wait.WithTimeout(3*time.Minute), wait.WithInterval(2*time.Second))

			if err != nil {
				t.Fatalf("pod was not recreated or did not become ready: %v", err)
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}
