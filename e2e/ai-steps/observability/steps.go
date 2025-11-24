package observability

import (
	"context"

	"github.com/cucumber/godog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/kubernetes/test/e2e/framework"
    
    monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
    // prometheusutil "github.com/carlory/ai-conformance/e2e/util/prometheus"
)

type contextKey string
const (
    frameworkKey contextKey = "framework"
    promClientKey contextKey = "promClient"
    promInstanceKey contextKey = "promInstance"
)

func InitializeSteps(ctx *godog.ScenarioContext) {
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        f := framework.NewDefaultFramework("godog-observability")
        c, err := framework.LoadClientset()
        if err != nil { return ctx, err }
        f.ClientSet = c
        cfg, err := framework.LoadConfig()
        if err != nil { return ctx, err }
        
        nsObj := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: "godog-obs-"}}
        ns, err := c.CoreV1().Namespaces().Create(ctx, nsObj, metav1.CreateOptions{})
        if err != nil { return ctx, err }
        f.Namespace = ns
        
        promClient, err := monitoring.NewForConfig(cfg)
        if err != nil { return ctx, err }

        ctx = context.WithValue(ctx, frameworkKey, f)
        ctx = context.WithValue(ctx, promClientKey, promClient)
        return ctx, nil
    })
    
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        val := ctx.Value(frameworkKey)
        if val == nil { return ctx, nil }
        f := val.(*framework.Framework)
        if f != nil && f.ClientSet != nil && f.Namespace != nil {
            f.ClientSet.CoreV1().Namespaces().Delete(ctx, f.Namespace.Name, metav1.DeleteOptions{})
        }
        return ctx, nil
    })

    ctx.Step(`^GPU metrics should be collected in Prometheus$`, gpuMetricsShouldBeCollectedInPrometheus)
    ctx.Step(`^a Prometheus instance$`, aPrometheusInstance)
    ctx.Step(`^I deploy a resource consumer with custom metrics$`, iDeployAResourceConsumerWithCustomMetrics)
    ctx.Step(`^I create a ServiceMonitor$`, iCreateAServiceMonitor)
    ctx.Step(`^the custom metrics should be collected$`, theCustomMetricsShouldBeCollected)
}

func gpuMetricsShouldBeCollectedInPrometheus(ctx context.Context) error {
    return nil
}

func aPrometheusInstance(ctx context.Context) (context.Context, error) {
    promClient := ctx.Value(promClientKey).(monitoring.Interface)
    promList, err := promClient.MonitoringV1().Prometheuses(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
    if err != nil { return ctx, err }
    if len(promList.Items) == 0 {
        return ctx, godog.ErrPending
    }
    // Need to store *monitoringv1.Prometheus. 
    // Items[0] is the struct (value? or pointer?) 
    // List returns []*Prometheus usually? No, PrometheusList.Items is []Prometheus.
    return context.WithValue(ctx, promInstanceKey, &promList.Items[0]), nil
}

func iDeployAResourceConsumerWithCustomMetrics(ctx context.Context) error {
    return nil
}

func iCreateAServiceMonitor(ctx context.Context) error {
    return nil
}

func theCustomMetricsShouldBeCollected(ctx context.Context) error {
    return nil
}
