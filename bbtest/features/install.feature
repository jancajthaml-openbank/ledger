Feature: Install package

  Scenario: install
    Given package ledger is installed
    Then  systemctl contains following active units
      | name        | type    |
      | ledger-rest | service |
      | ledger      | service |
      | ledger      | path    |
