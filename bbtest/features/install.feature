Feature: Install package

  Scenario: install
    Given package ledger is installed
    Then  systemctl contains following active units
      | name           | type    |
      | ledger         | service |
      | ledger-rest    | service |
      | ledger-watcher | path    |
      | ledger-watcher | service |
