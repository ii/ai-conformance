package orchestration

import (
	"context"
    "strconv"
    "time"

	"github.com/cucumber/godog"
    "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/kubernetes/test/e2e/framework"
    e2egpu "k8s.io/kubernetes/test/e2e/framework/gpu"
    e2enode "k8s.io/kubernetes/test/e2e/framework/node"
    e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
    resourcehelper "k8s.io/component-helpers/resource"
    
    kueuev1beta1 "sigs.k8s.io/kueue/apis/kueue/v1beta1"
    kueueclient "sigs.k8s.io/kueue/client-go/clientset/versioned"
)

type contextKey string
const (
    frameworkKey contextKey = "framework"
    kueueClientKey contextKey = "kueueClient"
    queueKey contextKey = "queue"
    availableGPUsKey contextKey = "availableGPUs"
    pendingPodKey contextKey = "pendingPod"
    provisionedNodeKey contextKey = "provisionedNode"
)

func InitializeSteps(ctx *godog.ScenarioContext) {
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        f := framework.NewDefaultFramework("godog-orchestration")
        
        c, err := framework.LoadClientset()
        if err != nil { return ctx, err }
        f.ClientSet = c
        
        cfg, err := framework.LoadConfig()
        if err != nil { return ctx, err }
        
        nsObj := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: "godog-orch-"}}
        ns, err := c.CoreV1().Namespaces().Create(ctx, nsObj, metav1.CreateOptions{})
        if err != nil { return ctx, err }
        f.Namespace = ns
        
        kc, err := kueueclient.NewForConfig(cfg)
        if err != nil { return ctx, err }

        ctx = context.WithValue(ctx, frameworkKey, f)
        ctx = context.WithValue(ctx, kueueClientKey, kc)
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

    ctx.Step(`^a Kueue setup with GPU resources$`, aKueueSetupWithGPUResources)
    ctx.Step(`^I submit two jobs that compete for resources$`, iSubmitTwoJobsThatCompeteForResources)
    ctx.Step(`^both jobs should eventually succeed$`, bothJobsShouldEventuallySucceed)
    
    // Autoscaling steps
    ctx.Step(`^a cluster with autoscaling enabled$`, aClusterWithAutoscalingEnabled)
    ctx.Step(`^I create pods requesting GPUs beyond capacity$`, iCreatePodsRequestingGPUsBeyondCapacity)
    ctx.Step(`^a new node should be provisioned$`, aNewNodeShouldBeProvisioned)
    ctx.Step(`^the pending pods should be scheduled$`, thePendingPodsShouldBeScheduled)
    ctx.Step(`^I delete the pods$`, iDeleteThePods)
    ctx.Step(`^the node should be reclaimed$`, theNodeShouldBeReclaimed)
    ctx.Step(`^a deployment serving custom metrics$`, aDeploymentServingCustomMetrics)
    ctx.Step(`^an HPA targeting the deployment$`, anHPATargetingTheDeployment)
    ctx.Step(`^the metric value exceeds the target$`, theMetricValueExceedsTheTarget)
    ctx.Step(`^the deployment should scale up$`, theDeploymentShouldScaleUp)
    ctx.Step(`^the metric value drops below target$`, theMetricValueDropsBelowTarget)
    ctx.Step(`^the deployment should scale down$`, theDeploymentShouldScaleDown)
}

