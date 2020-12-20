Feature: Metrics test

  Scenario: metrics measures expected stats
    Given vault is empty
    And   tenant M2 is onboarded

    Then metrics reports:
      | key                                    | type  |      tags | value |
      | openbank.ledger.transaction.promised   | count | tenant:M2 |     0 |
      | openbank.ledger.transfer.promised      | count | tenant:M2 |     0 |
      | openbank.ledger.transaction.committed  | count | tenant:M2 |     0 |
      | openbank.ledger.transfer.committed     | count | tenant:M2 |     0 |
      | openbank.ledger.transaction.rollbacked | count | tenant:M2 |     0 |
      | openbank.ledger.transfer.rollbacked    | count | tenant:M2 |     0 |

    When  pasive account M2/A with currency EUR exist
    And   pasive account M2/B with currency EUR exist
    And   1 EUR is transferred from M2/A to M2/B

    Then metrics reports:
      | key                                    | type  |      tags | value |
      | openbank.ledger.transaction.promised   | count | tenant:M2 |     1 |
      | openbank.ledger.transfer.promised      | count | tenant:M2 |     1 |
      | openbank.ledger.transaction.committed  | count | tenant:M2 |     1 |
      | openbank.ledger.transfer.committed     | count | tenant:M2 |     1 |
      | openbank.ledger.transaction.rollbacked | count | tenant:M2 |     0 |
      | openbank.ledger.transfer.rollbacked    | count | tenant:M2 |     0 |
