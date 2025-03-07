package orderedmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type marshallable int

func (m marshallable) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("#%d#", m)), nil
}

func (m *marshallable) UnmarshalText(text []byte) error {
	if len(text) < 3 {
		return errors.New("too short")
	}
	if text[0] != '#' || text[len(text)-1] != '#' {
		return errors.New("missing prefix or suffix")
	}

	value, err := strconv.Atoi(string(text[1 : len(text)-1]))
	if err != nil {
		return err
	}

	*m = marshallable(value)
	return nil
}

func TestMarshalJSON(t *testing.T) {
	t.Run("int key", func(t *testing.T) {
		om := New[int, any]()
		om.Set(1, "bar")
		om.Set(7, "baz")
		om.Set(2, 28)
		om.Set(3, 100)
		om.Set(4, "baz")
		om.Set(5, "28")
		om.Set(6, "100")
		om.Set(8, "baz")
		om.Set(8, "baz")

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"1":"bar","7":"baz","2":28,"3":100,"4":"baz","5":"28","6":"100","8":"baz"}`, string(b))
	})

	t.Run("string key", func(t *testing.T) {
		om := New[string, any]()
		om.Set("test", "bar")
		om.Set("abc", true)

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"test":"bar","abc":true}`, string(b))
	})

	t.Run("TextMarshaller key", func(t *testing.T) {
		om := New[marshallable, any]()
		om.Set(marshallable(1), "bar")
		om.Set(marshallable(28), true)

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{"#1#":"bar","#28#":true}`, string(b))
	})

	t.Run("empty map", func(t *testing.T) {
		om := New[string, any]()

		b, err := json.Marshal(om)
		assert.NoError(t, err)
		assert.Equal(t, `{}`, string(b))
	})
}

func TestUnmarshallJSON(t *testing.T) {
	t.Run("int key", func(t *testing.T) {
		data := `{"1":"bar","7":"baz","2":28,"3":100,"4":"baz","5":"28","6":"100","8":"baz"}`

		om := New[int, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]int{1, 7, 2, 3, 4, 5, 6, 8},
			[]any{"bar", "baz", float64(28), float64(100), "baz", "28", "100", "baz"})
	})

	t.Run("string key", func(t *testing.T) {
		data := `{"test":"bar","abc":true}`

		om := New[string, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]string{"test", "abc"},
			[]any{"bar", true})
	})

	t.Run("TextUnmarshaler key", func(t *testing.T) {
		data := `{"#1#":"bar","#28#":true}`

		om := New[marshallable, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertOrderedPairsEqual(t, om,
			[]marshallable{1, 28},
			[]any{"bar", true})
	})

	t.Run("when fed with an input that's not an object", func(t *testing.T) {
		for _, data := range []string{"true", `["foo"]`, "42", `"foo"`} {
			om := New[int, any]()
			require.Error(t, json.Unmarshal([]byte(data), &om))
		}
	})

	t.Run("empty map", func(t *testing.T) {
		data := `{}`

		om := New[int, any]()
		require.NoError(t, json.Unmarshal([]byte(data), &om))

		assertLenEqual(t, om, 0)
	})
}

func BenchmarkMarshalJSON(b *testing.B) {
	om := New[int, any]()
	om.Set(1, "bar")
	om.Set(7, "baz")
	om.Set(2, 28)
	om.Set(3, 100)
	om.Set(4, "baz")
	om.Set(5, "28")
	om.Set(6, "100")
	om.Set(8, "baz")
	om.Set(8, "baz")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(om)
	}
}
