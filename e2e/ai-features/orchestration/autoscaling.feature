Feature: AI Workload Autoscaling
  As an AI platform operator
  I want resources to scale dynamically
  In order to handle fluctuating AI workloads efficiently

  @AIConformance
  Scenario: Cluster Autoscaling for GPU workloads
    Given a cluster with autoscaling enabled
    When I create pods requesting GPUs beyond capacity
    Then a new node should be provisioned
    And the pending pods should be scheduled
    When I delete the pods
    Then the node should be reclaimed

  @AIConformance
  Scenario: Horizontal Pod Autoscaling with Custom Metrics
    Given a deployment serving custom metrics
    And an HPA targeting the deployment
    When the metric value exceeds the target
    Then the deployment should scale up
    When the metric value drops below target
    Then the deployment should scale down
