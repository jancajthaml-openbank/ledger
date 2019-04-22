Feature: REST

  Scenario: Transaction API
    Given vault is empty
    And   ledger is running
    And   tenant API is onbdoarded
    And   pasive account API/xxx with currency XXX exist
    And   pasive account API/yyy with currency XXX exist

    When I request curl POST https://127.0.0.1:4401/transaction/API
    """
      {
        "transfers": [{
          "credit": {
            "tenant": "API",
            "name": "Credit"
          },
          "debit": {
            "tenant": "API",
            "name": "Debit"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """
    Then curl responds with 417
    """
      {}
    """

    When I request curl GET https://127.0.0.1:4401/transaction/API/unique_transaction_id
    Then curl responds with 404
    """
      {}
    """

    When I request curl POST https://127.0.0.1:4401/transaction/API
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "API",
            "name": "xxx"
          },
          "debit": {
            "tenant": "API",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """
    Then curl responds with 200
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "API",
            "name": "xxx"
          },
          "debit": {
            "tenant": "API",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """

    When I request curl GET https://127.0.0.1:4401/transaction/API/unique_transaction_id
    Then curl responds with 200
    """
      {
        "id": "unique_transaction_id",
        "status": "committed",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "API",
            "name": "xxx"
          },
          "debit": {
            "tenant": "API",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """

    When I request curl GET https://127.0.0.1:4401/transaction/API
    Then curl responds with 200
    """
      [
        "unique_transaction_id"
      ]
    """

    When I request curl POST https://127.0.0.1:4401/transaction/API
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "API",
            "name": "xxx"
          },
          "debit": {
            "tenant": "API",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """
    Then curl responds with 200
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "API",
            "name": "xxx"
          },
          "debit": {
            "tenant": "API",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """

    When I request curl POST https://127.0.0.1:4401/transaction/API
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "API",
            "name": "xxx"
          },
          "debit": {
            "tenant": "API",
            "name": "yyy"
          },
          "amount": "2",
          "currency": "XXX"
        }]
      }
    """
    Then curl responds with 409
    """
      {}
    """
