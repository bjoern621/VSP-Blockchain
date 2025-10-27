package main

import "testing"

func TestAddition(t *testing.T) {
	sum := 2 + 3
	if sum != 5 {
		t.Errorf("Expected 5, got %d", sum)
	}
}

func TestSubtraction(t *testing.T) {
	diff := 10 - 4
	if diff != 6 {
		t.Errorf("Expected 6, got %d", diff)
	}
}

func TestMultiplication(t *testing.T) {
	prod := 3 * 7
	if prod != 21 {
		t.Errorf("Expected 21, got %d", prod)
	}
}

func TestDivision(t *testing.T) {
	quot := 8 / 2
	if quot != 4 {
		t.Errorf("Expected 4, got %d", quot)
	}
}
