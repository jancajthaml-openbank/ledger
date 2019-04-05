Feature: Properly behaving units

  Scenario: onboard
    Given tenant lorem is onbdoarded
    And   tenant ipsum is onbdoarded
    Then  systemctl contains following
    """
      ledger.service
      ledger-rest.service
      ledger-unit@lorem.service
      ledger-unit@ipsum.service
    """

    When stop unit "ledger-unit@lorem.service"
    Then unit "ledger-unit@lorem.service" is not running

    When start unit "ledger-unit@lorem.service"
    Then unit "ledger-unit@lorem.service" is running

    When restart unit "ledger-unit@ipsum.service"
    Then unit "ledger-unit@ipsum.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And   tenant ipsum is offboarded

    Then  systemctl does not contains following
    """
      ledger-unit@lorem.service
      ledger-unit@ipsum.service
    """
