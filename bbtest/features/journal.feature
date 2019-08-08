Feature: Persistent journal

  Scenario: create account
    Given tenant JOURNAL is onboarded

    When  pasive account JOURNAL/A with currency EUR exist
    And   pasive account JOURNAL/B with currency EUR exist
    When  0.00000000001 EUR is transferred with id X from JOURNAL/A to JOURNAL/B

    Then transaction X of tenant JOURNAL should be
      | key            | value     |
      | state          | committed |
