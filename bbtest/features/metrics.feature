Feature: Metrics test

  Scenario: metrics report expected results
    Given vault is empty
    And   tenant M1 is onbdoarded
    And   ledger is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
      METRICS_REFRESHRATE=1s
    """

    When  pasive account M1/A with currency EUR exist
    And   pasive account M1/B with currency EUR exist
    And   1 EUR is transferred from M1/A to M1/B

    Then  metrics for tenant M1 should report 1 promised transactions
    And   metrics for tenant M1 should report 1 committed transactions

  Scenario: metrics have expected keys
    Given vault is empty
    And   tenant M2 is onbdoarded
    And   ledger is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /reports/metrics.M2.json should have following keys:
    """
      promisedTransactions
      committedTransactions
      rollbackedTransactions
      forwardedTransactions
    """
    And metrics file /reports/metrics.json should have following keys:
    """
      createTransactionLatency
      forwardTransferLatency
      getTransactionLatency
      getTransactionsLatency
    """
