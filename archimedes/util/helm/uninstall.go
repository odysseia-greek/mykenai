package helm

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

// Uninstall removes an chart
func (h *Helm) Uninstall(chart string) (*release.UninstallReleaseResponse, error) {
	client := action.NewUninstall(h.ActionConfig)
	rls, err := client.Run(chart)
	if err != nil {
		return nil, err
	}

	return rls, nil
}
