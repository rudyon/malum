package identifier

import "testing"

func TestNewUUIDProducesVersionFourUUID(t *testing.T) {
	id, err := NewUUID()
	if err != nil {
		t.Fatalf("NewUUID() error = %v", err)
	}
	if !IsUUID(id) {
		t.Fatalf("NewUUID() = %q; want valid UUID", id)
	}
	if id[14] != '4' || (id[19] != '8' && id[19] != '9' && id[19] != 'a' && id[19] != 'b') {
		t.Fatalf("NewUUID() = %q; want RFC 4122 version 4 and variant bits", id)
	}
}
