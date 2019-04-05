@install
Feature: Install package

  Scenario: install
    Given package "ledger.deb" is installed
    Then  systemctl contains following
    """
      ledger.service
      ledger.path
      ledger-rest.service
    """
