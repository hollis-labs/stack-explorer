package git

import "testing"

func TestParseNameOnlyLog(t *testing.T) {
	raw := "aaa111\nfile1.go\nfile2.go\nfile2.go\n\nbbb222\nREADME.md\n\nccc333\n"
	got := ParseNameOnlyLog(raw)
	if len(got) != 2 {
		t.Fatalf("len(changeSets) = %d, want 2", len(got))
	}
	if got[0].Commit != "aaa111" || len(got[0].Files) != 2 {
		t.Fatalf("first changeset = %#v", got[0])
	}
	if got[1].Commit != "bbb222" || len(got[1].Files) != 1 || got[1].Files[0] != "README.md" {
		t.Fatalf("second changeset = %#v", got[1])
	}
}

func TestParseNameOnlyLogKeepsFilenamesWithoutSlashOrDot(t *testing.T) {
	raw := "aaa1111\nMakefile\nLICENSE\nDockerfile\n"
	got := ParseNameOnlyLog(raw)
	if len(got) != 1 {
		t.Fatalf("len(changeSets) = %d, want 1", len(got))
	}
	if len(got[0].Files) != 3 {
		t.Fatalf("len(files) = %d, want 3", len(got[0].Files))
	}
	for i, want := range []string{"Makefile", "LICENSE", "Dockerfile"} {
		if got[0].Files[i] != want {
			t.Fatalf("files[%d] = %q, want %q", i, got[0].Files[i], want)
		}
	}
}

func TestCoChangeWeights(t *testing.T) {
	changeSets := []ChangeSet{
		{Commit: "a", Files: []string{"a.go", "b.go", "c.go"}},
		{Commit: "b", Files: []string{"a.go", "b.go"}},
		{Commit: "c", Files: []string{"b.go", "c.go"}},
	}
	got := CoChangeWeights(changeSets)
	if len(got) != 3 {
		t.Fatalf("len(weights) = %d, want 3", len(got))
	}
	if got[0].Left != "a.go" || got[0].Right != "b.go" || got[0].Count != 2 || got[0].Weight != 1.0 {
		t.Fatalf("top weight = %#v", got[0])
	}
	if got[1].Count != 2 || got[1].Weight != 1.0 {
		t.Fatalf("second weight = %#v", got[1])
	}
	if got[2].Count != 1 || got[2].Weight != 0.5 {
		t.Fatalf("third weight = %#v", got[2])
	}
}
