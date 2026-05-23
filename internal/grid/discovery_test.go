package grid

import (
	"testing"

	"github.com/hypebeast/go-osc/osc"
)

func TestBuildSerialoscList(t *testing.T) {
	data, err := buildSerialoscList("127.0.0.1", 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty message")
	}

	// Round-trip: parse the message back.
	packet, err := osc.ParsePacket(string(data))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	msg, ok := packet.(*osc.Message)
	if !ok {
		t.Fatal("expected OSC message")
	}
	if msg.Address != "/serialosc/list" {
		t.Errorf("address = %q, want /serialosc/list", msg.Address)
	}
	if len(msg.Arguments) < 2 {
		t.Fatalf("expected at least 2 args, got %d", len(msg.Arguments))
	}
	host, _ := msg.Arguments[0].(string)
	port, _ := msg.Arguments[1].(int32)
	if host != "127.0.0.1" {
		t.Errorf("host = %q, want 127.0.0.1", host)
	}
	if port != 12345 {
		t.Errorf("port = %d, want 12345", port)
	}
}

func TestParseSerialoscDeviceValid(t *testing.T) {
	// Build a valid /serialosc/device message.
	msg := osc.NewMessage("/serialosc/device", "m0001234", "monome 128", int32(8080))
	data, err := msg.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	id, devType, port := parseSerialoscDevice(data)
	if id != "m0001234" {
		t.Errorf("id = %q, want m0001234", id)
	}
	if devType != "monome 128" {
		t.Errorf("devType = %q, want monome 128", devType)
	}
	if port != 8080 {
		t.Errorf("port = %d, want 8080", port)
	}
}

func TestParseSerialoscDeviceInvalidAddress(t *testing.T) {
	msg := osc.NewMessage("/wrong/address", "id", "type", int32(1))
	data, _ := msg.MarshalBinary()

	id, _, port := parseSerialoscDevice(data)
	if id != "" || port != 0 {
		t.Errorf("wrong address: expected empty result, got id=%q port=%d", id, port)
	}
}

func TestParseSerialoscDeviceTooFewArgs(t *testing.T) {
	msg := osc.NewMessage("/serialosc/device", "only_one_arg")
	data, _ := msg.MarshalBinary()

	id, _, port := parseSerialoscDevice(data)
	if id != "" || port != 0 {
		t.Errorf("too few args: expected empty result, got id=%q port=%d", id, port)
	}
}

func TestParseSerialoscDeviceGarbageData(t *testing.T) {
	id, _, port := parseSerialoscDevice([]byte{0x00, 0x01, 0x02, 0xFF})
	if id != "" || port != 0 {
		t.Errorf("garbage data: expected empty result, got id=%q port=%d", id, port)
	}
}

func TestBuildSerialoscListRoundTrip(t *testing.T) {
	// Full end-to-end: build, parse, extract fields.
	data, err := buildSerialoscList("10.0.0.1", 9999)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}

	packet, err := osc.ParsePacket(string(data))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	msg := packet.(*osc.Message)
	host, _ := msg.Arguments[0].(string)
	port, _ := msg.Arguments[1].(int32)

	if host != "10.0.0.1" {
		t.Errorf("host = %q", host)
	}
	if port != 9999 {
		t.Errorf("port = %d", port)
	}
}
