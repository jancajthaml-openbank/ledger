Feature: Metrics test

  Scenario: metrics have expected keys
    Given vault is empty
    And   tenant M1 is onboarded
    And   ledger is configured with
      | property            | value |
      | METRICS_REFRESHRATE |    1s |

    Then metrics file /tmp/reports/blackbox-tests/metrics/metrics.M1.json should have following keys:
      | key                             |
      | promisedTransactions            |
      | promisedTransfers               |
      | committedTransactions           |
      | committedTransfers              |
      | rollbackedTransactions          |
      | rollbackedTransfers             |
      | transactionFinalizerCronLatency |
    And metrics file /tmp/reports/blackbox-tests/metrics/metrics.M1.json has permissions -rw-r--r--

    And metrics file /tmp/reports/blackbox-tests/metrics/metrics.json should have following keys:
      | key                      |
      | createTransactionLatency |
      | memoryAllocated          |
    And metrics file /tmp/reports/blackbox-tests/metrics/metrics.json has permissions -rw-r--r--

  Scenario: metrics can remembers previous values after reboot
    Given vault is empty
    And   tenant M2 is onboarded
    And   ledger is configured with
      | property            | value |
      | METRICS_REFRESHRATE |    1s |

    Then metrics file /tmp/reports/blackbox-tests/metrics/metrics.M2.json reports:
      | key                             | value |
      | promisedTransactions            |     0 |
      | promisedTransfers               |     0 |
      | committedTransactions           |     0 |
      | committedTransfers              |     0 |
      | rollbackedTransactions          |     0 |
      | rollbackedTransfers             |     0 |
      | transactionFinalizerCronLatency |     0 |

    When  pasive account M2/A with currency EUR exist
    And   pasive account M2/B with currency EUR exist
    And   1 EUR is transferred from M2/A to M2/B

    Then metrics file /tmp/reports/blackbox-tests/metrics/metrics.M2.json reports:
      | key                             | value |
      | promisedTransactions            |     1 |
      | promisedTransfers               |     1 |
      | committedTransactions           |     1 |
      | committedTransfers              |     1 |
      | rollbackedTransactions          |     0 |
      | rollbackedTransfers             |     0 |
      | transactionFinalizerCronLatency |     0 |

    When restart unit "ledger-unit@M2.service"
    Then metrics file /tmp/reports/blackbox-tests/metrics/metrics.M2.json reports:
      | key                             | value |
      | promisedTransactions            |     1 |
      | promisedTransfers               |     1 |
      | committedTransactions           |     1 |
      | committedTransfers              |     1 |
      | rollbackedTransactions          |     0 |
      | rollbackedTransfers             |     0 |
      | transactionFinalizerCronLatency |     0 |
