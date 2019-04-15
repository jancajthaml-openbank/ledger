@uninstall
Feature: Uninstall package

  Scenario: uninstall
    Given package "ledger" is uninstalled
    Then  systemctl does not contains following
    """
      ledger.service
      ledger.path
      ledger-rest.service
    """
