package net

import (
	"net"
	"testing"
)

func TestSkipInterface(t *testing.T) {
	ifaceDown := net.Interface{Flags: 0}
	if !SkipInterface(ifaceDown) {
		t.Fatal("expected down interface to be skipped")
	}

	ifaceLoopback := net.Interface{Flags: net.FlagUp | net.FlagLoopback}
	if !SkipInterface(ifaceLoopback) {
		t.Fatal("expected loopback interface to be skipped")
	}

	ifaceUp := net.Interface{Flags: net.FlagUp}
	if SkipInterface(ifaceUp) {
		t.Fatal("expected active non-loopback interface not to be skipped")
	}
}

func TestAddrToIP(t *testing.T) {
	ipNet := &net.IPNet{IP: net.ParseIP("10.0.0.1")}
	if got := AddrToIP(ipNet); got.String() != "10.0.0.1" {
		t.Fatalf("expected IPNet IP, got %v", got)
	}

	ipAddr := &net.IPAddr{IP: net.ParseIP("10.0.0.2")}
	if got := AddrToIP(ipAddr); got.String() != "10.0.0.2" {
		t.Fatalf("expected IPAddr IP, got %v", got)
	}
}

func TestCalculateIP(t *testing.T) {
	got, err := CalculateIP("10.0.0.1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "10.0.0.4" {
		t.Fatalf("expected incremented ip 10.0.0.4, got %s", got)
	}
}

func TestCalculateIPRejectsNonIPv4(t *testing.T) {
	_, err := CalculateIP("not-an-ip", 1)
	if err == nil {
		t.Fatal("expected error for non ipv4 address")
	}
}

func TestCalculateIPRejectsOverflow(t *testing.T) {
	_, err := CalculateIP("10.0.0.255", 1)
	if err == nil {
		t.Fatal("expected overflow error when incrementing past 255")
	}
}

func TestGetIPUsesProvidedStartingAddress(t *testing.T) {
	got, err := GetIP(2, "10.0.0.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "10.0.0.7" {
		t.Fatalf("expected 10.0.0.7, got %s", got)
	}
}
