Feature: Service can be configured

  Scenario: configure log level
    Given tenant CONFIGURATION is onbdoarded
    And ledger is reconfigured with
    """
      LOG_LEVEL=DEBUG
    """
    Then journalctl of "ledger-unit@CONFIGURATION.service" contains following
    """
      Log level set to DEBUG
    """

    Given ledger is reconfigured with
    """
      LOG_LEVEL=ERROR
    """
    Then journalctl of "ledger-unit@CONFIGURATION.service" contains following
    """
      Log level set to ERROR
    """

    Given ledger is reconfigured with
    """
      LOG_LEVEL=INFO
    """
    Then journalctl of "ledger-unit@CONFIGURATION.service" contains following
    """
      Log level set to INFO
    """
