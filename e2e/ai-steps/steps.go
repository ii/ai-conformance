package aisteps

import (
    "testing"
    "github.com/cucumber/godog"
    "github.com/carlory/ai-conformance/e2e/ai-steps/common"
    "github.com/carlory/ai-conformance/e2e/ai-steps/accelerators"
    "github.com/carlory/ai-conformance/e2e/ai-steps/networking"
)

func InitializeScenario(ctx *godog.ScenarioContext, m *testing.M) {
    common.InitializeSteps(ctx, m)
    accelerators.InitializeSteps(ctx)
    networking.InitializeSteps(ctx)
}
