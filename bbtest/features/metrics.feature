@metrics
Feature: Metrics test

  Scenario: metrics have expected keys
    Given vault is empty
    And   tenant M2 is onbdoarded
    And   ledger is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /reports/metrics.M2.json should have following keys:
    """
      promisedTransactions
      promisedTransfers
      committedTransactions
      committedTransfers
      rollbackedTransactions
      rollbackedTransfers
      forwardedTransactions
      forwardedTransfers
    """
    And metrics file /reports/metrics.M2.json has permissions -rw-r--r--
    And metrics file /reports/metrics.json should have following keys:
    """
      createTransactionLatency
      forwardTransferLatency
      getTransactionLatency
      getTransactionsLatency
    """
    And metrics file /reports/metrics.json has permissions -rw-r--r--

  Scenario: metrics report expected results
    Given vault is empty
    And   tenant M1 is onbdoarded
    And   ledger is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    When  pasive account M1/A with currency EUR exist
    And   pasive account M1/B with currency EUR exist
    And   1 EUR is transferred from M1/A to M1/B

    Then metrics file /reports/metrics.M1.json reports:
    """
      promisedTransactions 1
      promisedTransfers 1
      committedTransactions 1
      committedTransfers 1
      rollbackedTransactions 0
      rollbackedTransfers 0
      forwardedTransactions 0
      forwardedTransfers 0
    """

  Scenario: metrics can remembers previous values after reboot
    Given vault is empty
    And   tenant M3 is onbdoarded
    And   ledger is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /reports/metrics.M3.json reports:
    """
      promisedTransactions 0
      promisedTransfers 0
      committedTransactions 0
      committedTransfers 0
      rollbackedTransactions 0
      rollbackedTransfers 0
      forwardedTransactions 0
      forwardedTransfers 0
    """

    When  pasive account M3/A with currency EUR exist
    And   pasive account M3/B with currency EUR exist
    And   1 EUR is transferred from M3/A to M3/B
    Then metrics file /reports/metrics.M3.json reports:
    """
      promisedTransactions 1
      promisedTransfers 1
      committedTransactions 1
      committedTransfers 1
      rollbackedTransactions 0
      rollbackedTransfers 0
      forwardedTransactions 0
      forwardedTransfers 0
    """

    When ledger is restarted
    Then metrics file /reports/metrics.M3.json reports:
    """
      promisedTransactions 1
      promisedTransfers 1
      committedTransactions 1
      committedTransfers 1
      rollbackedTransactions 0
      rollbackedTransfers 0
      forwardedTransactions 0
      forwardedTransfers 0
    """
