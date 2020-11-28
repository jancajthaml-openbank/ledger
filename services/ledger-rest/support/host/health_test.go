package host

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func TestSystemNotifyValidation(t *testing.T) {
	t.Log("NOTIFY_SOCKET is not defined")
	{
		os.Unsetenv("NOTIFY_SOCKET")
		if systemNotify("foo") == nil {
			t.Fatalf("expected to return error")
		}
	}

	t.Log("NOTIFY_SOCKET is set to invalid value")
	{
		os.Setenv("NOTIFY_SOCKET", "/dev/null")
		defer os.Unsetenv("NOTIFY_SOCKET")
		if systemNotify("foo") == nil {
			t.Fatalf("expected to return error")
		}
	}
}

func TestNotifyServiceStatus(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "unixgram-*")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(f.Name())

	os.Setenv("NOTIFY_SOCKET", f.Name())
	defer os.Unsetenv("NOTIFY_SOCKET")

	ta, err := net.ResolveUnixAddr("unixgram", f.Name())
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.ListenUnixgram("unixgram", ta)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		l.Close()
		os.Remove(f.Name())
	}()

	t.Log("NotifyServiceReady")
	{
		NotifyServiceReady()
		b := make([]byte, 64)
		n, _, err := l.ReadFrom(b)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b[:n], []byte("READY=1")) {
			t.Fatalf("got %s; want READY=1", string(b[:n]))
		}
	}

	t.Log("NotifyServiceStopping")
	{
		NotifyServiceStopping()
		b := make([]byte, 64)
		n, _, err := l.ReadFrom(b)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b[:n], []byte("STOPPING=1")) {
			t.Fatalf("got %s; want STOPPING=1", string(b[:n]))
		}
	}
}
