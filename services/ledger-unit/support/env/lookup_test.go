package env

import (
	"os"
	"time"
	"testing"
)

func TestEnvString(t *testing.T) {

	t.Log("TEST_STR missing")
	{
		if String("TEST_STR", "x") != "x" {
			t.Errorf("String did not provide default value")
		}
	}

	t.Log("TEST_STR present")
	{
		os.Setenv("TEST_STR", "y")
		defer os.Unsetenv("TEST_STR")

		if String("TEST_STR", "x") != "y" {
			t.Errorf("String did not obtain env value")
		}
	}
}

func TestEnvInteger(t *testing.T) {

	t.Log("TEST_INT missing")
	{
		if Int("TEST_INT", 0) != 0 {
			t.Errorf("Int did not provide default value")
		}
	}

	t.Log("TEST_INT present and valid")
	{
		os.Setenv("TEST_INT", "1")
		defer os.Unsetenv("TEST_INT")

		if Int("TEST_INT", 0) != 1 {
			t.Errorf("Int did not obtain env value")
		}
	}

	t.Log("TEST_INT present and invalid")
	{
		os.Setenv("TEST_INT", "x")
		defer os.Unsetenv("TEST_INT")

		if Int("TEST_INT", 0) != 0 {
			t.Errorf("Int did not fallback to default value")
		}
	}
}


func TestEnvDuration(t *testing.T) {

	t.Log("TEST_DUR missing")
	{
		if Duration("TEST_DUR", time.Second) != time.Second {
			t.Errorf("Duration did not provide default value")
		}
	}

	t.Log("TEST_DUR present and valid")
	{
		os.Setenv("TEST_DUR", "2s")
		defer os.Unsetenv("TEST_DUR")

		if Duration("TEST_DUR", time.Second) != 2*time.Second {
			t.Errorf("Duration did not obtain env value")
		}
	}

	t.Log("TEST_DUR present and invalid")
	{
		os.Setenv("TEST_DUR", "x")
		defer os.Unsetenv("TEST_DUR")

		if Duration("TEST_DUR", time.Second) != time.Second {
			t.Errorf("Duration did not obtain fallback to default value")
		}
	}
}

