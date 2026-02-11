package cluster

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestCiliumPodsHealthy(t *testing.T) {
	f := features.New("cilium pods are healthy").
		WithLabel("suite", "cilium").
		Assess("cilium is Ready", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Run(ciliumNs, func(t *testing.T) {
				assertAllPodsReady(ctx, t, cfg, ciliumNs, 2*time.Second)
			})
			return ctx
		}).Feature()

	testenv.Test(t, f)
}
