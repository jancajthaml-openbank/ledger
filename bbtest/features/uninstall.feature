Feature: Uninstall package

  Scenario: uninstall
    Given lake is not running
    And   package ledger is uninstalled
    Then  systemctl does not contain following active units
      | name           | type    |
      | ledger         | service |
      | ledger-rest    | service |
      | ledger-watcher | path    |
      | ledger-watcher | service |
