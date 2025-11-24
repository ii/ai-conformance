Feature: AI Networking Support
  As an AI platform provider
  I want to ensure networking capabilities are available
  In order to serve AI models

  @AIConformance
  Scenario: Gateway API CRDs availability
    Given a Kubernetes cluster
    Then the following Gateway API CRDs should be available:
      | gatewayclasses.gateway.networking.k8s.io  |
      | gateways.gateway.networking.k8s.io        |
      | httproutes.gateway.networking.k8s.io      |
      | grpcroutes.gateway.networking.k8s.io      |
      | referencegrants.gateway.networking.k8s.io |
