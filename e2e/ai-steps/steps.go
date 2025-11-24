package aisteps

import (
    "testing"
    "github.com/cucumber/godog"
    "github.com/carlory/ai-conformance/e2e/ai-steps/common"
    "github.com/carlory/ai-conformance/e2e/ai-steps/accelerators"
)

func InitializeScenario(ctx *godog.ScenarioContext, m *testing.M) {
    common.InitializeSteps(ctx, m)
    accelerators.InitializeSteps(ctx)
}