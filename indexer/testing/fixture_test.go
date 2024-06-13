package indexertesting

import "testing"

func TestListenerFixture(t *testing.T) {
	fixture := NewListenerTestFixture(StdoutListener(), ListenerTestFixtureOptions{})

	err := fixture.Initialize()
	if err != nil {
		t.Fatal(err)
	}

	err = fixture.NextBlock()
	if err != nil {
		t.Fatal(err)
	}
}
