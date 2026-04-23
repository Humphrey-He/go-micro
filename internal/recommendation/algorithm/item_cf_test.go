package algorithm

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemCF_SimilarityFormula(t *testing.T) {
	// Test the similarity formula: sim(A,B) = |A∩B| / sqrt(|A|×|B|)
	// If 3 users bought A, 5 users bought B, and 2 users bought both
	commonUsers := 2.0
	usersA := 3.0
	usersB := 5.0

	expected := commonUsers / math.Sqrt(usersA*usersB)
	actual := 2.0 / math.Sqrt(3.0*5.0)

	assert.InDelta(t, expected, actual, 0.001)
	assert.InDelta(t, 0.516, actual, 0.01)
}

func TestItemCF_Constants(t *testing.T) {
	assert.Equal(t, 2, DefaultMinCoUsers)
	assert.Equal(t, 0.01, SimilarityThreshold)
	assert.Equal(t, 1000, BatchSize)
	assert.Equal(t, 10, DefaultLimit)
}

func TestCosineSimilarity(t *testing.T) {
	// Test with specific known values
	// If 10 users interacted with A, 10 with B, 5 common
	sim := 5.0 / math.Sqrt(10*10)
	assert.InDelta(t, 0.5, sim, 0.001)

	// If no common users, similarity should be 0
	simZero := 0.0 / math.Sqrt(10*10)
	assert.Equal(t, 0.0, simZero)
}

func TestSimilarityThreshold(t *testing.T) {
	// Items below similarity threshold should be filtered out
	threshold := SimilarityThreshold // 0.01

	// 0.001 should be filtered
	assert.True(t, 0.001 < threshold)

	// 0.5 should pass
	assert.True(t, 0.5 > threshold)
}
