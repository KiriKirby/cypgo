package main

import (
	"testing"
)

func TestParse(t *testing.T) {
	rs, l, e := parse("../../raw/animals-lepidoptera.txt")
	if e != nil {
		t.Fatal(e)
	}
	if l != 845 || len(rs) != 94 {
		t.Fatalf("%d/%d", l, len(rs))
	}
	if rs[0].Species != "Manduca sexta" || rs[0].ID != "CA798770" {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[1].Species != "Lepidoptera public collection" {
		t.Fatalf("unresolved=%#v", rs[1])
	}
	if rs[34].Species != "Papilio polyxenes" {
		t.Fatalf("papilio=%#v", rs[34])
	}
	if rs[90].Species != "Helicoverpa zea" || rs[93].Species != "Manduca sexta" {
		t.Fatalf("tail=%#v %#v", rs[90], rs[93])
	}
}
