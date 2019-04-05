Feature: REST

  Scenario: Transaction API
    Given vault is empty
    And   tenant APITRN is onbdoarded
    And   ledger is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """
    And   pasive account APITRN/xxx with currency XXX exist
    And   pasive account APITRN/yyy with currency XXX exist

    When I request curl POST https://localhost/transaction/APITRN
    """
      {
        "transfers": [{
          "credit": {
            "tenant": "APITRN",
            "name": "Credit"
          },
          "debit": {
            "tenant": "APITRN",
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

    When I request curl GET https://localhost/transaction/APITRN/unique_transaction_id
    Then curl responds with 404
    """
      {}
    """

    When I request curl POST https://localhost/transaction/APITRN
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "APITRN",
            "name": "xxx"
          },
          "debit": {
            "tenant": "APITRN",
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
            "tenant": "APITRN",
            "name": "xxx"
          },
          "debit": {
            "tenant": "APITRN",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """

    When I request curl GET https://localhost/transaction/APITRN/unique_transaction_id
    Then curl responds with 200
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "APITRN",
            "name": "xxx"
          },
          "debit": {
            "tenant": "APITRN",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """

    When I request curl POST https://localhost/transaction/APITRN
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "APITRN",
            "name": "xxx"
          },
          "debit": {
            "tenant": "APITRN",
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
            "tenant": "APITRN",
            "name": "xxx"
          },
          "debit": {
            "tenant": "APITRN",
            "name": "yyy"
          },
          "amount": "1",
          "currency": "XXX"
        }]
      }
    """

    When I request curl POST https://localhost/transaction/APITRN
    """
      {
        "id": "unique_transaction_id",
        "transfers": [{
          "id": "unique_transfer_id",
          "valueDate": "2018-03-04T17:08:22Z",
          "credit": {
            "tenant": "APITRN",
            "name": "xxx"
          },
          "debit": {
            "tenant": "APITRN",
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
