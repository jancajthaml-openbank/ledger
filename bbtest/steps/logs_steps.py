from behave import *
from helpers.shell import execute
from helpers.eventually import eventually


@then('journalctl of "{unit}" contains following')
def step_impl(context, unit):
  expected_lines = [item.strip() for item in context.text.split('\n') if len(item.strip())]

  @eventually(5)
  def impl():
    service = unit.split('.service')[0].split('@')[0]
    (code, result, error) = execute([
      'journalctl', '-o', 'short-precise', '-t', service, '-u', unit, '--no-pager'
    ])

    assert code == 0

    actual_lines_merged = [item.strip() for item in result.split('\n') if len(item.strip())]
    actual_lines = []
    idx = len(actual_lines_merged) - 1

    while True:
      if idx < 0 or ("INFO >>> Start <<<" in actual_lines_merged[idx]):
        break
      actual_lines.append(actual_lines_merged[idx])
      idx -= 1

    actual_lines = reversed(actual_lines)

    for expected in expected_lines:
      found = False
      for actual in actual_lines:
        if expected in actual:
          found = True
          break

      assert found == True

  impl()
