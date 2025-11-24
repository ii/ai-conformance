package networking

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/framework"
    
    e2ecrd "github.com/carlory/ai-conformance/e2e/util/framework/crd"
)

func InitializeSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the following Gateway API CRDs should be available:$`, theFollowingGatewayAPICRDsShouldBeAvailable)
}

func theFollowingGatewayAPICRDsShouldBeAvailable(ctx context.Context, table *godog.Table) error {
    cfg, err := framework.LoadConfig()
    if err != nil {
        return fmt.Errorf("error loading config: %w", err)
    }
    c, err := apiextclientset.NewForConfig(cfg)
    if err != nil {
        return fmt.Errorf("error creating apiextensions client: %w", err)
    }

    expectedCrds := sets.New[string]()
    for _, row := range table.Rows {
        // Assuming single column
        if len(row.Cells) > 0 {
             expectedCrds.Insert(row.Cells[0].Value)
        }
    }

    crds, err := c.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("error listing CRDs: %w", err)
    }

    foundCrds := sets.New[string]()
    for _, crd := range crds.Items {
        if !expectedCrds.Has(crd.Name) {
            continue
        }
        foundCrds.Insert(crd.Name)
        
        err = e2ecrd.WaitForCrdEstablishedAndNamesAccepted(ctx, c, crd.Name)
        if err != nil {
             return fmt.Errorf("CRD %s not established: %w", crd.Name, err)
        }
    }

    if !foundCrds.Equal(expectedCrds) {
        missing := expectedCrds.Difference(foundCrds)
        return fmt.Errorf("missing gateway crds: %v", sets.List(missing))
    }

    return nil
}
