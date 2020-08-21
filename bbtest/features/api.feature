Feature: REST

  Scenario: Tenant API
    Given unit "ledger-rest.service" is running

    When I request HTTP https://127.0.0.1/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      []
      """

    When I request HTTP https://127.0.0.1/tenant/APITESTA
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant/APITESTB
      | key    | value |
      | method |  POST |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      [
        "APITESTB"
      ]
      """

    When I request HTTP https://127.0.0.1/tenant/APITESTC
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant/APITESTC
      | key    | value  |
      | method | DELETE |
    Then HTTP response is
      | key    | value  |
      | status | 200    |


  Scenario: Transaction API
    Given unit "ledger-rest.service" is running
    And   vault is empty
    And   tenant API is onboarded
    And   pasive account API/xxx with currency XXX exist
    And   pasive account API/yyy with currency XXX exist

    When I request HTTP https://127.0.0.1/transaction/API
      | key    | value |
      | method | POST  |
      """
      {
        "transfers": [
          {
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
          }
        ]
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 417   |

    When I request HTTP https://127.0.0.1/transaction/API/unique_transaction_id
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 404   |

    When I request HTTP https://127.0.0.1/transaction/API
      | key    | value |
      | method | POST  |
      """
      {
        "id": "unique_transaction_id",
        "transfers": [
          {
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
          }
        ]
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/transaction/API/unique_transaction_id
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      {
        "id": "unique_transaction_id",
        "status": "committed",
        "transfers": [
          {
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
          }
        ]
      }
      """

    When I request HTTP https://127.0.0.1/transaction/API
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      [
        "unique_transaction_id"
      ]
      """

    When I request HTTP https://127.0.0.1/transaction/API
      | key    | value |
      | method | POST  |
      """
      {
        "id": "unique_transaction_id",
        "transfers": [
          {
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
          }
        ]
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/transaction/API
      | key    | value |
      | method | POST  |
      """
      {
        "id": "unique_transaction_id",
        "transfers": [
          {
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
          }
        ]
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 409   |


  Scenario: Health API
    Given unit "ledger-rest.service" is running

    When I request HTTP https://127.0.0.1/health
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
