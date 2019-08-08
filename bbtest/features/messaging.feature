Feature: Messaging behaviour

  Scenario: create transaction
    Given vault is empty
    And   tenant MSG1 is onboarded
    And   pasive account MSG1/A with currency EUR exist
    And   pasive account MSG1/B with currency EUR exist

    When  lake recieves "LedgerUnit/MSG1 LedgerRest req_id req_id NT trn_id x;MSG1;B;MSG1;A;1;EUR;2019-04-05T22:36:06Z"

    Then  lake responds with "VaultUnit/MSG1 LedgerUnit/MSG1 B req_id NP trn_id 1 EUR"
    And   lake responds with "VaultUnit/MSG1 LedgerUnit/MSG1 A req_id NP trn_id -1 EUR"
    And   lake responds with "VaultUnit/MSG1 LedgerUnit/MSG1 B req_id NC trn_id 1 EUR"
    And   lake responds with "VaultUnit/MSG1 LedgerUnit/MSG1 A req_id NC trn_id -1 EUR"
    And   lake responds with "LedgerRest LedgerUnit/MSG1 req_id req_id T0 trn_id"