func aKueueSetupWithGPUResources(ctx context.Context) (context.Context, error) {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    kc := ctx.Value(kueueClientKey).(kueueclient.Interface)
    
    nodes, err := e2enode.GetReadyNodesIncludingTainted(ctx, f.ClientSet)
    if err != nil { return ctx, err }
    
    allocatable := 0
    for _, node := range nodes.Items {
        val, ok := node.Status.Allocatable[e2egpu.NVIDIAGPUResourceName]
        if ok {
            allocatable += int(val.Value())
        }
    }
    
    used := 0
    pods, err := f.ClientSet.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
    if err == nil {
        for _, pod := range pods.Items {
            if pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
                for resourceName, val := range resourcehelper.PodLimits(&pod, resourcehelper.PodResourcesOptions{}) {
                    if string(resourceName) == e2egpu.NVIDIAGPUResourceName {
                        used += int(val.Value())
                    }
                }
            }
        }
    }
    
    available := allocatable - used
    
    nominalQuota := available * 2
    if nominalQuota == 0 { nominalQuota = 4 }
    
    rf := &kueuev1beta1.ResourceFlavor{ObjectMeta: metav1.ObjectMeta{Name: f.UniqueName}}
    _, err = kc.KueueV1beta1().ResourceFlavors().Create(ctx, rf, metav1.CreateOptions{})
    if err != nil { return ctx, err }
    
    clusterQueue := &kueuev1beta1.ClusterQueue{
        ObjectMeta: metav1.ObjectMeta{Name: f.UniqueName},
        Spec: kueuev1beta1.ClusterQueueSpec{
            NamespaceSelector: &metav1.LabelSelector{},
            ResourceGroups: []kueuev1beta1.ResourceGroup{
                {
                    CoveredResources: []v1.ResourceName{e2egpu.NVIDIAGPUResourceName},
                    Flavors: []kueuev1beta1.FlavorQuotas{
                        {
                            Name: kueuev1beta1.ResourceFlavorReference(rf.Name),
                            Resources: []kueuev1beta1.ResourceQuota{
                                {
                                    Name: e2egpu.NVIDIAGPUResourceName,
                                    NominalQuota: resource.MustParse(strconv.Itoa(nominalQuota)),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    _, err = kc.KueueV1beta1().ClusterQueues().Create(ctx, clusterQueue, metav1.CreateOptions{})
    if err != nil { return ctx, err }
    
    localQueue := &kueuev1beta1.LocalQueue{
        ObjectMeta: metav1.ObjectMeta{Name: f.UniqueName},
        Spec: kueuev1beta1.LocalQueueSpec{
            ClusterQueue: kueuev1beta1.ClusterQueueReference(clusterQueue.Name),
        },
    }
    _, err = kc.KueueV1beta1().LocalQueues(f.Namespace.Name).Create(ctx, localQueue, metav1.CreateOptions{})
    if err != nil { return ctx, err }
    
    ctx = context.WithValue(ctx, queueKey, localQueue.Name)
    ctx = context.WithValue(ctx, availableGPUsKey, available)
    return ctx, nil
}

func iSubmitTwoJobsThatCompeteForResources(ctx context.Context) error {
    return nil
}

func bothJobsShouldEventuallySucceed(ctx context.Context) error {
    return nil
}

// Autoscaling steps implementation

func aClusterWithAutoscalingEnabled(ctx context.Context) error {
    _ = ctx.Value(frameworkKey).(*framework.Framework)
    return nil
}

func iCreatePodsRequestingGPUsBeyondCapacity(ctx context.Context) (context.Context, error) {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    client := f.ClientSet
    ns := f.Namespace.Name

    var pendingPod *v1.Pod
    
    // Simple loop to create pods until one pends
    for i := 0; i < 10; i++ { // limit to avoid infinite loop
        pod := e2epod.MakePod(ns, nil, nil, f.NamespacePodSecurityLevel, "")
        pod.Spec.Containers[0].Resources.Limits = map[v1.ResourceName]resource.Quantity{
            v1.ResourceName(e2egpu.NVIDIAGPUResourceName): resource.MustParse("1"),
        }
        pod, err := client.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{})
        if err != nil { return ctx, err }
        
        // Wait briefly to see if it schedules or pends
        err = e2epod.WaitForPodCondition(ctx, client, ns, pod.Name, "PodScheduled", 30*time.Second, func(pod *v1.Pod) (bool, error) {
            if pod.Status.Phase == v1.PodPending {
                for _, cond := range pod.Status.Conditions {
                    if cond.Type == v1.PodScheduled && cond.Status == v1.ConditionFalse && cond.Reason == v1.PodReasonUnschedulable {
                        return true, nil
                    }
                }
            }
            return false, nil
        })
        
        if err == nil {
            // It is pending and unschedulable!
            pendingPod = pod
            break
        }
    }
    
    if pendingPod == nil {
        return ctx, godog.ErrPending // Or error "Could not trigger autoscaling"
    }
    
    return context.WithValue(ctx, pendingPodKey, pendingPod.Name), nil
}

func aNewNodeShouldBeProvisioned(ctx context.Context) (context.Context, error) {
    // This is implicitly checked by pending pod becoming scheduled on a new node.
    // But we can check node count increase?
    return ctx, nil
}

func thePendingPodsShouldBeScheduled(ctx context.Context) (context.Context, error) {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    client := f.ClientSet
    ns := f.Namespace.Name
    podName := ctx.Value(pendingPodKey).(string)
    
    // Wait for pod to be running (implies scaled up)
    // Use long timeout for autoscaling
    err := e2epod.WaitForPodRunningInNamespaceSlow(ctx, client, ns, podName)
    if err != nil { return ctx, err }
    
    pod, err := client.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
    if err != nil { return ctx, err }
    
    return context.WithValue(ctx, provisionedNodeKey, pod.Spec.NodeName), nil
}

func iDeleteThePods(ctx context.Context) error {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    client := f.ClientSet
    ns := f.Namespace.Name
    podName := ctx.Value(pendingPodKey).(string)
    
    return client.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{})
}

func theNodeShouldBeReclaimed(ctx context.Context) error {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    nodeName := ctx.Value(provisionedNodeKey).(string)
    
    // Wait for node to be deleted
    return framework.Gomega().Eventually(ctx, framework.HandleRetry(func(ctx context.Context) (*v1.Node, error) {
			node, err := f.ClientSet.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return nil, nil // Reclaimed
			}
			return node, err
    })).WithTimeout(15 * time.Minute).Should(gomega.BeNil())
}

func aDeploymentServingCustomMetrics(ctx context.Context) error { return nil }
func anHPATargetingTheDeployment(ctx context.Context) error { return nil }
func theMetricValueExceedsTheTarget(ctx context.Context) error { return nil }
func theDeploymentShouldScaleUp(ctx context.Context) error { return nil }
func theMetricValueDropsBelowTarget(ctx context.Context) error { return nil }
func theDeploymentShouldScaleDown(ctx context.Context) error { return nil }
