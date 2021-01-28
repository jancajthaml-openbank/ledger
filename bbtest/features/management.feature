Feature: System control

  Scenario: check units presence
    Then  systemctl contains following active units
      | name              | type    |
      | ledger            | path    |
      | ledger            | service |
      | ledger-rest       | service |
      | ledger-unit@lorem | service |
      | ledger-unit@ipsum | service |

  Scenario: onboard
    Given tenant lorem is onboarded
    And   tenant ipsum is onboarded

    Then unit "ledger-unit@lorem.service" is running
    And unit "ledger-unit@ipsum.service" is running

  Scenario: stop
    When stop unit "ledger.service"
    Then unit "ledger-unit@lorem.service" is not running
    And  unit "ledger-unit@ipsum.service" is not running

  Scenario: start
    When start unit "ledger.service"
    Then unit "ledger-unit@lorem.service" is running
    And  unit "ledger-unit@ipsum.service" is running

  Scenario: restart
    When restart unit "ledger-unit@lorem.service"
    Then unit "ledger-unit@lorem.service" is running
    And  unit "ledger-unit@ipsum.service" is running

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
