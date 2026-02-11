package cluster

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const eleusinianNs = "eleusinian"
const teleteSecret = "telete"

func TestEleusinianSecrets(t *testing.T) {
	f := features.New("eleusinian namespace secrets").
		WithLabel("suite", "secrets").
		Assess("namespace eleusinian exists", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var ns corev1.Namespace
			if err := cfg.Client().Resources().Get(ctx, eleusinianNs, eleusinianNs, &ns); err != nil {
				t.Fatalf("namespace %s does not exist: %v", eleusinianNs, err)
			}
			return ctx
		}).
		Assess("secret telete exists and has required keys", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var secret corev1.Secret
			if err := cfg.Client().Resources().Get(ctx, teleteSecret, eleusinianNs, &secret); err != nil {
				t.Fatalf("secret %s does not exist in namespace %s: %v", teleteSecret, eleusinianNs, err)
			}

			// Check hierophant key exists and has value longer than 1
			hierophantValue, hierophantExists := secret.Data["hierophant"]
			if !hierophantExists {
				t.Fatalf("secret %s does not contain key 'hierophant'", teleteSecret)
			}
			if len(hierophantValue) <= 1 {
				t.Fatalf("secret %s key 'hierophant' value is too short (length: %d)", teleteSecret, len(hierophantValue))
			}

			// Check mystery key exists and has value longer than 1
			mysteryValue, mysteryExists := secret.Data["mystery"]
			if !mysteryExists {
				t.Fatalf("secret %s does not contain key 'mystery'", teleteSecret)
			}
			if len(mysteryValue) <= 1 {
				t.Fatalf("secret %s key 'mystery' value is too short (length: %d)", teleteSecret, len(mysteryValue))
			}

			t.Logf("secret %s verified: hierophant (length: %d), mystery (length: %d)",
				teleteSecret, len(hierophantValue), len(mysteryValue))

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}
