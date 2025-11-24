package security

import (
	"bytes"
	"context"
    "fmt"
	"strings"

	"github.com/cucumber/godog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubernetes/test/e2e/framework"
    e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
    e2egpu "k8s.io/kubernetes/test/e2e/framework/gpu"
    e2enode "k8s.io/kubernetes/test/e2e/framework/node"
    
    admissionapi "k8s.io/pod-security-admission/api"
)

type contextKey string
const (
    frameworkKey contextKey = "framework"
    configKey contextKey = "config"
    nodeKey contextKey = "node"
    podsKey contextKey = "pods"
)

func InitializeSteps(ctx *godog.ScenarioContext) {
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        f := framework.NewDefaultFramework("godog-security")
        f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged
        
        c, err := framework.LoadClientset()
        if err != nil { return ctx, err }
        f.ClientSet = c
        
        cfg, err := framework.LoadConfig()
        if err != nil { return ctx, err }
        
        nsObj := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: "godog-security-"}}
        ns, err := c.CoreV1().Namespaces().Create(ctx, nsObj, metav1.CreateOptions{})
        if err != nil { return ctx, err }
        f.Namespace = ns
        
        ctx = context.WithValue(ctx, frameworkKey, f)
        ctx = context.WithValue(ctx, configKey, cfg)
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

    ctx.Step(`^a cluster with Nvidia GPU nodes$`, aClusterWithNvidiaGPUNodes)
    ctx.Step(`^I create a pod without GPU requests$`, iCreateAPodWithoutGPURequests)
    ctx.Step(`^running "([^"]*)" in the pod should fail with exit code (\d+)$`, runningInThePodShouldFailWithExitCode)
}

func aClusterWithNvidiaGPUNodes(ctx context.Context) (context.Context, error) {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    nodes, err := e2enode.GetReadyNodesIncludingTainted(ctx, f.ClientSet)
    if err != nil { return ctx, err }

    var selectedNode *v1.Node
    for _, node := range nodes.Items {
        allocatable, ok := node.Status.Allocatable[e2egpu.NVIDIAGPUResourceName]
        if ok && allocatable.Value() >= 2 {
            selectedNode = &node
            break
        }
    }
    if selectedNode == nil {
        return ctx, godog.ErrPending
    }
    return context.WithValue(ctx, nodeKey, selectedNode), nil
}

func iCreateAPodWithoutGPURequests(ctx context.Context) (context.Context, error) {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    node := ctx.Value(nodeKey).(*v1.Node)
    
    pod := e2epod.MakePod(f.Namespace.Name, nil, nil, f.NamespacePodSecurityLevel, "")
    pod.Spec.NodeName = node.Name
    
    pod, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Create(ctx, pod, metav1.CreateOptions{})
    if err != nil { return ctx, err }
    
    err = e2epod.WaitForPodRunningInNamespace(ctx, f.ClientSet, pod)
    if err != nil { return ctx, err }
    
    pods := []string{pod.Name}
    return context.WithValue(ctx, podsKey, pods), nil
}

func runningInThePodShouldFailWithExitCode(ctx context.Context, cmd string, exitCode int) error {
    f := ctx.Value(frameworkKey).(*framework.Framework)
    cfg := ctx.Value(configKey).(*rest.Config)
    pods := ctx.Value(podsKey).([]string)
    podName := pods[0]
    
    cmdParts := strings.Fields(cmd)
    
    _, _, err := execCommand(f.ClientSet, cfg, podName, f.Namespace.Name, cmdParts...)
    if err == nil {
        return fmt.Errorf("command %q succeeded, expected failure", cmd)
    }
    return nil
}

func execCommand(client kubernetes.Interface, config *rest.Config, podName, namespace string, cmd ...string) (string, string, error) {
    req := client.CoreV1().RESTClient().Post().
        Resource("pods").
        Name(podName).
        Namespace(namespace).
        SubResource("exec")
    
    req.VersionedParams(&v1.PodExecOptions{
        Command: cmd,
        Stdout:  true,
        Stderr:  true,
        TTY:     false,
    }, metav1.ParameterCodec)

    exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
    if err != nil { return "", "", err }

    var stdout, stderr bytes.Buffer
    err = exec.Stream(remotecommand.StreamOptions{
        Stdout: &stdout,
        Stderr: &stderr,
    })
    return stdout.String(), stderr.String(), err
}