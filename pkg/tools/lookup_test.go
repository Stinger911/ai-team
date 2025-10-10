package tools

import "testing"

func TestLookupArgFlexible_Variants(t *testing.T) {
    args := map[string]interface{}{
        "file_path": "a.txt",
        "Content": "hello",
        "filePath": "b.txt",
    }

    if v, ok := lookupArgFlexible(args, "file_path"); !ok || v != "a.txt" {
        t.Fatalf("expected to find file_path -> a.txt, got %v, %v", v, ok)
    }

    if v, ok := lookupArgFlexible(args, "filePath"); !ok || (v != "a.txt" && v != "b.txt") {
        t.Fatalf("expected to find filePath variant, got %v, %v", v, ok)
    }

    if v, ok := lookupArgFlexible(args, "content"); !ok || v != "hello" {
        t.Fatalf("expected to find content case-insensitive -> hello, got %v, %v", v, ok)
    }
}
