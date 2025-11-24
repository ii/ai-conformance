Feature: AI Workload Orchestration with Kueue
  As an AI platform operator
  I want to ensure efficient job scheduling
  In order to maximize resource utilization

  @AIConformance
  Scenario: Gang Scheduling of Jobs
    Given a Kueue setup with GPU resources
    When I submit two jobs that compete for resources
    Then both jobs should eventually succeed
