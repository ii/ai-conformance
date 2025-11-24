Feature: Dynamic Resource Allocation Support
  As an AI platform provider
  I want to ensure DRA is supported
  In order to manage hardware accelerators efficiently

  @AIConformance
  Scenario: DRA API availability
    Given a Kubernetes cluster
    Then the "resource.k8s.io/v1" API group should be available
