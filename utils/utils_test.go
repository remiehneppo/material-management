package utils

import "testing"

func TestIndexPathRoundTrip(t *testing.T) {
	tests := []string{
		"1",
		"1.2",
		"1.2.3",
		"5.10.15",
		"63.63.63",
		"1.1.1.1.1",
		"10.20.30.40.50",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// Encode
			encoded, err := StringToIndexPath(input)
			if err != nil {
				t.Fatalf("StringToIndexPath(%q) error: %v", input, err)
			}

			// Decode
			decoded := IndexPathToString(encoded)

			// Verify
			if decoded != input {
				t.Errorf("Round trip failed:\n  Input:   %q\n  Encoded: %d (0x%X)\n  Decoded: %q",
					input, encoded, encoded, decoded)
			}
		})
	}
}

func TestIndexPathEdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		path, err := StringToIndexPath("")
		if err == nil {
			t.Error("Expected error for empty string")
		}
		if path != 0 {
			t.Errorf("Expected 0, got %d", path)
		}

		str := IndexPathToString(0)
		if str != "" {
			t.Errorf("Expected empty string, got %q", str)
		}
	})

	t.Run("value out of range", func(t *testing.T) {
		_, err := StringToIndexPath("64.1")
		if err == nil {
			t.Error("Expected error for value > 63")
		}
	})

	t.Run("max depth", func(t *testing.T) {
		// 10 levels max (60 bits / 6 bits per level)
		input := "1.2.3.4.5.6.7.8.9.10"
		encoded, err := StringToIndexPath(input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		decoded := IndexPathToString(encoded)
		if decoded != input {
			t.Errorf("Max depth failed: got %q, want %q", decoded, input)
		}
	})
}

func TestIndexPathManualVerify(t *testing.T) {
	// Test "1.2" manually
	// Expected: (1 << 54) | (2 << 48) - using bits 54-59 and 48-53
	expected := (int64(1) << 54) | (int64(2) << 48)

	result, err := StringToIndexPath("1.2")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if result != expected {
		t.Errorf("StringToIndexPath(\"1.2\") = %d (0x%X), want %d (0x%X)",
			result, result, expected, expected)
	}

	decoded := IndexPathToString(result)
	if decoded != "1.2" {
		t.Errorf("IndexPathToString(%d) = %q, want \"1.2\"", result, decoded)
	}
}

func TestIndexPathSortOrder(t *testing.T) {
	t.Run("all children of 1.x must be less than 2.x", func(t *testing.T) {
		children := []string{
			"1.1",
			"1.2",
			"1.2.4",
			"1.5.10",
			"1.63.63.63", // Maximum possible children of 1
		}

		parents := []string{
			"2",
			"2.1",
			"2.5.10",
		}

		// Encode all children
		childValues := make([]int64, len(children))
		for i, child := range children {
			val, err := StringToIndexPath(child)
			if err != nil {
				t.Fatalf("Failed to encode %q: %v", child, err)
			}
			childValues[i] = val
		}

		// Encode all parents
		parentValues := make([]int64, len(parents))
		for i, parent := range parents {
			val, err := StringToIndexPath(parent)
			if err != nil {
				t.Fatalf("Failed to encode %q: %v", parent, err)
			}
			parentValues[i] = val
		}

		// Verify: every child < every parent
		for i, childVal := range childValues {
			for j, parentVal := range parentValues {
				if childVal >= parentVal {
					t.Errorf("Sort order violation: %q (%d) >= %q (%d)",
						children[i], childVal, parents[j], parentVal)
				}
			}
		}
	})

	t.Run("hierarchical sort order verification", func(t *testing.T) {
		// List in expected sorted order
		sortedPaths := []string{
			"1",
			"1.1",
			"1.2",
			"1.2.1",
			"1.2.2",
			"1.2.10",
			"1.3",
			"1.10",
			"1.63",
			"2",
			"2.1",
			"2.2",
			"3",
			"3.1.1",
			"10",
			"63",
		}

		// Encode all paths
		encodedValues := make([]int64, len(sortedPaths))
		for i, path := range sortedPaths {
			val, err := StringToIndexPath(path)
			if err != nil {
				t.Fatalf("Failed to encode %q: %v", path, err)
			}
			encodedValues[i] = val
		}

		// Verify that encoded values are in ascending order
		for i := 0; i < len(encodedValues)-1; i++ {
			if encodedValues[i] >= encodedValues[i+1] {
				t.Errorf("Sort order violation at position %d:\n  %q (%d / 0x%X)\n  is not less than\n  %q (%d / 0x%X)",
					i,
					sortedPaths[i], encodedValues[i], encodedValues[i],
					sortedPaths[i+1], encodedValues[i+1], encodedValues[i+1])
			}
		}
	})

	t.Run("sibling comparison", func(t *testing.T) {
		tests := []struct {
			smaller string
			larger  string
		}{
			{"1.1", "1.2"},
			{"1.1", "1.10"},
			{"1.9", "1.10"},
			{"1.1.1", "1.1.2"},
			{"1.1.5", "1.2.1"},
			{"1.63", "2.1"},
		}

		for _, tt := range tests {
			smallerVal, err := StringToIndexPath(tt.smaller)
			if err != nil {
				t.Fatalf("Failed to encode %q: %v", tt.smaller, err)
			}

			largerVal, err := StringToIndexPath(tt.larger)
			if err != nil {
				t.Fatalf("Failed to encode %q: %v", tt.larger, err)
			}

			if smallerVal >= largerVal {
				t.Errorf("Expected %q (%d) < %q (%d), but got %q >= %q",
					tt.smaller, smallerVal, tt.larger, largerVal,
					tt.smaller, tt.larger)
			}
		}
	})

	t.Run("parent always less than children", func(t *testing.T) {
		tests := []struct {
			parent   string
			children []string
		}{
			{
				parent:   "1",
				children: []string{"1.1", "1.2", "1.10", "1.63"},
			},
			{
				parent:   "1.2",
				children: []string{"1.2.1", "1.2.2", "1.2.10", "1.2.63"},
			},
			{
				parent:   "5",
				children: []string{"5.1", "5.20.30"},
			},
		}

		for _, tt := range tests {
			parentVal, err := StringToIndexPath(tt.parent)
			if err != nil {
				t.Fatalf("Failed to encode parent %q: %v", tt.parent, err)
			}

			for _, child := range tt.children {
				childVal, err := StringToIndexPath(child)
				if err != nil {
					t.Fatalf("Failed to encode child %q: %v", child, err)
				}

				if parentVal >= childVal {
					t.Errorf("Parent %q (%d) should be < child %q (%d)",
						tt.parent, parentVal, child, childVal)
				}
			}
		}
	})
}
