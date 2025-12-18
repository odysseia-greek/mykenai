package cluster

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		Assess("background drift correction works (1m interval)", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			podName := "flux-test-pod"

			// Delete and wait for Flux to notice on its own schedule
			if err := deletePodIfExists(ctx, cfg, podName, lydiaNs); err != nil {
				t.Fatalf("failed to delete pod for background test: %v", err)
			}

			// Since interval is 1m, we wait up to 2.5m to be safe
			err := wait.For(conditions.New(cfg.Client().Resources()).PodReady(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: lydiaNs},
			}), wait.WithTimeout(150*time.Second), wait.WithInterval(5*time.Second))

			if err != nil {
				t.Fatalf("background drift correction failed: %v", err)
			}
			return ctx
		}).
		Assess("manual trigger forces immediate recreation", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			podName := "flux-test-pod"

			if err := deletePodIfExists(ctx, cfg, podName, lydiaNs); err != nil {
				t.Fatalf("failed to delete pod for manual test: %v", err)
			}

			// Force Flux to reconcile immediately via annotation
			if err := triggerFluxReconcile(ctx, cfg, lydiaNs, fluxNs); err != nil {
				t.Logf("warning: could not trigger manual reconcile: %v", err)
			}

			// This should be much faster (15-30s)
			err := wait.For(conditions.New(cfg.Client().Resources()).PodReady(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: lydiaNs},
			}), wait.WithTimeout(45*time.Second), wait.WithInterval(2*time.Second))

			if err != nil {
				t.Fatalf("immediate recreation failed: %v", err)
			}
			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func deletePodIfExists(ctx context.Context, cfg *envconf.Config, name, ns string) error {
	var pod corev1.Pod
	err := cfg.Client().Resources().Get(ctx, name, ns, &pod)
	if err != nil {
		return nil // already gone or doesn't exist
	}
	return cfg.Client().Resources().Delete(ctx, &pod)
}

func triggerFluxReconcile(ctx context.Context, cfg *envconf.Config, name, ns string) error {
	ks := &unstructured.Unstructured{}
	ks.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kustomize.toolkit.fluxcd.io",
		Version: "v1",
		Kind:    "Kustomization",
	})
	if err := cfg.Client().Resources().Get(ctx, name, ns, ks); err != nil {
		return err
	}
	ksAnnotation := ks.GetAnnotations()
	if ksAnnotation == nil {
		ksAnnotation = make(map[string]string)
	}
	ksAnnotation["reconcile.fluxcd.io/requestedAt"] = time.Now().Format(time.RFC3339)
	ks.SetAnnotations(ksAnnotation)
	return cfg.Client().Resources().Update(ctx, ks)
}
