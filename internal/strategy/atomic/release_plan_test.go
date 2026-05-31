package atomic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPurgeCandidates_PurgesOldest(t *testing.T) {
	result := purgeCandidates([]string{"r1", "r2", "r3"}, 2)
	assert.Equal(t, []string{"r1"}, result)
}

func TestPurgeCandidates_NothingToPurge(t *testing.T) {
	result := purgeCandidates([]string{"r1", "r2"}, 2)
	assert.Nil(t, result)
}

func TestPurgeCandidates_PurgesMultiple(t *testing.T) {
	result := purgeCandidates([]string{"r1", "r2", "r3", "r4", "r5"}, 2)
	assert.Equal(t, []string{"r1", "r2", "r3"}, result)
}

func TestPurgeCandidates_Empty(t *testing.T) {
	result := purgeCandidates(nil, 2)
	assert.Nil(t, result)
}

// purgeFromList tests — verifies the active-release protection logic.

func TestPurgeFromList_ReservesOneSlotForActive(t *testing.T) {
	// keepN=3 but one slot is reserved for the active release,
	// so only keepN-1=2 old releases survive. Oldest is removed.
	result := purgeFromList([]string{"v1.0.0", "v1.0.1", "v1.0.3"}, 3)
	assert.Equal(t, []string{"v1.0.0"}, result)
}

func TestPurgeFromList_NothingToPurge(t *testing.T) {
	result := purgeFromList([]string{"v1.0.0", "v1.0.1"}, 3)
	assert.Nil(t, result)
}

func TestPurgeFromList_PurgesMultiple(t *testing.T) {
	result := purgeFromList([]string{"r1", "r2", "r3", "r4", "r5"}, 3)
	assert.Equal(t, []string{"r1", "r2", "r3"}, result)
}

func TestPurgeFromList_KeepOne_PurgesAllOld(t *testing.T) {
	// keepN=1 means only the active release is kept — all old ones go.
	result := purgeFromList([]string{"v1.0.0", "v1.0.1"}, 1)
	assert.Equal(t, []string{"v1.0.0", "v1.0.1"}, result)
}

func TestPurgeFromList_Empty(t *testing.T) {
	result := purgeFromList(nil, 3)
	assert.Nil(t, result)
}
