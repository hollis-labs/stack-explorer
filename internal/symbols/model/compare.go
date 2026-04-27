package model

func CopyMissingSymbolFields(dst, src *Symbol) {
	if dst == nil || src == nil {
		return
	}
	if dst.Docstring == "" {
		dst.Docstring = src.Docstring
	}
	if dst.ParentSymbolID == nil && src.ParentSymbolID != nil {
		id := *src.ParentSymbolID
		dst.ParentSymbolID = &id
	}
}

func SameStoredSymbol(left, right *Symbol) bool {
	if left == nil || right == nil {
		return false
	}
	return left.RepoID == right.RepoID &&
		left.Kind == right.Kind &&
		left.Name == right.Name &&
		left.QualifiedName == right.QualifiedName &&
		left.FilePath == right.FilePath &&
		sameIntPtr(left.LineStart, right.LineStart) &&
		sameIntPtr(left.LineEnd, right.LineEnd) &&
		left.ContentHash == right.ContentHash &&
		left.SignatureHash == right.SignatureHash &&
		sameInt64Ptr(left.ParentSymbolID, right.ParentSymbolID) &&
		left.Language == right.Language &&
		left.Visibility == right.Visibility &&
		left.Docstring == right.Docstring &&
		sameStringPtr(left.StaleSinceCommit, right.StaleSinceCommit)
}

func sameIntPtr(left, right *int) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func sameInt64Ptr(left, right *int64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func sameStringPtr(left, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
