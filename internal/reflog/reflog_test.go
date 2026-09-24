package reflog

import "testing"

const sample = "0000000000000000000000000000000000000000 1111111111111111111111111111111111111111 A U <a@u> 1700000000 +0000\tcommit (initial): first\n" +
	"1111111111111111111111111111111111111111 2222222222222222222222222222222222222222 A U <a@u> 1700000100 +0000\tcommit: second\n" +
	"2222222222222222222222222222222222222222 1111111111111111111111111111111111111111 A U <a@u> 1700000200 +0000\treset: moving to HEAD~1\n"

func TestParseLogNewestFirst(t *testing.T) {
	got := ParseLog(sample)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Message != "reset: moving to HEAD~1" {
		t.Fatalf("newest message = %q, want reset entry first", got[0].Message)
	}
	if got[0].Old != "2222222222222222222222222222222222222222" || got[0].New != "1111111111111111111111111111111111111111" {
		t.Fatalf("newest hashes = %q..%q", got[0].Old, got[0].New)
	}
}

func TestParseLogSkipsBlankAndMalformed(t *testing.T) {
	got := ParseLog("\n   \nnotabhere no message\n")
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}
