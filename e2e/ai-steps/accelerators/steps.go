package accelerators

import (
    "context"
    "fmt"
    "github.com/cucumber/godog"
    "k8s.io/kubernetes/test/e2e/framework"
)

func InitializeSteps(ctx *godog.ScenarioContext) {
    ctx.Step(`^the "([^"]*)" API group should be available$`, theAPIGroupShouldBeAvailable)
}

func theAPIGroupShouldBeAvailable(ctx context.Context, groupVersion string) error {
    c, err := framework.LoadClientset()
    if err != nil { return err }
    
    resources, err := c.Discovery().ServerResourcesForGroupVersion(groupVersion)
    if err != nil { return err }
    if resources == nil || len(resources.APIResources) == 0 {
        return fmt.Errorf("API group %s empty or nil", groupVersion)
    }
    return nil
}
