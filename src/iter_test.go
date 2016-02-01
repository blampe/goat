package goaat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpDown(t *testing.T) {
	c := upDown(2, 2)

	result := make([]Index, 0, 4)

	for tup := range c {
		result = append(result, tup)
	}

	expected := []Index{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}

	assert.Equal(t, expected, result)
}

func TestLeftRight(t *testing.T) {
	c := leftRight(2, 2)

	result := make([]Index, 0, 4)

	for tup := range c {
		result = append(result, tup)
	}

	expected := []Index{
		{0, 0},
		{1, 0},
		{0, 1},
		{1, 1},
	}

	assert.Equal(t, expected, result)
}

func TestDiagUp(t *testing.T) {
	c := diagUp(2, 3)

	result := make([]Index, 0, 6)

	for tup := range c {
		result = append(result, tup)
	}

	// 1 2
	// 2 3
	// 3 4

	expected := []Index{
		{0, 0},
		{0, 1},
		{1, 0},
		{0, 2},
		{1, 1},
		{1, 2},
	}

	assert.Equal(t, expected, result)
}

//func TestDiagDown(t *testing.T) {
//c := diagDown(2, 3)

//result := make([]Index, 0, 6)

//for tup := range c {
//result = append(result, tup)
//}

//// 2 1
//// 3 2
//// 4 3

//expected := []Index{
//{1, 0},
//{0, 0},
//{1, 1},
//{0, 1},
//{1, 2},
//{0, 2},
//}

//assert.Equal(t, expected, result)
//}

func TestDiagDown(t *testing.T) {
	c := diagDown(3, 2)

	result := make([]Index, 0, 6)

	for tup := range c {
		result = append(result, tup)
	}

	// 3 2 1
	// 4 3 2

	expected := []Index{
		{0, 1}, // x - y == -1
		{0, 0}, // x - y == 0
		{1, 1}, // x - y == 0
		{1, 0}, // x - y == 1
		{2, 1}, // x - y == 1
		{2, 0}, // x - y == 2
	}

	assert.Equal(t, expected, result)
}
