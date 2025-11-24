package aisteps

import (
    "testing"
    "github.com/cucumber/godog"
    "github.com/carlory/ai-conformance/e2e/ai-steps/common"
)

func InitializeScenario(ctx *godog.ScenarioContext, m *testing.M) {
    common.InitializeSteps(ctx, m)
}
