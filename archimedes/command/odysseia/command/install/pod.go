package install

import (
	"fmt"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/command"
	"strings"
	"time"
)

func (a *AppInstaller) checkStatusOfPod(podName string, timeToCheckInSeconds time.Duration) error {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(timeToCheckInSeconds)
	var ready bool
	for {
		select {
		case <-ticker.C:
			pods, err := a.Kube.Workload().List(command.DefaultNamespace)
			if err != nil {
				continue
			}

			podsReady := true
			for _, pod := range pods.Items {
				if !podsReady {
					break
				}
				if strings.Contains(pod.Name, podName) {
					if pod.Status.Phase != "Running" {
						glg.Infof("pod: %s not ready", pod.Name)
						podsReady = false
					}

					readyStatusFound := false
					for _, condition := range pod.Status.Conditions {
						if condition.Type == "Ready" && condition.Status == "True" {
							readyStatusFound = true
							break
						}
					}

					if !readyStatusFound {
						glg.Infof("pod: %s not ready", pod.Name)
						podsReady = false
					}
				}
			}

			if podsReady {
				ready = true
				ticker.Stop()
			} else {
				continue
			}

		case <-timeout:
			glg.Error("timed out")
			ticker.Stop()
		}
		break
	}

	if !ready {
		return fmt.Errorf("%s pods have not become healthy after %v seconds", podName, timeToCheckInSeconds)
	}
	return nil
}

func (a *AppInstaller) podIsRunning(podName string, timeToCheckInSeconds time.Duration) error {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(timeToCheckInSeconds)

	for {
		select {
		case <-ticker.C:
			pod, err := a.Kube.Workload().GetPodByName(command.DefaultNamespace, podName)
			if err != nil {
				continue
			}

			if pod.Status.Phase == "Running" {
				glg.Infof("pod: %s running: %s", pod.Name, pod.Status.Phase)
				ticker.Stop()
			} else {
				glg.Infof("pod: %s not running: %s", pod.Name, pod.Status.Phase)
				continue
			}
		case <-timeout:
			glg.Error("timed out")
			ticker.Stop()
			return fmt.Errorf("%s pod not running after %v seconds", podName, timeToCheckInSeconds)
		}
		break
	}
	return nil
}

func (a *AppInstaller) checkStatusOfNamedPod(podName string, timeToCheckInSeconds time.Duration) error {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(timeToCheckInSeconds)
	var ready bool

	for {
		select {
		case <-ticker.C:
			pod, err := a.Kube.Workload().GetPodByName(command.DefaultNamespace, podName)
			if err != nil {
				continue
			}

			if pod.Status.Phase != "Running" {
				glg.Infof("pod: %s not running: %s", pod.Name, pod.Status.Phase)
				continue
			}

			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					glg.Infof("pod: %s ready: %s", pod.Name, condition.Type)
					ready = true
				}
			}

			if ready {
				glg.Infof("pod: %s ready", pod.Name)
				ticker.Stop()
			} else {
				glg.Infof("pod: %s not ready", pod.Name)
				continue
			}

		case <-timeout:
			glg.Error("timed out")
			ticker.Stop()
			return fmt.Errorf("%s pods have not become healthy after %v seconds", podName, timeToCheckInSeconds)
		}
		break
	}

	if !ready {
		return fmt.Errorf("%s pod has not become healthy after %v seconds", podName, timeToCheckInSeconds)
	}

	return nil
}

func (a *AppInstaller) checkIfPodIsRunning(podName string, timeToCheckInSeconds time.Duration) error {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(timeToCheckInSeconds)
	var ready bool
	for {
		select {
		case <-ticker.C:
			pods, err := a.Kube.Workload().List(command.DefaultNamespace)
			if err != nil {
				continue
			}

			podsRunning := true
			for _, pod := range pods.Items {
				if !podsRunning {
					break
				}
				if strings.Contains(pod.Name, podName) {
					if pod.Status.Phase != "Running" {
						glg.Infof("pod: %s not ready", pod.Name)
						podsRunning = false
					}
				}
			}

			if podsRunning {
				ready = true
				ticker.Stop()
			} else {
				continue
			}

		case <-timeout:
			glg.Error("timed out")
			ticker.Stop()
		}
		break
	}

	if !ready {
		return fmt.Errorf("%s pods have not become healthy after %v seconds", podName, timeToCheckInSeconds)
	}
	return nil
}
