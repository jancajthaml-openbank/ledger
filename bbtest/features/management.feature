Feature: Properly behaving units

  Scenario: onboard
    Given tenant lorem is onboarded
    And   tenant ipsum is onboarded
    Then  systemctl contains following active units
      | name              | type    |
      | ledger            | path    |
      | ledger            | service |
      | ledger-rest       | service |
      | ledger-unit@lorem | service |
      | ledger-unit@ipsum | service |
    And unit "ledger-unit@lorem.service" is running
    And unit "ledger-unit@ipsum.service" is running

    When stop unit "ledger-unit@lorem.service"
    Then unit "ledger-unit@lorem.service" is not running
    And  unit "ledger-unit@ipsum.service" is running

    When start unit "ledger-unit@lorem.service"
    Then unit "ledger-unit@lorem.service" is running

    When restart unit "ledger-unit@lorem.service"
    Then unit "ledger-unit@lorem.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And   tenant ipsum is offboarded
    Then  systemctl does not contain following active units
      | name              | type    |
      | ledger-unit@lorem | service |
      | ledger-unit@ipsum | service |
    And systemctl contains following active units
      | name              | type    |
      | ledger            | path    |
      | ledger            | service |
      | ledger-rest       | service |
