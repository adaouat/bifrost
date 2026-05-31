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
