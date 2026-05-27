package command

import (
	"reflect"
	"strings"
	"testing"
)

func TestCreateDefaultsToHA(t *testing.T) {
	cmd := Create()
	flag := cmd.PersistentFlags().Lookup("ha")
	if flag == nil {
		t.Fatal("expected --ha flag")
	}
	if flag.DefValue != "true" {
		t.Fatalf("expected --ha default true, got %q", flag.DefValue)
	}
}

func TestLimaCreateStepsHAUseRomaioiNames(t *testing.T) {
	steps := limaCreateSteps(true, "/repo/lykourgos/lima")

	wantContains := []string{
		`--name=byzantion`,
		`k0s-ha.yaml`,
		`--name=trapezous`,
		`--set=.additionalDisks[0].name=\"pyxis-trapezous\"`,
		`--name=nikaia`,
		`--set=.additionalDisks[0].name=\"pyxis-nikaia\"`,
	}

	for _, want := range wantContains {
		if !stepsContainScript(steps, want) {
			t.Fatalf("expected generated Lima start script to contain %q in %#v", want, collectStepArgs(steps))
		}
	}
}

func TestAnsibleStepsUseSplitRomaioiInventories(t *testing.T) {
	ha := ansibleSteps(true, "/repo/lykourgos/ansible")
	if len(ha) != 1 {
		t.Fatalf("expected one HA ansible step, got %d", len(ha))
	}
	wantHA := []string{"-i", "inventories/romaioi/ha/hosts.yaml", "playbooks/k0s-lima-ha.yaml"}
	if !reflect.DeepEqual(ha[0].args, wantHA) {
		t.Fatalf("unexpected HA ansible args: got %v want %v", ha[0].args, wantHA)
	}

	single := ansibleSteps(false, "/repo/lykourgos/ansible")
	if len(single) != 1 {
		t.Fatalf("expected one single ansible step, got %d", len(single))
	}
	wantSingle := []string{"-i", "inventories/romaioi/single/hosts.yaml", "playbooks/k0s-lima-single.yaml"}
	if !reflect.DeepEqual(single[0].args, wantSingle) {
		t.Fatalf("unexpected single ansible args: got %v want %v", single[0].args, wantSingle)
	}
}

func TestLimaDeleteStepsHAUseRomaioiNames(t *testing.T) {
	steps := limaDeleteSteps(true, "/repo/lykourgos/lima")
	got := collectStepArgs(steps)

	wantContains := [][]string{
		{"stop", "byzantion"},
		{"stop", "trapezous"},
		{"stop", "nikaia"},
		{"delete", "byzantion"},
		{"delete", "trapezous"},
		{"delete", "nikaia"},
		{"-c", "limactl disk delete pyxis-byzantion pyxis-trapezous pyxis-nikaia"},
	}

	for _, want := range wantContains {
		if !containsArgs(got, want) {
			t.Fatalf("expected args %v in %#v", want, got)
		}
	}
}

func collectStepArgs(steps []commandStep) [][]string {
	args := make([][]string, 0, len(steps))
	for _, step := range steps {
		args = append(args, step.args)
	}

	return args
}

func containsArgs(all [][]string, want []string) bool {
	for _, got := range all {
		if reflect.DeepEqual(got, want) {
			return true
		}
	}

	return false
}

func stepsContainScript(steps []commandStep, want string) bool {
	for _, step := range steps {
		if len(step.args) == 2 && step.args[0] == "-c" && strings.Contains(step.args[1], want) {
			return true
		}
	}

	return false
}
