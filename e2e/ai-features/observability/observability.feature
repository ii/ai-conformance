Feature: AI Observability
  As an AI platform operator
  I want visibility into AI workloads
  In order to monitor performance and health

  @AIConformance
  Scenario: Nvidia GPU Metrics collection
    Given a cluster with Nvidia GPU nodes
    Then GPU metrics should be collected in Prometheus

  @AIConformance
  Scenario: AI Service Metrics collection
    Given a Prometheus instance
    When I deploy a resource consumer with custom metrics
    And I create a ServiceMonitor
    Then the custom metrics should be collected
