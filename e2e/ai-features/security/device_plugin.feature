Feature: Secure Accelerator Access with Device Plugin
  As an AI platform operator
  I want to ensure GPU access is secure and isolated
  In order to prevent unauthorized usage and ensure fairness

  Background:
    Given a cluster with Nvidia GPU nodes

  @AIConformance
  Scenario: Pods without device requests cannot access devices
    When I create a pod without GPU requests
    Then running "nvidia-smi" in the pod should fail with exit code 127

  @AIConformance
  Scenario: Different pods get different devices
    When I create two pods each requesting 1 GPU
    Then the devices assigned to them should be different
