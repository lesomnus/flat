//go:build !linux

package flob

func nlink(p string) (int, error) {
	// Return always 1 results delete of global blob regardless of the number of hard links to the file.
	// Even though the blobs for each repo are still available.
	return 1, nil
}
