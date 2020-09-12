Feature: Service can be configured

  Scenario: configure log level to DEBUG
    Given tenant CONFIGURATION_DEBUG is onboarded
    And   ledger is configured with
      | property  | value |
      | LOG_LEVEL | DEBUG |

    Then journalctl of "ledger-unit@CONFIGURATION_DEBUG.service" contains following
    """
      Log level set to DEBUG
    """

  Scenario: configure log level to ERROR
    Given tenant CONFIGURATION_ERROR is onboarded
    And   ledger is configured with
      | property  | value |
      | LOG_LEVEL | ERROR |

    Then journalctl of "ledger-unit@CONFIGURATION_ERROR.service" contains following
    """
      Log level set to ERROR
    """

  Scenario: configure log level to INFO
    Given tenant CONFIGURATION_INFO is onboarded
    And   ledger is configured with
      | property  | value |
      | LOG_LEVEL | INFO  |

    Then journalctl of "ledger-unit@CONFIGURATION_INFO.service" contains following
    """
      Log level set to INFO
    """

 Scenario: configure log level to INVALID
    Given lake is configured with
      | property  | value   |
      | LOG_LEVEL | INVALID |

    Then journalctl of "ledger-unit@CONFIGURATION_INFO.service" contains following
    """
      Log level set to INFO
    """
