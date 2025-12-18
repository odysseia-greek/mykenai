package cluster

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
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

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}
