Feature: Metrics test

  Scenario: metrics measures expected stats
    Given vault is empty
    And   tenant M2 is onboarded

    Then metrics reports:
      | key                                       | type  | value |
      | openbank.ledger.M2.transaction.promised   | count |     0 |
      | openbank.ledger.M2.transfer.promised      | count |     0 |
      | openbank.ledger.M2.transaction.committed  | count |     0 |
      | openbank.ledger.M2.transfer.committed     | count |     0 |
      | openbank.ledger.M2.transaction.rollbacked | count |     0 |
      | openbank.ledger.M2.transfer.rollbacked    | count |     0 |

    When  pasive account M2/A with currency EUR exist
    And   pasive account M2/B with currency EUR exist
    And   1 EUR is transferred from M2/A to M2/B

    Then metrics reports:
      | key                                       | type  | value |
      | openbank.ledger.M2.transaction.promised   | count |     1 |
      | openbank.ledger.M2.transfer.promised      | count |     1 |
      | openbank.ledger.M2.transaction.committed  | count |     1 |
      | openbank.ledger.M2.transfer.committed     | count |     1 |
      | openbank.ledger.M2.transaction.rollbacked | count |     0 |
      | openbank.ledger.M2.transfer.rollbacked    | count |     0 |
