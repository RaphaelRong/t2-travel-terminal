package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStoreImplementations 确保所有 PostgreSQL store 都实现了对应接口。
func TestStoreImplementations(t *testing.T) {
	assert.Implements(t, (*SoulStore)(nil), NewPGSoulStore())
	assert.Implements(t, (*MemoryStore)(nil), NewPGMemoryStore())
	assert.Implements(t, (*SessionStore)(nil), NewPGSessionStore())
	assert.Implements(t, (*ProfileStore)(nil), NewPGProfileStore())
	assert.Implements(t, (*GodConfigStore)(nil), NewPGGodConfigStore())
	assert.Implements(t, (*UserDataStore)(nil), NewPGUserDataStore())
}
