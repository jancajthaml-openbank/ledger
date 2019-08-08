from behave import *
from helpers.eventually import eventually


@when('lake recieves "{data}"')
def lake_recieves(context, data):
  context.zmq.send(data)


@then('lake responds with "{data}"')
def lake_responds_with(context,  data):
  @eventually(2)
  def impl():
    assert data in context.zmq.backlog
    context.zmq.ack(data)
  impl()
