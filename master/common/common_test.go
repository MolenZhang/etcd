package common

import "testing"

func TestUUID(t *testing.T) {
	ids := map[string]struct{}{}
	for i := 0; i < 10; i++ {
		id := GenerateID()
		t.Logf("--%v--%v", i, id)
		if _, ok := ids[id]; ok {
			t.Errorf("Unexpected id")
		}
		ids[id] = struct{}{}
	}
}
