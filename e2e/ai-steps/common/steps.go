package common

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
	"k8s.io/kubernetes/test/e2e/framework"
)

var TestingMain *testing.M

func InitializeSteps(ctx *godog.ScenarioContext, m *testing.M) {
	TestingMain = m
    ctx.Step(`^a Kubernetes cluster$`, ensureCluster)
}

func ensureCluster(ctx context.Context) error {
    // Load clientset to ensure we are connected
    _, err := framework.LoadClientset()
    return err
}
