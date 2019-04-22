Feature: High level Transaction workflow

  Scenario: integrity
    Given vault is empty
    And   tenant TRN is onbdoarded
    And   ledger is restarted

    Given pasive account TRN/A with currency EUR exist
    And   pasive account TRN/B with currency EUR exist
    And   TRN/A balance should be 0 EUR
    And   TRN/B balance should be 0 EUR

    When  0.00000000001 EUR is transferred from TRN/A to TRN/B
    Then  request should succeed
    And   transaction of tenant TRN should exist
    And   TRN/A balance should be -0.00000000001 EUR
    And   TRN/B balance should be 0.00000000001 EUR

    When  0.00000000001 EUR is transferred from TRN/C to TRN/D
    Then  request should fail
    And   transaction of tenant TRN should not exist

    When  0.00000000001 EUR is transferred from TRN/C to TRN/B
    Then  request should fail
    And   transaction of tenant TRN should exist
    And   TRN/B balance should be 0.00000000001 EUR

    When  0.00000000001 EUR is transferred from TRN/A to TRN/D
    Then  request should fail
    And   transaction of tenant TRN should exist
    And   TRN/A balance should be -0.00000000001 EUR

    Given active account TRN/E with currency EUR exist
    And   active account TRN/F with currency EUR exist

    When  0.00000000001 EUR is transferred from TRN/E to TRN/F
    Then  request should fail
    And   transaction of tenant TRN should exist
    And   TRN/E balance should be 0 EUR
    And   TRN/F balance should be 0 EUR

  Scenario: transaction between tenants
    Given vault is empty
    And   tenant T1 is onbdoarded
    And   tenant T2 is onbdoarded
    And   ledger is restarted

    Given pasive account T1/A with currency EUR exist
    And   pasive account T2/B with currency EUR exist
    And   T1/A balance should be 0 EUR
    And   T2/B balance should be 0 EUR

    When  0.00000000001 EUR is transferred from T1/A to T2/B
    Then  request should succeed
    And   transaction of tenant T1 should exist
    And   transaction of tenant T2 should not exist
    And   T1/A balance should be -0.00000000001 EUR
    And   T2/B balance should be 0.00000000001 EUR
